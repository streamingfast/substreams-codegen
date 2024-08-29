package soltransactions

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

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
		"Simplest Substreams to get you started on solana",
		`This creating the most simple substreams on Solana`,
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

	// Remote build part removed for the moment
	// if !p.confirmDoCompile && !p.confirmDownloadOnly {
	// 	return cmd(codegen.AskConfirmCompile{})
	// }

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
			TextInput(fmt.Sprintf("Filter the transactions based on one or several Program IDs.\n\nSupported operators are: logical or '||', logical and '&&' and parenthesis: '()'. \n\nExample: to only consume TRANSACTIONS containing Token or ComputeBudget instructions: 'program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA || program:ComputeBudget111111111111111111111111111111'. \n\nTransactions containing 'Vote111111111111111111111111111111111111111' instructions are always excluded."), "Submit").
			DefaultValue("program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA").
			Cmd()

	case InputProgramId:
		c.state.ProgramId = msg.Value
		fmt.Printf("%s", msg.Value)
		return c.NextStep()

	// Remote build part removed for the moment
	// case codegen.InputConfirmCompile:
	// 	if msg.Affirmative {
	// 		c.state.confirmDoCompile = true
	// 	} else {
	// 		c.state.confirmDownloadOnly = true
	// 	}
	// 	return c.NextStep()

	case codegen.RunGenerate:
		return loop.Seq(
			cmdGenerate(c.state),
		)

	// Remote build part removed for the moment
	// case codegen.AskConfirmCompile:
	// 	return c.action(codegen.InputConfirmCompile{}).
	// 		Confirm("Should we build the Substreams package for you?", "Yes, build it", "No").
	// 		Cmd()

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

	case codegen.RunBuild:
		// Remote build part removed for the moment
		// Do not run the build, the user only wants to download the files
		// if c.state.confirmDownloadOnly {
		// 	return cmd(codegen.ReturnBuild{
		// 		Err:       nil,
		// 		Artifacts: nil,
		// 	})
		// }

		return cmd(codegen.ReturnBuild{
			Err:       nil,
			Artifacts: nil,
		})

		// Remote build part removed for the moment
		// return cmdBuild(c.state)

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
				cmdBuildFailed(nil, errors.New("build response is nil")),
			)
		}

		if resp.Error != "" {
			// dont fail the command line yet, go to the return build step
			return loop.Seq(
				// This is not an error, send a loading false to remove the loading spinner
				c.msg().Loading(false, "").Cmd(),
				cmdBuildFailed(resp.Logs, errors.New(resp.Error)),
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

		if len(resp.Artifacts) == 0 {
			if len(c.remoteBuildState.Logs) == 0 {
				// don't accumulate any empty logs, just keep looping
				return loop.Seq(
					cmd(codegen.CompilingBuild{
						FirstTime:       false,
						RemoteBuildChan: msg.RemoteBuildChan,
					}), // keep staying in the CompilingBuild state
				)
			}

			return cmd(codegen.CompilingBuild{
				FirstTime:       false,
				RemoteBuildChan: msg.RemoteBuildChan,
			})
		}

		// done, we have the artifacts
		return loop.Seq(
			// This is not an error, send a loading false to remove the loading spinner
			c.msg().Loading(false, "").Cmd(),
			cmdBuildCompleted(c.remoteBuildState),
		)

	case codegen.ReturnBuild:
		// Remote build part removed for the moment
		// if msg.Err != nil {
		// 	return loop.Seq(
		// 		c.msg().Messagef("Remote build failed with error: %q. See full logs in `{project-path}/logs.txt`", msg.Err).Cmd(),
		// 		c.msg().Messagef("You will need to unzip the 'substreams-src.zip' file and run `make package` to try and generate the .spkg file.").Cmd(),
		// 		c.action(codegen.PackageDownloaded{}).
		// 			DownloadFiles().
		// 			AddFile("logs.txt", []byte(msg.Logs), `text/x-logs`, "").
		// 			Cmd(),
		// 	)
		// }
		// if c.state.confirmDoCompile {
		// 	return loop.Seq(
		// 		c.msg().Messagef("Build completed successfully, took %s", time.Since(c.state.buildStarted)).Cmd(),
		// 		c.action(codegen.PackageDownloaded{}).
		// 			DownloadFiles().
		// 			// In both AddFile(...) calls, do not show any description, as we already have enough description in the substreams init part of the conversation
		// 			AddFile(msg.Artifacts[0].Filename, msg.Artifacts[0].Content, `application/x-protobuf+sf.substreams.v1.Package`, "").
		// 			AddFile("logs.txt", []byte(msg.Logs), `text/x-logs`, "").
		// 			Cmd(),
		// 	)
		// }

		return loop.Seq(
			c.msg().Message(codegen.ReturnBuildMessage(c.state.Name)).Cmd(),
			loop.Quit(nil),
		)

	case codegen.PackageDownloaded:
		return loop.Quit(nil)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
