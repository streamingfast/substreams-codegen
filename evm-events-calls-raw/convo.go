package evm_events_calls_raw

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

const UNISWAP_V3_FACTORY_ADDRESS = "0x1f98431c8ad98523631ae4a59f267346ea31f984"

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))

func init() {
	codegen.RegisterConversation(
		"evm-events-calls-raw",
		"(without ABI) Get raw Ethereum events/calls and create a Substreams as source",
		"Given a list of contract addresses, ge the raw events and calls, without using an ABI",
		codegen.ConversationFactory(New),
		82,
		"evm",
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
		act := c.Action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...)
		act.DefaultValue("mainnet")
		return act.
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
				DefaultValue(c.State.ChainConfig().ExampleContract).
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Ethereum address: 0x followed by 40 hex characters.").Cmd(),
		)

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
			return ReturnFetchContractInitialBlock{InitialBlock: 0, Err: nil}
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
		if contract.Address == c.State.ChainConfig().ExampleContract {
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
		act := c.Action(InputContractTrackWhat{}).
			ListSelect("What do you want to track for this contract?").
			Labels("Events", "Calls", "Both events and calls").
			Values("events", "calls", "both")
		if contract.Address == UNISWAP_V3_FACTORY_ADDRESS {
			act = act.DefaultValue("events")
		}
		return act.Cmd()

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

	case AskAddContract:
		message := []string{
			"Configured contracts: [" + strings.Join(contractNames(c.State.Contracts), ", ") + "]",
		}
		if len(c.State.DynamicContracts) != 0 {
			message = append(message, "Dynamically created contracts: ["+strings.Join(dynamicContractNames(c.State.DynamicContracts), ", ")+"]")
		}

		return loop.Seq(
			c.Msg().Message(strings.Join(message, "\n")).Cmd(),
			c.Action(InputAddContract{}).
				Confirm("Add another contract ?", "Yes", "No").
				Cmd(),
		)

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
