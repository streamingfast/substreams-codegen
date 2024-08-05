package ethfull

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"golang.org/x/exp/maps"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))

type outputType string

const outputTypeSQL = "sql"
const outputTypeSubgraph = "subgraph"
const outputTypeSubstreams = "substreams"

type Convo struct {
	factory    *codegen.MsgWrapFactory
	state      *Project
	outputType outputType

	remoteBuildState *codegen.RemoteBuildState
}

func init() {
	supportedChains := make([]string, 0, len(ChainConfigs))
	for _, conf := range ChainConfigs {
		supportedChains = append(supportedChains, conf.DisplayName)
	}
	codegen.RegisterConversation(
		"evm-events-calls",
		"Decode Ethereum events/calls and generate a fully functional substreams",
		`Given a list of contracts and their ABIs, this will build an Ethereum substreams that decodes events and/or calls.
Supported networks: `+strings.Join(supportedChains, ", "),
		codegen.ConversationFactory(NewOnlySubstreams),
		80,
	)
	codegen.RegisterConversation(
		"evm-subgraph",
		"Decode Ethereum events/calls and and use them as triggers to feed your Subgraph",
		`Given a list of contracts and their ABIs, this will build an Ethereum substreams that decodes events and/or calls
and creates entities that can be sent directly into a Subgraph.
Supported networks: `+strings.Join(supportedChains, ", "),
		codegen.ConversationFactory(NewWithSubgraph),
		80,
	)

	codegen.RegisterConversation(

		"evm-sql",
		"Decode Ethereum events/calls and insert them into PostgreSQL or Clickhouse",
		`Given a list of contracts and their ABIs, this will build an Ethereum substreams that can be used to fill

	your SQL database using https://github.com/streamingfast/substreams-sink-sql.
	Supported networks: `+strings.Join(supportedChains, ", "),

		codegen.ConversationFactory(NewWithSQL),
		60,
	)
}

func NewWithSQL(factory *codegen.MsgWrapFactory) codegen.Conversation {
	h := &Convo{
		state:            &Project{currentContractIdx: -1},
		factory:          factory,
		outputType:       outputTypeSQL,
		remoteBuildState: &codegen.RemoteBuildState{},
	}
	return h
}

func NewWithSubgraph(factory *codegen.MsgWrapFactory) codegen.Conversation {
	h := &Convo{
		state:            &Project{currentContractIdx: -1},
		factory:          factory,
		outputType:       outputTypeSubgraph,
		remoteBuildState: &codegen.RemoteBuildState{},
	}
	return h
}

func NewOnlySubstreams(factory *codegen.MsgWrapFactory) codegen.Conversation {
	h := &Convo{
		state:            &Project{currentContractIdx: -1},
		factory:          factory,
		outputType:       outputTypeSubstreams,
		remoteBuildState: &codegen.RemoteBuildState{},
	}
	return h
}

func (h *Convo) msg() *codegen.MsgWrap { return h.factory.NewMsg(h.state) }
func (h *Convo) action(element any) *codegen.MsgWrap {
	return h.factory.NewInput(element, h.state)
}

func cmd(msg any) loop.Cmd {
	return func() loop.Msg {
		return msg
	}
}

// This function does NOT mutate anything. Only reads.

func (c *Convo) contextContract() *Contract {
	if c.state.currentContractIdx == -1 || c.state.currentContractIdx > len(c.state.Contracts)-1 {
		return nil
	}
	return c.state.Contracts[c.state.currentContractIdx]
}

func (c *Convo) validate() error {
	if _, err := json.Marshal(c.state); err != nil {
		return fmt.Errorf("validating state format: %w", err)
	}

	switch c.outputType {
	case outputTypeSQL:
		if c.state.SubgraphOutputFlavor != "" {
			return fmt.Errorf("cannot have SubgraphOutputFlavor set on this code generator")
		}
	case outputTypeSubgraph:
		if c.state.SqlOutputFlavor != "" {
			return fmt.Errorf("cannot have SqlOutputFlavor set on this code generator")
		}
	case outputTypeSubstreams:
	default:
		return fmt.Errorf("invalid output type %q (should not happen, this is a bug)", c.outputType)
	}
	c.state.outputType = c.outputType
	return nil
}

func (c *Convo) NextStep() loop.Cmd {
	if err := c.validate(); err != nil {
		return loop.Quit(err)
	}
	return c.state.NextStep()
}

func (p *Project) NextStep() (out loop.Cmd) {
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

		if contract.abi == nil {
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
			if dynContract.abi == nil {
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

	if !p.confirmEnoughContracts {
		return cmd(AskAddContract{})
	}

	switch p.outputType {
	case outputTypeSQL:
		if p.SqlOutputFlavor == "" {
			return cmd(codegen.AskSqlOutputFlavor{})
		}
	case outputTypeSubgraph:
		if p.SubgraphOutputFlavor == "" {
			return cmd(codegen.AskSubgraphOutputFlavor{})
		}
	}

	if !p.generatedCodeCompleted {
		return cmd(codegen.RunGenerate{})
	}

	if !p.confirmDoCompile && !p.confirmDownloadOnly {
		return cmd(codegen.AskConfirmCompile{})
	}

	return cmd(codegen.RunBuild{})
}

func (p *Project) dynamicContractOf(contractName string) (out *DynamicContract) {
	for _, dynContract := range p.DynamicContracts {
		if dynContract.ParentContractName == contractName {
			out = dynContract
			break
		}
	}
	if out == nil {
		out = &DynamicContract{
			ParentContractName: contractName,
		}
		p.DynamicContracts = append(p.DynamicContracts, out)
	}
	return
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	if os.Getenv("SUBSTREAMS_DEV_DEBUG_CONVERSATION") == "true" {
		fmt.Printf("convo Update message: %T %#v\n-> state: %#v\n\n", msg, msg, c.state)
	}

	switch msg := msg.(type) {
	case codegen.MsgStart:
		var msgCmd loop.Cmd
		if msg.Hydrate != nil {
			if err := json.Unmarshal([]byte(msg.Hydrate.SavedState), &c.state); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, here's an error message to share with our devs (%s); we've notified them already`, err))
			}

			if err := validateIncomingState(c.state); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, the initial state has not been validated: %w`, err))
			}

			msgCmd = c.msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, c.NextStep())

	case codegen.AskProjectName:
		return c.action(codegen.InputProjectName{}).
			TextInput("Please enter the project name", "Submit").
			Description("Identifier with only letters and numbers").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The project name must be a valid identifier with only letters and numbers, and no spaces.").
			Cmd()

	case codegen.InputProjectName:
		c.state.Name = msg.Value
		return c.NextStep()

	case codegen.AskChainName:
		var labels, values []string
		for _, conf := range ChainConfigs {
			labels = append(labels, conf.DisplayName)
			values = append(values, conf.ID)
		}
		return c.action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.state.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.state.ChainName = msg.Value
		if isValidChainName(msg.Value) {
			return loop.Seq(
				c.msg().Messagef("Got it, will be using chain %q", c.state.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case StartFirstContract:
		c.state.Contracts = append(c.state.Contracts, &Contract{})
		return c.NextStep()

	case MsgContractSwitch:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		switch {
		case contract.Name != "":
			return c.msg().Messagef("Ok, now let's talk about contract %q (%s)",
				contract.Name,
				contract.Address,
			).Cmd()
		case contract.Address != "":
			return c.msg().Messagef("Ok, now let's talk about contract at address %s",
				contract.Address,
			).Cmd()
		default:
			// TODO: humanize ordinal "1st", etc..
			return c.msg().Messagef("Ok, so there's missing info for the %s contract. Let's fill that in.",
				humanize.Ordinal(c.state.currentContractIdx+1),
			).Cmd()
		}

	case AskContractAddress:
		return loop.Seq(
			c.msg().Messagef("We're tackling the %s contract.", humanize.Ordinal(c.state.currentContractIdx+1)).Cmd(),
			c.action(InputContractAddress{}).TextInput("Please enter the contract address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Ethereum address.\nFor example, the Uniswap v3 factory address: 0x1f98431c8ad98523631ae4a59f267346ea31f984").
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Ethereum address").Cmd(),
		)

	case AskDynamicContractAddress:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		return c.action(InputDynamicContractAddress{}).TextInput(fmt.Sprintf("Please enter an example contract created by the %q factory", factory.Name), "Submit").
			Description("Format it with 0x prefix and make sure it's a valid Ethereum address.\nFor example, the UNI/ETH pool at: 0x1d42064fc4beb5f8aaf85f4617ae8b3b5b8bd801").
			Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Ethereum address").Cmd()

	case InputDynamicContractAddress:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.state, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskDynamicContractAddress{}))
		}

		contract := c.state.dynamicContractOf(factory.Name)
		contract.referenceContractAddress = inputAddress

		return c.NextStep()

	case AskContractABI:
		return c.action(InputContractABI{}).TextInput("Please paste the contract ABI", "Submit").
			Cmd()

	case AskDynamicContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		return c.action(InputDynamicContractABI{}).TextInput(fmt.Sprintf("Please paste the ABI for contracts that will be created by the event %q", contract.FactoryCreationEventName()), "Submit").
			Cmd()

	case InputContractABI:
		// FIXME: dedupe all these QuitInvalidContext!
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		rawMessage := json.RawMessage(msg.Value)
		if _, err := json.Marshal(rawMessage); err != nil {
			return loop.Seq(c.msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABI{}))
		}

		contract.RawABI = rawMessage

		return c.NextStep()

	case InputDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		contract := c.state.dynamicContractOf(factory.Name)

		rawMessage := json.RawMessage(msg.Value)
		if _, err := json.Marshal(rawMessage); err != nil {
			return loop.Seq(c.msg().Messagef("ABI %q isn't valid: %q", msg.Value, err).Cmd(), cmd(AskContractABI{}))
		}

		contract.RawABI = rawMessage
		return c.NextStep()

	case InputContractAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.state, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskContractAddress{}))
		}

		contract.Address = inputAddress

		return c.NextStep()

	case MsgInvalidContractAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		return c.msg().
			Messagef("Input address isn't valid : %q", msg.Err).
			Cmd()

	case FetchContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		config := c.state.ChainConfig()
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
				c.msg().Messagef("Cannot fetch the ABI for contract %q (%s)", contract.Address, msg.err).Cmd(),
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
		contract := c.state.dynamicContractOf(factory.Name)
		config := c.state.ChainConfig()
		if config.ApiEndpoint == "" {
			return cmd(AskDynamicContractABI{})
		}
		return func() loop.Msg {
			abi, err := contract.FetchABI(c.state.ChainConfig())
			return ReturnFetchDynamicContractABI{abi: abi, err: err}
		}

	case ReturnFetchDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		contract := c.state.dynamicContractOf(factory.Name)
		if msg.err != nil {
			return loop.Seq(
				c.msg().Messagef("Cannot fetch the ABI for contract %q (%s)", contract.referenceContractAddress, msg.err).Cmd(),
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
		return cmdDecodeABI(contract)

	case ReturnRunDecodeContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		if msg.err != nil {
			return loop.Quit(fmt.Errorf("decoding ABI for contract %q: %w", contract.Name, msg.err))
		}
		contract.abi = msg.abi
		evt := contract.EventModels()
		calls := contract.CallModels()

		if !contract.abiFetchedInThisSession {
			return c.NextStep()
		}

		// the 'printf' is a hack because we can't do arithmetics in the template
		// it means '+1'
		peekABI := c.msg().MessageTpl(`Ok, here's what the ABI would produce:

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
		return loop.Seq(peekABI, cmd(AskConfirmContractABI{}))

	case AskConfirmContractABI:
		return c.action(InputConfirmContractABI{}).
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
		contract := c.state.dynamicContractOf(factory.Name)
		return cmdDecodeDynamicABI(contract)

	case ReturnRunDecodeDynamicContractABI:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		if msg.err != nil {
			return loop.Quit(fmt.Errorf("decoding ABI for dynamic contract of %q: %w", factory.Name, msg.err))
		}
		contract := c.state.dynamicContractOf(factory.Name)
		contract.abi = msg.abi
		evt := contract.EventModels()
		calls := contract.CallModels()

		if !contract.abiFetchedInThisSession {
			return c.NextStep()
		}
		// the 'printf' is a hack because we can't do arithmetics in the template
		// it means '+1'
		peekABI := c.msg().MessageTpl(`Ok, here's what the ABI would produce:

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
		config := c.state.ChainConfig()
		if config.ApiEndpoint == "" {
			return cmd(AskContractInitialBlock{})
		}
		return func() loop.Msg {
			initialBlock, err := contract.FetchInitialBlock(config)
			return ReturnFetchContractInitialBlock{InitialBlock: initialBlock, Err: err}
		}

	case AskContractInitialBlock:
		return c.action(InputContractInitialBlock{}).TextInput("Please enter the contract initial block number", "Submit").
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
				c.msg().Messagef("Cannot parse the block number %q: %s", msg.Value, err).Cmd(),
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

		return c.action(InputContractInitialBlock{}).TextInput("Please enter the contract initial block number", "Submit").
			DefaultValue(fmt.Sprintf("%d", msg.InitialBlock)).
			Validation(`^\d+$`, "Please enter a valid block number").
			Cmd()

	case AskContractName:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		return c.action(InputContractName{}).TextInput(fmt.Sprintf("Choose a short name for the contract at address %q (lowercase and numbers only)", contract.Address), "Submit").
			Description("Lowercase and numbers only").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The name should be short, and contain only lowercase characters and numbers, and not start with a number.").Cmd()

	case InputContractName:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		if err := validateContractName(c.state, msg.Value); err != nil {
			return loop.Seq(cmd(MsgInvalidContractName{err}), cmd(AskContractName{}))
		}
		contract.Name = msg.Value
		return c.NextStep()

	case MsgInvalidContractName:
		return c.msg().
			Messagef("Invalid contract name: %q", msg.Err).
			Cmd()

	case AskDynamicContractName:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		return c.action(InputDynamicContractName{}).TextInput(fmt.Sprintf("Choose a short name for the contract that will be created by the factory %q (lowercase and numbers only)", factory.Name), "Submit").
			Description("Lowercase and numbers only").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The name should be short, and contain only lowercase characters and numbers, and not start with a number.").Cmd()

	case InputDynamicContractName:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		if err := validateContractName(c.state, msg.Value); err != nil {
			return loop.Seq(cmd(MsgInvalidDynamicContractName{err}), cmd(AskDynamicContractName{}))
		}

		contract := c.state.dynamicContractOf(factory.Name)
		contract.Name = msg.Value
		return c.NextStep()

	case MsgInvalidDynamicContractName:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}

		return c.msg().
			Messagef("Invalid dynamic contract name: %q", msg.Err).
			Cmd()

	case AskContractTrackWhat:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		if !c.state.ChainConfig().SupportsCalls {
			contract.TrackEvents = true
			contract.TrackCalls = false
			return c.NextStep()
		}

		return c.action(InputContractTrackWhat{}).
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
		return c.action(InputDynamicContractTrackWhat{}).
			ListSelect("What do you want to track for the contracts that will be created by this factory ?").
			Labels("Events", "Calls", "Both events and calls").
			Values("events", "calls", "both").
			Cmd()

	case InputDynamicContractTrackWhat:
		factory := c.contextContract()
		if factory == nil {
			return QuitInvalidContext
		}
		contract := c.state.dynamicContractOf(factory.Name)
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

		return c.action(InputContractIsFactory{}).
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

		events := contract.abi.EventIDsToSig()

		values := make([]string, 0)

		keys := maps.Keys(events)

		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, k := range keys {
			values = append(values, events[k])
		}

		return c.action(InputFactoryCreationEvent{}).
			ListSelect("Choose the event signaling a new contract deployment").
			Labels(values...).
			Values(keys...).
			Cmd()

	case InputFactoryCreationEvent:
		contract := c.state.Contracts[c.state.currentContractIdx]

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
			c.msg().
				Message("Great, now which field in the event payload contains the address of the newly created contract?").
				Cmd(),
			c.action(InputFactoryCreationEventField{}).
				ListSelect("Choose the field containing the contract address").
				Labels(params...).
				Values(indexes...).
				Cmd(),
		)

	case InputFactoryCreationEventField:
		contract := c.state.Contracts[c.state.currentContractIdx]
		idx, err := strconv.ParseInt(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid field index %q: %w", msg.Value, err))
		}
		contract.FactoryCreationEventFieldIdx = &idx
		return c.NextStep()

	case AskAddContract:
		out := []loop.Cmd{
			c.msg().Message("Current contracts: [" + strings.Join(contractNames(c.state.Contracts), ", ") + "]").Cmd(),
		}

		if len(c.state.DynamicContracts) != 0 {
			out = append(out, c.msg().Message("Dynamic contracts: ["+strings.Join(dynamicContractNames(c.state.DynamicContracts), ", ")+"]").Cmd())
		}

		out = append(out,
			c.action(InputAddContract{}).
				Confirm("Add another contract ?", "Yes", "No").
				Cmd())

		return loop.Seq(out...)

	case InputAddContract:
		if msg.Affirmative {
			c.state.Contracts = append(c.state.Contracts, &Contract{})
			c.state.currentContractIdx = len(c.state.Contracts) - 1
		} else {
			c.state.confirmEnoughContracts = true
		}
		return c.NextStep()

	case codegen.AskSqlOutputFlavor:
		return c.action(codegen.InputSQLOutputFlavor{}).ListSelect("Please select the type of SQL output").
			Labels("PostgreSQL", "Clickhouse").
			Values("sql", "clickhouse").
			Cmd()

	case codegen.InputSQLOutputFlavor:
		c.state.SqlOutputFlavor = msg.Value
		return c.NextStep()

	case codegen.AskSubgraphOutputFlavor:
		return c.action(codegen.InputSubgraphOutputFlavor{}).ListSelect("Choose how you want to use feed your subgraph.").
			Labels("Triggers (allows you to write logic in your subgraph in AssemblyScript)", "Entities (direct, no AssemblyScript code allowed)").
			Values("trigger", "entity").
			Cmd()

	case codegen.InputSubgraphOutputFlavor:
		c.state.SubgraphOutputFlavor = msg.Value
		return c.NextStep()

	case codegen.InputConfirmCompile:
		if msg.Affirmative {
			c.state.confirmDoCompile = true
		} else {
			c.state.confirmDownloadOnly = true
		}
		return c.NextStep()

	case codegen.RunGenerate:
		return loop.Seq(
			cmdGenerate(c.state, c.outputType),
		)

	case codegen.AskConfirmCompile:
		return c.action(codegen.InputConfirmCompile{}).
			Confirm("Should we build the Substreams package for you?", "Yes, build it", "No").
			Cmd()

	case codegen.ReturnGenerate:
		if msg.Err != nil {
			return loop.Seq(
				c.msg().Messagef("Code generation failed with error: %s", msg.Err).Cmd(),
				loop.Quit(msg.Err),
			)
		}

		c.state.sourceFiles = msg.SourceFiles
		c.state.projectFiles = msg.ProjectFiles
		c.state.generatedCodeCompleted = true

		downloadCmd := c.action(codegen.InputSourceDownloaded{}).DownloadFiles()

		for fileName, fileContent := range msg.SourceFiles {
			fileDescription := ""
			if _, ok := codegen.FileDescriptions[fileName]; ok {
				fileDescription = codegen.FileDescriptions[fileName]
			}

			downloadCmd.AddFile(fileName, fileContent, "text/plain", fileDescription)
		}

		for fileName, fileContent := range msg.ProjectFiles {
			fileDescription := ""
			if _, ok := codegen.FileDescriptions[fileName]; ok {
				fileDescription = codegen.FileDescriptions[fileName]
			}

			downloadCmd.AddFile(fileName, fileContent, "text/plain", fileDescription)
		}

		return loop.Seq(c.msg().Messagef("Code generation complete!").Cmd(), downloadCmd.Cmd())

	case codegen.InputSourceDownloaded:
		return c.NextStep()

	case codegen.RunBuild:
		// Do not run the build, the user only wants to download the files
		if c.state.confirmDownloadOnly {
			return cmd(codegen.ReturnBuild{
				Err:       nil,
				Artifacts: nil,
			})
		}

		return cmdBuild(c.state)

	case codegen.CompilingBuild:
		resp, ok := <-msg.RemoteBuildChan

		if !ok {
			// the channel has been closed, we are done
			return loop.Seq(
				c.msg().StopLoading().Cmd(),
				cmdBuildCompleted(c.remoteBuildState),
			)
		}

		if resp == nil {
			// dont fail the command line yet, go to the return build step
			return loop.Seq(
				c.msg().StopLoading().Cmd(),
				cmdBuildFailed(nil, errors.New("build response is nil")),
			)
		}

		if resp.Error != "" {
			// dont fail the command line yet, go to the return build step
			return loop.Seq(
				// This is not an error, send a loading false to remove the loading spinner
				c.msg().Loading(false, "").Cmd(),
				cmdBuildFailed(resp.Logs, errors.New(resp.Error)),
			)
		}

		c.remoteBuildState.Update(resp)

		// the first time, we want to show a message stating that we have started the build
		if msg.FirstTime {
			return loop.Seq(
				c.msg().Loadingf(true, "Compiling your Substreams, build started at %s. This normally takes around 1 minute...", c.state.buildStarted.Format(time.UnixDate)).Cmd(),
				cmd(codegen.CompilingBuild{
					FirstTime:       false,
					RemoteBuildChan: msg.RemoteBuildChan,
				}), // keep staying in the CompilingBuild state
			)
		}

		if len(resp.Artifacts) == 0 {
			if len(c.remoteBuildState.Logs) == 0 {
				// don't accumulate any empty logs, just keep looping
				return loop.Seq(
					cmd(codegen.CompilingBuild{
						FirstTime:       false,
						RemoteBuildChan: msg.RemoteBuildChan,
					}), // keep staying in the CompilingBuild state
				)
			}

			return cmd(codegen.CompilingBuild{
				FirstTime:       false,
				RemoteBuildChan: msg.RemoteBuildChan,
			})
		}

		// done, we have the artifacts
		return loop.Seq(
			// This is not an error, send a loading false to remove the loading spinner
			c.msg().Loading(false, "").Cmd(),
			cmdBuildCompleted(c.remoteBuildState),
		)

	case codegen.ReturnBuild:
		if msg.Err != nil {
			if c.state.outputType == outputTypeSubstreams {
				return loop.Seq(
					c.msg().Messagef("Remote build failed with error: %q. See full logs in `{project-path}/logs.txt`", msg.Err).Cmd(),
					c.msg().Messagef("You will need to pack your substreams using `substreams pack` command").Cmd(),
					c.action(codegen.PackageDownloaded{}).
						DownloadFiles().
						AddFile("logs.txt", []byte(msg.Logs), `text/x-logs`, "").
						Cmd(),
				)
			}

			return loop.Seq(
				c.msg().Messagef("Remote build failed with error: %q. See full logs in `{project-path}/logs.txt`", msg.Err).Cmd(),
				c.msg().Messagef("You will need to unzip the 'substreams-src.zip' file and run `make package` to try and generate the .spkg file.").Cmd(),
				c.action(codegen.PackageDownloaded{}).
					DownloadFiles().
					AddFile("logs.txt", []byte(msg.Logs), `text/x-logs`, "").
					Cmd(),
			)
		}
		if c.state.confirmDoCompile {
			return loop.Seq(
				c.msg().Messagef("Build completed successfully, took %s", time.Since(c.state.buildStarted)).Cmd(),
				c.action(codegen.PackageDownloaded{}).
					DownloadFiles().
					// In both AddFile(...) calls, do not show any description, as we already have enough description in the substreams init part of the conversation
					AddFile(msg.Artifacts[0].Filename, msg.Artifacts[0].Content, `application/x-protobuf+sf.substreams.v1.Package`, "").
					AddFile("logs.txt", []byte(msg.Logs), `text/x-logs`, "").
					Cmd(),
			)
		}

		if c.state.outputType == outputTypeSubstreams {
			return loop.Seq(
				c.msg().Messagef("Substreams Package was not compiled: You will need to pack your substreams using `substreams pack` command").Cmd(),
				loop.Quit(nil),
			)
		}

		return loop.Seq(
			c.msg().Messagef("Substreams Package was not compiled: You will need to unzip the 'substreams-src.zip' file and run `make package` to generate the .spkg file.").Cmd(),
			loop.Quit(nil),
		)

	case codegen.PackageDownloaded:
		return loop.Quit(nil)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
