package codegen

import (
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type AskProjectName struct{}
type InputProjectName struct{ pbconvo.UserInput_TextInput }

type AskChainName struct{}
type MsgInvalidChainName struct{}
type InputChainName struct{ pbconvo.UserInput_Selection }

type InputSourceDownloaded struct{ pbconvo.UserInput_Confirmation }
type PackageDownloaded struct{ pbconvo.UserInput_Confirmation }

type AskConfirmCompile struct{}
type InputConfirmCompile struct{ pbconvo.UserInput_Confirmation } // SQL specific

type AskInitialStartBlockType struct{}
type InputAskInitialStartBlockType struct{ pbconvo.UserInput_TextInput }

func InputAskInitialStartBlockTypeTextInput() string {
	return "At what block do you want to start indexing data?"
}

func InputAskInitialStartBlockTypeRegex() string {
	return `^\d+$`
}

func InputAskInitialStartBlockTypeValidation() string {
	return "The start block cannot be empty and must be a number"
}

type RunGenerate struct{}

type ReturnGenerate struct {
	Err          error
	ProjectFiles map[string][]byte
}

func (c ReturnGenerate) Error(msg *MsgWrap) loop.Cmd {
	return loop.Seq(
		msg.Messagef("Code generation failed with error: %s", c.Err).Cmd(),
		loop.Quit(c.Err),
	)
}

type MsgGenerateProgress struct {
	Progress int
	Logs     []string

	Continue bool
}
