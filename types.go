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

func ReturnBuildMessage(isMinimal bool) string {
	var minimalStr string

	if isMinimal {
		minimalStr = "* Inspect and edit the the `./src/lib.rs` file\n"
	}

	return cli.Dedent(fmt.Sprintf(
		"Your Substreams project is ready! Follow the next steps to start streaming:\n\n"+
			"%s"+
			"* Build it: `substreams build`\n"+
			"* Authenticate: `substreams auth`\n"+
			"* Stream it: `substreams gui`\n\n"+
			"* Build a *Subgraph* from this substreams: `substreams codegen subgraph`\n"+
			"* Feed your SQL database with this substreams: `substreams codegen sql`\n",
		minimalStr))
}
