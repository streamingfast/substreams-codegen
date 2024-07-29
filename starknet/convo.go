package starknet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	codegen "github.com/streamingfast/substreams-codegen"

	"github.com/streamingfast/substreams-codegen/loop"
)

type outputType string

const outputTypeSQL = "sql"

const sqlTypeSQL = "sql"
const sqlTypeClickhouse = "clickhouse"

func init() {
	codegen.RegisterConversation(
		"starknet-sql",
		"Inject Starknet transaction data into a database",
		"",
		codegen.ConversationFactory(NewWithSql),
		20,
	)
}

type Convo struct {
	factory    *codegen.MsgWrapFactory
	state      *Project
	outputType outputType

	remoteBuildState *codegen.RemoteBuildState
}

func NewWithSql(factory *codegen.MsgWrapFactory) codegen.Conversation {
	c := &Convo{
		factory:          factory,
		state:            &Project{},
		remoteBuildState: &codegen.RemoteBuildState{},
		outputType:       outputTypeSQL,
	}
	return c
}

func (c *Convo) msg() *codegen.MsgWrap { return c.factory.NewMsg(c.state) }

func (c *Convo) action(element any) *codegen.MsgWrap {
	return c.factory.NewInput(element, c.state)
}

func (c *Convo) validate() error {
	if _, err := json.Marshal(c.state); err != nil {
		return fmt.Errorf("validating state format: %w", err)
	}

	switch c.outputType {
	case outputTypeSQL:
		//
	default:
		return fmt.Errorf("invalid output type %q (should not happen, this is a bug)", c.outputType)
	}
	c.state.outputType = c.outputType
	return nil
}

func (c *Convo) NextStep() loop.Cmd {
	if err := c.validate(); err != nil {
		return loop.Quit(err)
	}
	return c.state.NextStep()
}

func cmd(msg any) loop.Cmd {
	return func() loop.Msg {
		return msg
	}
}

func (p *Project) NextStep() (out loop.Cmd) {
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}
	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if !isValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if p.TransactionFilter == "" && !p.filterAsked {
		return cmd(AskTransactionFilter{})
	}

	if p.SqlOutputFlavor == "" {
		return cmd(codegen.AskSqlOutputFlavor{})
	}

	if !p.generatedCodeCompleted {
		return cmd(codegen.RunGenerate{})
	}

	if !p.confirmDoCompile && !p.confirmDownloadOnly {
		return cmd(codegen.AskConfirmCompile{})
	}

	return cmd(codegen.RunBuild{})
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
			TextInput("Please enter the project name", "Submit").
			Description("Identifier with only letters and numbers").
			Validation(`^([a-z][a-z0-9_]{0,63})$`, "The project name must be a valid identifier with only letters and numbers, and no spaces").
			Cmd()

	case codegen.InputProjectName:
		c.state.Name = msg.Value
		return c.NextStep()

	case codegen.AskChainName:
		var labels, values []string
		for _, conf := range ChainConfigs {
			labels = append(labels, conf.DisplayName)
			values = append(values, conf.ID)
		}
		return c.action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.state.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.state.ChainName = msg.Value
		if isValidChainName(msg.Value) {
			return loop.Seq(
				c.msg().Messagef("Got it, will be using chain %q", c.state.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case codegen.RunGenerate:
		return loop.Seq(
			c.msg().Message("Generating Substreams module code").Cmd(),
			loop.Batch(
				cmdGenerate(c.state, c.outputType),
			),
		)

	case codegen.ReturnGenerate:
		if msg.Err != nil {
			return loop.Seq(
				c.msg().Message("Build failed!").Cmd(),
				c.msg().Messagef("The build failed with error: %s", msg.Err).Cmd(),
				loop.Quit(msg.Err),
			)
		}

		c.state.projectFiles = msg.ProjectFiles
		c.state.sourceFiles = msg.SourceFiles
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

	case codegen.RunBuild:
		return cmdBuild(c.state)

	case codegen.CompilingBuild:
		resp, ok := <-msg.RemoteBuildChan

		if !ok {
			// the channel has been closed, we are done
			return loop.Seq(
				c.msg().StopLoading().Cmd(),
				cmdBuildCompleted(c.remoteBuildState),
			)
		}

		if resp == nil {
			// dont fail the command line yet, go to the return build step
			return loop.Seq(
				c.msg().StopLoading().Cmd(),
				cmdBuildFailed(errors.New("build response is nil")),
			)
		}

		if resp.Error != "" {
			// dont fail the command line yet, go to the return build step
			return loop.Seq(
				c.msg().StopLoading().Cmd(),
				cmdBuildFailed(errors.New(resp.Error)),
			)
		}

		c.remoteBuildState.Update(resp)

		// the first time, we want to show a message stating that we have started the build
		if msg.FirstTime {
			return loop.Seq(
				c.msg().Loadingf(true, "Compiling your Substreams, build started at %s. This normally takes around 1 minute...", c.state.buildStarted.Format(time.UnixDate)).Cmd(),
				cmd(codegen.CompilingBuild{
					FirstTime:       false,
					RemoteBuildChan: msg.RemoteBuildChan,
				}), // keep staying in the CompilingBuild state
			)
		}

		if c.remoteBuildState.Error != "" {
			// dont fail the command line yet, go to the return build step
			return loop.Seq(
				c.msg().StopLoading().Cmd(),
				cmdBuildFailed(errors.New(c.remoteBuildState.Error)),
			)
		}

		if len(c.remoteBuildState.Artifacts) == 0 {
			if len(c.remoteBuildState.Logs) == 0 {
				// don't accumulate any empty logs, just keep looping
				return cmd(codegen.CompilingBuild{
					FirstTime:       false,
					RemoteBuildChan: msg.RemoteBuildChan,
				}) // keep staying in the CompilingBuild state
			}

			return cmd(codegen.CompilingBuild{
				FirstTime:       false,
				RemoteBuildChan: msg.RemoteBuildChan,
			})
		}

		// done, we have the artifacts
		return loop.Seq(
			c.msg().StopLoading().Cmd(),
			cmdBuildCompleted(c.remoteBuildState),
		)

	case codegen.ReturnBuild:
		if msg.Err != nil {
			return loop.Seq(
				c.msg().Messagef("Remote build failed with error: %s\nYou can package your Substreams with \"make package\".", msg.Err).Cmd(),
				loop.Quit(nil),
			)
		}

		return loop.Seq(
			c.msg().Messagef("Build completed successfully, took %s", time.Since(c.state.buildStarted)).Cmd(),
			c.action(codegen.PackageDownloaded{}).
				DownloadFiles().
				// In both AddFile(...) calls, do not show any description, as we already have enough description in the substreams init part of the conversation
				AddFile(msg.Artifacts[0].Filename, msg.Artifacts[0].Content, `application/x-protobuf+sf.substreams.v1.Package`, "").
				AddFile("logs.txt", []byte(msg.Logs), `text/x-logs`, "").
				Cmd(),
		)

	case codegen.PackageDownloaded:
		return loop.Quit(nil)

	case codegen.AskSqlOutputFlavor:
		return c.action(codegen.InputSQLOutputFlavor{}).ListSelect("Please select the type of SQL output").
			Labels("PostgreSQL", "Clickhouse").
			Values(sqlTypeSQL, sqlTypeClickhouse).
			Cmd()

	case codegen.InputSQLOutputFlavor:
		c.state.SqlOutputFlavor = msg.Value
		return c.NextStep()

	case AskTransactionFilter:
		c.state.filterAsked = true
		return c.action(InputTransactionFilter{}).
			TextInput(`Please enter the transaction filter (leave blank for default: "(rc:execution_status:1)")`, "Submit").
			Cmd()

	case InputTransactionFilter:
		c.state.TransactionFilter = msg.Value // Accept the value directly, even if it's blank
		return c.NextStep()
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
}

func isTestnet(input string) bool {
	return ChainConfigByID[input].Network == "injective-testnet"
}
