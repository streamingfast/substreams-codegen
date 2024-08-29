package codegen

import (
	"fmt"
	"time"

	"github.com/streamingfast/cli"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
)

type RemoteBuildState struct {
	BuildStartedAt time.Time

	Artifacts []*pbbuild.BuildResponse_BuildArtifact
	Error     string
	Logs      []string
}

func (c *RemoteBuildState) Update(resp *RemoteBuildState) bool {
	c.BuildStartedAt = resp.BuildStartedAt
	c.Error = resp.Error

	if resp.Logs != nil {
		c.Logs = append(c.Logs, resp.Logs...)
	}

	if resp.Artifacts != nil {
		c.Artifacts = append(c.Artifacts, resp.Artifacts...)
	}

	return false
}

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
	SourceFiles  map[string][]byte
	ProjectFiles map[string][]byte
}

type RunBuild struct {
	pbconvo.UserInput_Confirmation
}

type MsgGenerateProgress struct {
	Progress int
	Logs     []string

	Continue bool
}

type CompilingBuild struct {
	FirstTime       bool
	RemoteBuildChan chan *RemoteBuildState
}

type ReturnBuild struct {
	Err       error
	Logs      string
	Artifacts []*pbbuild.BuildResponse_BuildArtifact
}

func ReturnBuildMessage(projectName string) string {
	return cli.Dedent(fmt.Sprintf(
		"Your Substreams project is ready! Now follow the next steps:\n\n"+
			"Inspect the 'lib.rs' file, and build with:\n\n"+
			"`substreams build`\n\n"+
			"Authenticate with:\n\n"+
			"`substreams auth`\n\n"+
			"Then start streaming data with:\n\n"+
			"`substreams gui`\n\n"+
			"If you want to have a substreams powered subgraph, run:\n\n"+
			"`substreams codegen subgraph`\n\n"+
			"If you want to generate an SQL sink, run:\n\n"+
			"`substreams codegen sql`",
		projectName))
}
