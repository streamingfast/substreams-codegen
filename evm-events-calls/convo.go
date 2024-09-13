package evm_events_calls

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"golang.org/x/exp/maps"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))
var AbiFilepathPrefix = "file://"

func init() {
	codegen.RegisterConversation(
		"evm-events-calls",
		"Decode Ethereum events/calls and create a substreams as source",
		"Given a list of contracts and their ABIs, this will build an Ethereum substreams that decodes events and/or calls",
		codegen.ConversationFactory(New),
		82,
	)
}

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{currentContractIdx: -1},
	}}
}

func (c *Convo) contextContract() *Contract {
	p := c.State
	if p.currentContractIdx == -1 || p.currentContractIdx > len(p.Contracts)-1 {
		return nil
	}
	return p.Contracts[p.currentContractIdx]
}

func (c *Convo) NextStep() (out loop.Cmd) {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if !isValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if len(p.Contracts) == 0 {
		return cmd(StartFirstContract{})
	}

	previousContractIdx := p.currentContractIdx
	for idx, contract := range p.Contracts {
		p.currentContractIdx = idx

		notifyContext := func(next loop.Cmd) loop.Cmd {
			if previousContractIdx != p.currentContractIdx {
				return loop.Seq(cmd(MsgContractSwitch{}), next)
			}
			return next
		}

		if contract.Address == "" {
			return cmd(AskContractAddress{})
		}

		if contract.Abi == nil || contract.Abi.abi == nil {
			// if the user pasted an empty ABI, we would restart the process or choosing a contract address
			if contract.emptyABI {
				contract.Address = ""     // reset the address
				contract.emptyABI = false // reset the flag
				return notifyContext(cmd(AskContractAddress{}))
			}
			if contract.RawABI == nil {
				return notifyContext(cmd(FetchContractABI{}))
			}
			return notifyContext(cmd(RunDecodeContractABI{}))
		}

		if contract.InitialBlock == nil {
			return notifyContext(cmd(FetchContractInitialBlock{}))
		}

		// TODO: can we infer the name from what we find through the ABI discovery?
		// otherwise, ask for a shortname
		if contract.Name == "" {
			return notifyContext(cmd(AskContractName{}))
		}

		if !contract.TrackEvents && !contract.TrackCalls {
			return notifyContext(cmd(AskContractTrackWhat{}))
		}

		if contract.TrackFactory == nil {
			return notifyContext(cmd(AskContractIsFactory{}))
		}

		if *contract.TrackFactory {
			if contract.FactoryCreationEvent == "" {
				return notifyContext(cmd(AskFactoryCreationEvent{}))
			}
			if contract.FactoryCreationEventFieldIdx == nil {
				return notifyContext(cmd(AskFactoryCreationEventField{}))
			}

			dynContract := p.dynamicContractOf(contract.Name)

			if dynContract.Name == "" {
				return notifyContext(cmd(AskDynamicContractName{}))
			}

			if dynContract.parentContract == nil {
				dynContract.parentContract = contract
			}

			if !dynContract.TrackEvents && !dynContract.TrackCalls {
				return notifyContext(cmd(AskDynamicContractTrackWhat{}))
			}
			if dynContract.Abi == nil {
				// if the user pasted an empty ABI, we would restart the process or choosing a contract address
				if dynContract.emptyABI {
					dynContract.referenceContractAddress = "" // reset the reference address
					dynContract.emptyABI = false              // reset the flag
					return notifyContext(cmd(AskContractAddress{}))
				}
				if dynContract.RawABI == nil {
					if dynContract.referenceContractAddress == "" {
						if p.ChainConfig().ApiEndpoint == "" {
							return notifyContext(cmd(AskDynamicContractABI{}))
						}
						return notifyContext(cmd(AskDynamicContractAddress{}))
					}
					return notifyContext(cmd(FetchDynamicContractABI{}))
				}
				return notifyContext(cmd(RunDecodeDynamicContractABI{}))
			}
		}

	}

	p.currentContractIdx = -1

	if !p.ConfirmEnoughContracts {
		return cmd(AskAddContract{})
	}

	return cmd(codegen.RunGenerate{})
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	switch msg := msg.(type) {
	case codegen.MsgStart:
		var msgCmd loop.Cmd
		if msg.Hydrate != nil {
			if err := json.Unmarshal([]byte(msg.Hydrate.SavedState), &c.State); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, here's an error message to share with our devs (%s); we've notified them already`, err))
			}

			if err := validateIncomingState(c.State); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, the initial state has not been validated: %w`, err))
			}

			msgCmd = c.Msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.Msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, c.NextStep())

	case codegen.AskProjectName:
		return c.CmdAskProjectName()

	case codegen.InputProjectName:
		c.State.Name = msg.Value
		return c.NextStep()

	case codegen.AskChainName:
		var labels, values []string
		for _, conf := range ChainConfigs {
			labels = append(labels, conf.DisplayName)
			values = append(values, conf.ID)
		}
		return c.Action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.Msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.State.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
		if isValidChainName(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case StartFirstContract:
		c.State.Contracts = append(c.State.Contracts, &Contract{})
		return c.NextStep()

	case MsgContractSwitch:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		switch {
		case contract.Name != "":
			return c.Msg().Messagef("Ok, now let's talk about contract %q (%s)",
				contract.Name,
				contract.Address,
			).Cmd()
		case contract.Address != "":
			return c.Msg().Messagef("Ok, now let's talk about contract at address %s",
				contract.Address,
			).Cmd()
		default:
			// TODO: humanize ordinal "1st", etc..
			return c.Msg().Messagef("Ok, so there's missing info for the %s contract. Let's fill that in.",
				humanize.Ordinal(c.State.currentContractIdx+1),
			).Cmd()
		}

	case AskContractAddress:
		return loop.Seq(
			c.Msg().Messagef("We're tackling the %s contract.", humanize.Ordinal(c.State.currentContractIdx+1)).Cmd(),
			c.Action(InputContractAddress{}).TextInput("Please enter the contract address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Ethereum address.\nThe default value is the Uniswap v3 factory address.").
				DefaultValue("0x1f98431c8ad98523631ae4a59f267346ea31f984").
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Ethereum address: 0x followed by 40 hex characters.").Cmd(),
		)

	case AskDynamicContractAddress:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		return c.Action(InputDynamicContractAddress{}).TextInput(fmt.Sprintf("Please enter an example contract created by the %q factory", factory.Name), "Submit").
			Description("Format it with 0x prefix and make sure it's a valid Ethereum address.\nThe default value is the USDC/ETH pool address.").
			DefaultValue("0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640").
			Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Ethereum address: 0x followed by 40 hex characters.").Cmd()

	case InputDynamicContractAddress:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.State, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskDynamicContractAddress{}))
		}

		contract := c.State.dynamicContractOf(factory.Name)
		contract.referenceContractAddress = inputAddress

		return c.NextStep()

	case AskContractABI:
		return c.Action(InputContractABI{}).TextInput(fmt.Sprintf("Please paste the contract ABI or the full JSON ABI file path starting with %sfullpath/to/Abi.json", AbiFilepathPrefix), "Submit").
			Cmd()

	case AskDynamicContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		return c.Action(InputDynamicContractABI{}).TextInput(fmt.Sprintf("Please paste the ABI for contracts that will be created by the event %q", contract.FactoryCreationEventName()), "Submit").
			Cmd()

	case InputContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		// if the user pasted and empty string or hit the enter button by not supplying anything,
		// we want to go back to the ABI question
		if msg.Value == "" {
			contract.emptyABI = true
			return c.NextStep()
		}

		var rawMessage json.RawMessage

		if strings.HasPrefix(msg.Value, AbiFilepathPrefix) {
			abiPath := strings.TrimPrefix(msg.Value, AbiFilepathPrefix)

			fileBytes, err := os.ReadFile(abiPath)
			if err != nil {
				return loop.Seq(c.Msg().Messagef("Cannot read the ABI file %q: %s", abiPath, err).Cmd(), cmd(AskContractABI{}))
			}

			rawMessage = json.RawMessage(fileBytes)
		} else {
			rawMessage = json.RawMessage(msg.Value)
		}

		if _, err := json.Marshal(rawMessage); err != nil {
			return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABI{}))
		}

		contract.RawABI = rawMessage

		return c.NextStep()

	case InputDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		contract := c.State.dynamicContractOf(factory.Name)

		rawMessage := json.RawMessage(msg.Value)
		if _, err := json.Marshal(rawMessage); err != nil {
			return loop.Seq(c.Msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABI{}))
		}

		contract.RawABI = rawMessage
		return c.NextStep()

	case InputContractAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.State, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskContractAddress{}))
		}

		contract.Address = inputAddress

		return c.NextStep()

	case MsgInvalidContractAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		return c.Msg().
			Messagef("Input address isn't valid : %q", msg.Err).
			Cmd()

	case FetchContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		config := c.State.ChainConfig()
		if config.ApiEndpoint == "" {
			return cmd(AskContractABI{})
		}

		return func() loop.Msg {
			abi, err := contract.FetchABI(config)
			return ReturnFetchContractABI{abi: abi, err: err}
		}

	case ReturnFetchContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		if msg.err != nil {
			return loop.Seq(
				c.Msg().Messagef("Cannot fetch the ABI for contract %q (%s)", contract.Address, msg.err).Cmd(),
				cmd(AskContractABI{}),
			)
		}
		contract.RawABI = []byte(msg.abi)
		contract.abiFetchedInThisSession = true
		return c.NextStep()

	case FetchDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		contract := c.State.dynamicContractOf(factory.Name)
		config := c.State.ChainConfig()
		if config.ApiEndpoint == "" {
			return cmd(AskDynamicContractABI{})
		}
		return func() loop.Msg {
			abi, err := contract.FetchABI(c.State.ChainConfig())
			return ReturnFetchDynamicContractABI{abi: abi, err: err}
		}

	case ReturnFetchDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		contract := c.State.dynamicContractOf(factory.Name)
		if msg.err != nil {
			return loop.Seq(
				c.Msg().Messagef("Cannot fetch the ABI for contract %q (%s)", contract.referenceContractAddress, msg.err).Cmd(),
				cmd(AskDynamicContractABI{}),
			)
		}
		contract.RawABI = []byte(msg.abi)
		contract.abiFetchedInThisSession = true
		return c.NextStep()

	case RunDecodeContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		return CmdDecodeABI(contract)

	case ReturnRunDecodeContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		if msg.Err != nil {
			return loop.Quit(fmt.Errorf("decoding ABI for contract %q: %w", contract.Name, msg.Err))
		}
		contract.Abi = msg.Abi
		evt := contract.EventModels()
		calls := contract.CallModels()

		if !contract.abiFetchedInThisSession {
			return c.NextStep()
		}

		// the 'printf' is a hack because we can't do arithmetics in the template
		// it means '+1'
		peekABI := c.Msg().MessageTpl(`Ok, here's what the ABI would produce:

`+"```"+`protobuf
// Events
{{- range .events }}
message {{.Proto.MessageName}} {{.Proto.OutputModuleFieldName}} {
  {{- range $idx, $field := .Proto.Fields }}
  {{$field.Type}} {{$field.Name}} = {{ len (printf "a%*s" $idx "") }};
  {{- end}}
}
{{- end}}

// Calls
{{- range .calls }}
message {{.Proto.MessageName}} {{.Proto.OutputModuleFieldName}} {
  {{- range $idx, $field := .Proto.Fields }}
  {{$field.Type}} {{$field.Name}} = {{ len (printf "a%*s" $idx "") }};
  {{- end}}
}
{{- end}}
`+"```"+`
`, map[string]any{"events": evt, "calls": calls}).Cmd()
		return loop.Seq(peekABI, cmd(AskConfirmContractABI{}))

	case AskConfirmContractABI:
		return c.Action(InputConfirmContractABI{}).
			Confirm("Do you want to proceed with this ABI?", "Yes", "No").
			Cmd()

	case InputConfirmContractABI:
		if msg.Affirmative {
			return c.NextStep()
		}
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		contract.RawABI = nil
		contract.abiFetchedInThisSession = false
		return cmd(AskContractABI{})

	case RunDecodeDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		contract := c.State.dynamicContractOf(factory.Name)
		return cmdDecodeDynamicABI(contract)

	case ReturnRunDecodeDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		if msg.err != nil {
			return loop.Quit(fmt.Errorf("decoding ABI for dynamic contract of %q: %w", factory.Name, msg.err))
		}
		contract := c.State.dynamicContractOf(factory.Name)
		contract.Abi = msg.abi
		evt := contract.EventModels()
		calls := contract.CallModels()

		if !contract.abiFetchedInThisSession {
			return c.NextStep()
		}
		// the 'printf' is a hack because we can't do arithmetics in the template
		// it means '+1'
		peekABI := c.Msg().MessageTpl(`Ok, here's what the ABI would produce:

`+"```"+`protobuf
// Events
{{- range .events }}
message {{.Proto.MessageName}} {{.Proto.OutputModuleFieldName}} {
  {{- range $idx, $field := .Proto.Fields }}
  {{$field.Type}} ({{$field.Name}}) = {{ len (printf "a%*s" $idx "") }};
  {{- end}}
}
{{- end}}

// Calls
{{- range .calls }}
message {{.Proto.MessageName}} {{.Proto.OutputModuleFieldName}} {
  {{- range $idx, $field := .Proto.Fields }}
  {{$field.Type}} ({{$field.Name}}) = {{ len (printf "a%*s" $idx "") }};
  {{- end}}
}
{{- end}}
`+"```"+`
		`, map[string]any{"events": evt, "calls": calls}).Cmd()
		return loop.Seq(peekABI, c.NextStep())

	case FetchContractInitialBlock:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		config := c.State.ChainConfig()
		if config.ApiEndpoint == "" {
			return cmd(AskContractInitialBlock{})
		}
		return func() loop.Msg {
			initialBlock, err := contract.FetchInitialBlock(config)
			return ReturnFetchContractInitialBlock{InitialBlock: initialBlock, Err: err}
		}

	case AskContractInitialBlock:
		return c.Action(InputContractInitialBlock{}).TextInput("Please enter the contract initial block number", "Submit").
			Validation(`^\d+$`, "Please enter a valid block number").
			Cmd()

	case InputContractInitialBlock:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		blk, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Seq(
				c.Msg().Messagef("Cannot parse the block number %q: %s", msg.Value, err).Cmd(),
				cmd(AskContractInitialBlock{}),
			)
		}
		contract.InitialBlock = &blk
		return c.NextStep()

	case ReturnFetchContractInitialBlock:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		return c.Action(InputContractInitialBlock{}).TextInput("Please enter the contract initial block number", "Submit").
			DefaultValue(fmt.Sprintf("%d", msg.InitialBlock)).
			Validation(`^\d+$`, "Please enter a valid block number").
			Cmd()

	case AskContractName:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		act := c.Action(InputContractName{}).TextInput(fmt.Sprintf("Choose a short name for the contract at address %q (lowercase and numbers only)", contract.Address), "Submit").
			Description("Lowercase and numbers only").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The name should be short, and contain only lowercase characters and numbers, and not start with a number.")
		if contract.Address == "0x1f98431c8ad98523631ae4a59f267346ea31f984" {
			act = act.DefaultValue("factory")
		}
		return act.Cmd()

	case InputContractName:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		if err := validateContractName(c.State, msg.Value); err != nil {
			return loop.Seq(cmd(MsgInvalidContractName{err}), cmd(AskContractName{}))
		}
		contract.Name = msg.Value
		return c.NextStep()

	case MsgInvalidContractName:
		return c.Msg().
			Messagef("Invalid contract name: %q", msg.Err).
			Cmd()

	case AskDynamicContractName:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		act := c.Action(InputDynamicContractName{}).TextInput(fmt.Sprintf("Choose a short name for the contract that will be created by the factory %q (lowercase and numbers only)", factory.Name), "Submit").
			Description("Lowercase and numbers only").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The name should be short, and contain only lowercase characters and numbers, and not start with a number.")
		if factory.Address == "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640" {
			act = act.DefaultValue("pool")
		}
		return act.Cmd()

	case InputDynamicContractName:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		if err := validateContractName(c.State, msg.Value); err != nil {
			return loop.Seq(cmd(MsgInvalidDynamicContractName{err}), cmd(AskDynamicContractName{}))
		}

		contract := c.State.dynamicContractOf(factory.Name)
		contract.Name = msg.Value
		return c.NextStep()

	case MsgInvalidDynamicContractName:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		return c.Msg().
			Messagef("Invalid dynamic contract name: %q", msg.Err).
			Cmd()

	case AskContractTrackWhat:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		if !c.State.ChainConfig().SupportsCalls {
			contract.TrackEvents = true
			contract.TrackCalls = false
			return c.NextStep()
		}

		return c.Action(InputContractTrackWhat{}).
			ListSelect("What do you want to track for this contract?").
			Labels("Events", "Calls", "Both events and calls").
			Values("events", "calls", "both").
			Cmd()

	case InputContractTrackWhat:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		switch msg.Value {
		case "events":
			contract.TrackEvents = true
		case "calls":
			contract.TrackCalls = true
		case "both":
			contract.TrackEvents = true
			contract.TrackCalls = true
		default:
			return loop.Quit(fmt.Errorf("invalid selection input value %q, expected 'events', 'calls' or 'both'", msg.Value))
		}
		return c.NextStep()

	case AskDynamicContractTrackWhat:
		return c.Action(InputDynamicContractTrackWhat{}).
			ListSelect("What do you want to track for the contracts that will be created by this factory ?").
			Labels("Events", "Calls", "Both events and calls").
			Values("events", "calls", "both").
			Cmd()

	case InputDynamicContractTrackWhat:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		contract := c.State.dynamicContractOf(factory.Name)
		switch msg.Value {
		case "events":
			contract.TrackEvents = true
		case "calls":
			contract.TrackCalls = true
		case "both":
			contract.TrackEvents = true
			contract.TrackCalls = true
		default:
			return loop.Quit(fmt.Errorf("invalid selection input value %q, expected 'events', 'calls' or 'both'", msg.Value))
		}
		return c.NextStep()

	case AskContractIsFactory:
		contract := c.contextContract()
		if !contract.TrackEvents && contract.TrackCalls {
			return loop.Seq(cmd(InputContractIsFactory{}))
		}

		return c.Action(InputContractIsFactory{}).
			Confirm("Is this contract a factory that will create more contracts that you want to track ?", "Yes", "No").
			Cmd()

	case InputContractIsFactory:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		contract.TrackFactory = &msg.Affirmative
		return c.NextStep()

	case AskFactoryCreationEvent:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		events := contract.Abi.EventIDsToSig()

		values := make([]string, 0)

		keys := maps.Keys(events)

		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, k := range keys {
			values = append(values, events[k])
		}

		return c.Action(InputFactoryCreationEvent{}).
			ListSelect("Choose the event signaling a new contract deployment").
			Labels(values...).
			Values(keys...).
			Cmd()

	case InputFactoryCreationEvent:
		contract := c.State.Contracts[c.State.currentContractIdx]

		contract.FactoryCreationEvent = msg.Value
		return c.NextStep()

	case AskFactoryCreationEventField:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		eventFields, err := contract.EventFields(contract.FactoryCreationEvent)
		if err != nil {
			return loop.Quit(fmt.Errorf("cannot get event fields for contract %q: %w", contract.Name, err))
		}

		var params []string
		var indexes []string
		for i, param := range eventFields {
			indexes = append(indexes, fmt.Sprintf("%d", i))
			params = append(params, fmt.Sprintf("%d - %s (%s)", i, param.Name, param.TypeName))
		}

		return loop.Seq(
			c.Msg().
				Message("Great, now which field in the event payload contains the address of the newly created contract?").
				Cmd(),
			c.Action(InputFactoryCreationEventField{}).
				ListSelect("Choose the field containing the contract address").
				Labels(params...).
				Values(indexes...).
				Cmd(),
		)

	case InputFactoryCreationEventField:
		contract := c.State.Contracts[c.State.currentContractIdx]
		idx, err := strconv.ParseInt(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid field index %q: %w", msg.Value, err))
		}
		contract.FactoryCreationEventFieldIdx = &idx
		return c.NextStep()

	case AskAddContract:
		out := []loop.Cmd{
			c.Msg().Message("Current contracts: [" + strings.Join(contractNames(c.State.Contracts), ", ") + "]").Cmd(),
		}

		if len(c.State.DynamicContracts) != 0 {
			out = append(out, c.Msg().Message("Dynamic contracts: ["+strings.Join(dynamicContractNames(c.State.DynamicContracts), ", ")+"]").Cmd())
		}

		out = append(out,
			c.Action(InputAddContract{}).
				Confirm("Add another contract ?", "Yes", "No").
				Cmd())

		return loop.Seq(out...)

	case InputAddContract:
		if msg.Affirmative {
			c.State.Contracts = append(c.State.Contracts, &Contract{})
			c.State.currentContractIdx = len(c.State.Contracts) - 1
		} else {
			c.State.ConfirmEnoughContracts = true
		}
		return c.NextStep()

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)

	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd
