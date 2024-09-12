package ethminimal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

func init() {
	supportedChains := make([]string, 0, len(ChainConfigs))
	for _, conf := range ChainConfigs {
		supportedChains = append(supportedChains, conf.DisplayName)
	}
	codegen.RegisterConversation(
		"evm-minimal",
		"Simplest Substreams to get you started on EVM and compatible chains",
		`Supported networks: `+strings.Join(supportedChains, ", "),
		codegen.ConversationFactory(New),
		83,
	)
}

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
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

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
	}

	if !p.generatedCodeCompleted {
		return cmd(codegen.RunGenerate{})
	}

	return loop.Quit(nil)
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
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

	case codegen.AskInitialStartBlockType:
		return c.Action(codegen.InputAskInitialStartBlockType{}).
			TextInput(codegen.InputAskInitialStartBlockTypeTextInput(), "Submit").
			DefaultValue("0").
			Validation(codegen.InputAskInitialStartBlockTypeRegex(), codegen.InputAskInitialStartBlockTypeValidation()).
			Cmd()

	case codegen.InputAskInitialStartBlockType:
		initialBlock, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid start block input value %q, expected a number", msg.Value))
		}

		c.State.InitialBlock = initialBlock
		c.State.InitialBlockSet = true
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

var cmd = codegen.Cmd
