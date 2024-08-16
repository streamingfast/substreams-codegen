package codegen

import (
	"time"

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

// SQL specific
type AskSqlOutputFlavor struct{}
type InputSQLOutputFlavor struct{ pbconvo.UserInput_Selection }

// Subgraph specific
type AskSubgraphOutputFlavor struct{}
type InputSubgraphOutputFlavor struct{ pbconvo.UserInput_Selection }

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

func ReturnBuildMessage() string {
	return "Substreams Package successfully downloaded. Open the created README.md for more information on how to build and run your Substreams."
}
