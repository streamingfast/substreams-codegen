package codegen

import (
	"maps"
	"slices"

	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	"go.uber.org/zap/zapcore"
)

type AskProjectName struct{}
type InputProjectName struct{ pbconvo.UserInput_TextInput }

type AskChainName struct{}
type MsgInvalidChainName struct{}
type InputChainName struct{ pbconvo.UserInput_Selection }

type InputSourceDownloaded struct{ pbconvo.UserInput_TextInput }
type PackageDownloaded struct{ pbconvo.UserInput_Confirmation }

type AskConfirmCompile struct{}
type InputConfirmCompile struct{ pbconvo.UserInput_Confirmation } // SQL specific

type AskInitialStartBlockType struct{}
type InputAskInitialStartBlockType struct{ pbconvo.UserInput_TextInput }

type AskSubstreamsConsumptionChoice struct{}
type InputSubstreamsConsumptionChoice struct{ pbconvo.UserInput_Selection }

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

func (c ReturnGenerate) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddReflected("files", slices.Collect(maps.Keys(c.ProjectFiles)))
	if c.Err != nil {
		enc.AddString("error", c.Err.Error())
	}
	return nil
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

//go:generate go run github.com/abice/go-enum@v0.7.0 -f=$GOFILE --forcelower --names --values --marshal

// SubstreamsSinkChoice represents the type of a Substreams sink choice
// that are possible for a user to select when generating a Substreams package.
//
// Use the lower case versions of the enum when sending it to the user.
// Remain also backward compatible with the previous naming convention.
//
// ENUM(none, postgres, clickhouse, parquet, golang, rust, javascript, python)
type SubstreamsSinkChoice string
