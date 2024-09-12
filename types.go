package codegen

import (
	"fmt"

	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type AskProjectName struct{}
type InputProjectName struct{ pbconvo.UserInput_TextInput }

func InputProjectNameTextInput() string {
	return "Please enter the project name"
}

func InputProjectNameDescription() string {
	return "Identifier with only lowercase letters, numbers and underscores, up to 64 characters."
}

func InputProjectNameRegex() string {
	return "^([a-z][a-z0-9_]{0,63})$"
}

func InputProjectNameValidation() string {
	return "The project name must be a valid identifier with only lowercase letters, numbers and underscores, up to 64 characters."
}

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

// func (c ReturnGenerate) AddFiles(state any, action *MsgWrap, msg *MsgWrap) loop.Cmd {
// 	downloadCmd := action(codegen.InputSourceDownloaded{}).DownloadFiles()

// 	for fileName, fileContent := range c.ProjectFiles {
// 		fileDescription := ""
// 		if _, ok := FileDescriptions[fileName]; ok {
// 			fileDescription = FileDescriptions[fileName]
// 		}

// 		downloadCmd.AddFile(fileName, fileContent, "text/plain", fileDescription)
// 	}

// 	return loop.Seq(msg.Messagef("Code generation complete!").Cmd(), downloadCmd.Cmd())

// }

type MsgGenerateProgress struct {
	Progress int
	Logs     []string

	Continue bool
}

func ReturnBuildMessage(isMinimal bool) string {
	// TODO: this isn't a `Build` message output, it's just a Generate final message.
	// It's also not standardized.
	var minimalStr string

	if isMinimal {
		minimalStr = "* Inspect and edit the the `./src/lib.rs` file\n"
	}

	return fmt.Sprintf(
		"Your Substreams project is ready! Follow the next steps to start streaming:\n\n"+
			"%s"+
			"\n    substreams build\n"+
			"    substreams auth\n"+
			"    substreams gui\n\n"+
			"    substreams codegen subgraph\n"+
			"    substreams codegen sql\n",
		minimalStr)
}
