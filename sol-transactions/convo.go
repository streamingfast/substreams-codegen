package soltransactions

import (
	"encoding/json"
	"fmt"
	"strconv"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

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
		"sol-transactions",
		"Get Solana transactions filtered by one or several Program IDs",
		"Allows you to specified a regex containing the Program IDs used to filter the Solana transactions",
		codegen.ConversationFactory(New),
		100,
	)
}

var cmd = codegen.Cmd

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
	}

	if p.ProgramId == "" {
		return cmd(AskProgramId{})
	}

	if !p.generatedCodeCompleted {
		return cmd(codegen.RunGenerate{})
	}

	return cmd(ShowInstructions{})
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

	case AskProgramId:
		return c.Action(InputProgramId{}).
			TextInput(fmt.Sprintf("Filter the transactions based on one or several Program IDs.\nSupported operators are: logical or '||', logical and '&&' and parenthesis: '()'. \nExample: to only consume TRANSACTIONS containing Token or ComputeBudget instructions: 'program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA || program:ComputeBudget111111111111111111111111111111'."), "Submit").
			DefaultValue("program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA").
			Cmd()

	case InputProgramId:
		c.State.ProgramId = msg.Value
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

	case ShowInstructions:
		return loop.Seq(
			c.Msg().Message(codegen.ReturnBuildMessage(false)).Cmd(),
			loop.Quit(nil),
		)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
