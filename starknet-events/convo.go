package starknet_events

import (
	"encoding/json"
	"fmt"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))

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

	if !p.IsValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if p.Contract.Address == "" {
		return cmd(AskContractAddress{})
	}

	if p.Contract.InitialBlock == nil {
		return cmd(AskContractInitialBlock{})
	}

	if p.Contract.TrackedEvents == nil {
		return cmd(AskEventAddress{})
	}

	if !p.EventsTrackCompleted {
		return cmd(AskEventAddress{})
	}

	//if p.contract.abi == nil {
	//	// if the user pasted an empty ABI, we would restart the process or choosing a contract address
	//	if p.contract.emptyABI {
	//		p.contract.Address = ""     // reset the address
	//		p.contract.emptyABI = false // reset the flag
	//		return cmd(AskContractAddress{})
	//	}
	//
	//	if p.contract.RawABI == nil {
	//		return cmd(FetchContractABI{})
	//	}
	//	return cmd(RunDecodeContractABI{})
	//}

	return loop.Quit(nil)
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
		return c.Action(codegen.InputProjectName{}).
			TextInput(codegen.InputProjectNameTextInput(), "Submit").
			Description(codegen.InputProjectNameDescription()).
			DefaultValue("my_project").
			Validation(codegen.InputProjectNameRegex(), codegen.InputProjectNameValidation()).
			Cmd()

	case MsgInvalidContractAddress:
		contract := c.State.Contract
		if contract == nil {
			return QuitInvalidContext
		}
		return c.Msg().
			Messagef("Input address isn't valid : %q", msg.Err).
			Cmd()

	case AskContractAddress:
		return loop.Seq(
			c.Action(InputContractAddress{}).TextInput("Please enter the contract address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Starknet address.\nFor example, the Ekubo Positions contract address: 0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				DefaultValue("0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Starknet address").Cmd(),
		)

	case AskEventAddress:
		return loop.Seq(
			c.Action(InputContractAddress{}).TextInput("Please enter the event address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Starknet Event address.\nFor example, the Transfer event address: 0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				DefaultValue("0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Starknet address").Cmd(),
		)

	case InputEventAddress:
		contract := c.State.Contract
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		//Change to validateEventAddress
		if err := validateContractAddress(c.State, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidEventAddress{err}), cmd(AskEventAddress{}))
		}

	case InputContractAddress:
		contract := c.State.Contract
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.State, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskContractAddress{}))
		}

		contract.Address = inputAddress

		return c.NextStep()

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
		if msg.Err != nil {
			return loop.Seq(
				c.Msg().Messagef("Code generation failed with error: %s", msg.Err).Cmd(),
				loop.Quit(msg.Err),
			)
		}

		c.State.projectFiles = msg.ProjectFiles
		c.State.generatedCodeCompleted = true

		downloadCmd := c.Action(codegen.InputSourceDownloaded{}).DownloadFiles()

		for fileName, fileContent := range msg.ProjectFiles {
			fileDescription := ""
			if _, ok := codegen.FileDescriptions[fileName]; ok {
				fileDescription = codegen.FileDescriptions[fileName]
			}

			downloadCmd.AddFile(fileName, fileContent, "text/plain", fileDescription)
		}

		return loop.Seq(c.Msg().Messagef("Code generation complete!").Cmd(), downloadCmd.Cmd())

	case codegen.InputSourceDownloaded:
		return c.NextStep()
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

func validateContractAddress(p *Project, address string) error {
	if !strings.HasPrefix(address, "0x") && len(address) == 42 {
		return fmt.Errorf("contract address %s is invalid, it must be a 42 character hex string starting with 0x", address)
	}

	if p.Contract.Address == address {
		return fmt.Errorf("contract address %s already exists in the project", address)
	}

	return nil
}
