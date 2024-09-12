package injectiveminimal

import (
	"encoding/json"
	"fmt"
	"strconv"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var InjectiveTestnetDefaultStartBlock uint64 = 37368800

type Convo struct {
	*codegen.Conversation[*Project]
}

func init() {
	codegen.RegisterConversation(
		"injective-minimal",
		"Simplest Substreams to get you started on Injective Mainnet",
		"This creating the most simple substreams on Injective Mainnet",
		codegen.ConversationFactory(New),
		72,
	)
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
}

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

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
	}

	if !p.generatedCodeCompleted {
		return cmd(codegen.RunGenerate{})
	}

	// Remote build part removed for the moment
	// if !p.confirmDoCompile && !p.confirmDownloadOnly {
	// 	return cmd(codegen.AskConfirmCompile{})
	// }

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

	case codegen.AskInitialStartBlockType:
		textInputMessage := "At what block do you want to start indexing data?"
		defaultValue := "0"
		if c.State.IsTestnet(c.State.ChainName) {
			defaultValue = fmt.Sprintf("%d", InjectiveTestnetDefaultStartBlock)
			textInputMessage = fmt.Sprintf("At what block do you want to start indexing data? (the first available block on %s is: %s)", c.State.ChainName, defaultValue)
		}
		return c.Action(codegen.InputAskInitialStartBlockType{}).
			TextInput(textInputMessage, "Submit").
			DefaultValue(defaultValue).
			Validation(codegen.InputAskInitialStartBlockTypeRegex(), codegen.InputAskInitialStartBlockTypeValidation()).
			Cmd()

	case codegen.InputAskInitialStartBlockType:
		initialBlock, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid start block input value %q, expected a number", msg.Value))
		}
		if c.State.IsTestnet(c.State.ChainName) && initialBlock < InjectiveTestnetDefaultStartBlock {
			initialBlock = InjectiveTestnetDefaultStartBlock
		}

		c.State.InitialBlock = initialBlock
		c.State.InitialBlockSet = true
		return c.NextStep()

	// case codegen.InputConfirmCompile:
	// 	if msg.Affirmative {
	// 		c.State.confirmDoCompile = true
	// 	} else {
	// 		c.State.confirmDownloadOnly = true
	// 	}
	// 	return c.NextStep()

	case codegen.RunGenerate:
		return loop.Seq(
			codegen.CmdGenerate(c.State.Generate),
		)

	// Remote build part removed for the moment
	// case codegen.AskConfirmCompile:
	// 	return c.Action(codegen.InputConfirmCompile{}).
	// 		Confirm("Should we build the Substreams package for you?", "Yes, build it", "No").
	// 		Cmd()

	case codegen.ReturnGenerate:
		if msg.Err != nil {
			return loop.Seq(
				c.Msg().Messagef("Code generation failed with error: %s", msg.Err).Cmd(),
				loop.Quit(msg.Err),
			)
		}

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

func cmd(msg any) loop.Cmd {
	return func() loop.Msg {
		return msg
	}
}
