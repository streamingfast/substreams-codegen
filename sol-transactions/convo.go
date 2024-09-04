package soltransactions

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

type Convo struct {
	factory          *codegen.MsgWrapFactory
	state            *Project
	remoteBuildState *codegen.RemoteBuildState
}

func init() {
	codegen.RegisterConversation(
		"sol-transactions",
		"Get Solana transactions filtered by one or several Program IDs.",
		`Allows you to specified a regex containing the Program IDs used to filter the Solana transactions.`,
		codegen.ConversationFactory(New),
		100,
	)
}

func New(factory *codegen.MsgWrapFactory) codegen.Conversation {
	h := &Convo{
		state:            &Project{},
		factory:          factory,
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

func (c *Convo) validate() error {
	if _, err := json.Marshal(c.state); err != nil {
		return fmt.Errorf("validating state format: %w", err)
	}
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

			msgCmd = c.msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, c.NextStep())

	case codegen.AskProjectName:
		return c.action(codegen.InputProjectName{}).
			TextInput(codegen.InputProjectNameTextInput(), "Submit").
			Description(codegen.InputProjectNameDescription()).
			DefaultValue("my_project").
			Validation(codegen.InputProjectNameRegex(), codegen.InputProjectNameValidation()).
			Cmd()

	case codegen.InputProjectName:
		c.state.Name = msg.Value
		return c.NextStep()

	case codegen.AskInitialStartBlockType:
		return c.action(codegen.InputAskInitialStartBlockType{}).
			TextInput(codegen.InputAskInitialStartBlockTypeTextInput(), "Submit").
			DefaultValue("0").
			Validation(codegen.InputAskInitialStartBlockTypeRegex(), codegen.InputAskInitialStartBlockTypeValidation()).
			Cmd()

	case codegen.InputAskInitialStartBlockType:
		initialBlock, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid start block input value %q, expected a number", msg.Value))
		}

		c.state.InitialBlock = initialBlock
		c.state.InitialBlockSet = true
		return c.NextStep()

	case AskProgramId:
		return c.action(InputProgramId{}).
			TextInput(fmt.Sprintf("Filter the transactions based on one or several Program IDs.\nSupported operators are: logical or '||', logical and '&&' and parenthesis: '()'. \nExample: to only consume TRANSACTIONS containing Token or ComputeBudget instructions: 'program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA || program:ComputeBudget111111111111111111111111111111'. \nTransactions containing 'Vote111111111111111111111111111111111111111' instructions are always excluded."), "Submit").
			DefaultValue("program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA").
			Cmd()

	case InputProgramId:
		c.state.ProgramId = msg.Value
		return c.NextStep()

	case codegen.RunGenerate:
		return loop.Seq(
			cmdGenerate(c.state),
		)

	case codegen.ReturnGenerate:
		if msg.Err != nil {
			return loop.Seq(
				c.msg().Messagef("Code generation failed with error: %s", msg.Err).Cmd(),
				loop.Quit(msg.Err),
			)
		}

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

	case ShowInstructions:
		return loop.Seq(
			c.msg().Message(codegen.ReturnBuildMessage(false)).Cmd(),
			loop.Quit(nil),
		)

	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
