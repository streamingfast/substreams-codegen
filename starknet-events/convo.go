package starknet_events

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))
var AbiFilepathPrefix = "file://"

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
}

func init() {
	codegen.RegisterConversation(
		"starknet-events",
		"Filtered and decode desired Starknet events and create a substreams as source",
		"Given a list of contracts and their ABIs, this will build an Starknet substreams that decodes events",
		codegen.ConversationFactory(New),
		72,
	)
}

var cmd = codegen.Cmd

func (c *Convo) NextStep() loop.Cmd {
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

		if contract.Abi == nil || contract.Abi.decodedAbi == nil {
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
			return notifyContext(cmd(AskContractInitialBlock{}))
		}

		// TODO: can we infer the name from what we find through the ABI discovery?
		// otherwise, ask for a shortname
		if contract.Name == "" {
			return notifyContext(cmd(AskContractName{}))
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

			msgCmd = c.Msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.Msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, c.NextStep())

	case codegen.AskProjectName:
		return c.CmdAskProjectName()

	case FetchContractABI:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		config := c.State.ChainConfig()
		if config.EndpointEnvVar == "" {
			return cmd(AskContractABI{})
		}

		return func() loop.Msg {
			abi, err := contract.fetchABI(config)
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

	case MsgInvalidContractAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		return c.Msg().
			Messagef("Input address isn't valid : %q", msg.Err).
			Cmd()

	case AskContractABI:
		return c.Action(InputContractABI{}).TextInput(fmt.Sprintf("Please paste the contract ABI or the full JSON ABI file path starting with %sfullpath/to/Abi.json", AbiFilepathPrefix), "Submit").
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

	case AskAddContract:
		out := []loop.Cmd{
			c.Msg().Message("Current contracts: [" + strings.Join(contractNames(c.State.Contracts), ", ") + "]").Cmd(),
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

	case AskContractName:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}
		act := c.Action(InputContractName{}).TextInput(fmt.Sprintf("Choose a short name for the contract at address %q (lowercase and numbers only)", contract.Address), "Submit").
			Description("Lowercase and numbers only").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The name should be short, and contain only lowercase characters and numbers, and not start with a number.")
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

		contract.SetAliases()

		if !contract.abiFetchedInThisSession {
			return c.NextStep()
		}

		peekABI := c.Msg().Message(string(contract.RawABI)).Cmd()
		return loop.Seq(peekABI, cmd(AskConfirmContractABI{}))

	case AskContractAddress:
		return loop.Seq(
			c.Action(InputContractAddress{}).TextInput("Please enter the contract address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Starknet address.\nFor example, the Ekubo Positions contract address: 0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				DefaultValue("0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				Validation("^0x(0{0,63}[a-fA-F0-9]{1,63}|0{64})$", "Please enter a valid Starknet address").Cmd(),
		)

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

	case InputEventAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		//Change to validateEventAddress
		if err := validateContractAddress(c.State, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidEventAddress{err}), cmd(AskEventAddress{}))
		}

	case InputContractAddress:
		contract := c.contextContract()
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.State, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskContractAddress{}))
		}

		contract.handleContractAddress(inputAddress)

		return c.NextStep()

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
		if c.State.IsValidChainName(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

func validateContractAddress(p *Project, address string) error {
	if !strings.HasPrefix(address, "0x") && len(address) == 42 {
		return fmt.Errorf("contract address %s is invalid, it must be a 42 character hex string starting with 0x", address)
	}

	for _, contract := range p.Contracts {
		if contract.Address == address {
			return fmt.Errorf("contract address %s already exists in the project", address)
		}
	}
	return nil
}

func (c *Convo) contextContract() *Contract {
	p := c.State
	if p.currentContractIdx == -1 || p.currentContractIdx > len(p.Contracts)-1 {
		return nil
	}
	return p.Contracts[p.currentContractIdx]
}
