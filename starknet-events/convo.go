package starknet_events

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))

type Convo struct {
	factory          *codegen.MsgWrapFactory
	state            *Project
	remoteBuildState *codegen.RemoteBuildState
}

func init() {
	codegen.RegisterConversation(
		"starknet-events",
		"Filtered and decode desired Starknet events and create a substreams as source",
		"Given a list of contracts and their ABIs, this will build an Starknet substreams that decodes events",
		codegen.ConversationFactory(New),
		72,
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

	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if !p.IsValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if p.Contract.Address == "" {
		return cmd(AskContractAddress{})
	}

	if p.Contract.InitialBlock == nil {
		return cmd(AskContractInitialBlock{})
	}

	if p.Contract.TrackedEvents == nil {
		return cmd(AskEventAddress{})
	}

	if !p.EventsTrackCompleted {
		return cmd(AskEventAddress{})
	}

	//if p.contract.abi == nil {
	//	// if the user pasted an empty ABI, we would restart the process or choosing a contract address
	//	if p.contract.emptyABI {
	//		p.contract.Address = ""     // reset the address
	//		p.contract.emptyABI = false // reset the flag
	//		return cmd(AskContractAddress{})
	//	}
	//
	//	if p.contract.RawABI == nil {
	//		return cmd(FetchContractABI{})
	//	}
	//	return cmd(RunDecodeContractABI{})
	//}

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

	case MsgInvalidContractAddress:
		contract := c.state.Contract
		if contract == nil {
			return QuitInvalidContext
		}
		return c.msg().
			Messagef("Input address isn't valid : %q", msg.Err).
			Cmd()

	case AskContractAddress:
		return loop.Seq(
			c.action(InputContractAddress{}).TextInput("Please enter the contract address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Starknet address.\nFor example, the Ekubo Positions contract address: 0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				DefaultValue("0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Starknet address").Cmd(),
		)

	case AskEventAddress:
		return loop.Seq(
			c.action(InputContractAddress{}).TextInput("Please enter the event address", "Submit").
				Description("Format it with 0x prefix and make sure it's a valid Starknet Event address.\nFor example, the Transfer event address: 0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				DefaultValue("0x02e0af29598b407c8716b17f6d2795eca1b471413fa03fb145a5e33722184067").
				Validation("^0x[a-fA-F0-9]{40}$", "Please enter a valid Starknet address").Cmd(),
		)

	case InputEventAddress:
		contract := c.state.Contract
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		//Change to validateEventAddress
		if err := validateContractAddress(c.state, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidEventAddress{err}), cmd(AskEventAddress{}))
		}

	case InputContractAddress:
		contract := c.state.Contract
		if contract == nil {
			return QuitInvalidContext
		}

		inputAddress := strings.ToLower(msg.Value)
		if err := validateContractAddress(c.state, inputAddress); err != nil {
			return loop.Seq(cmd(MsgInvalidContractAddress{err}), cmd(AskContractAddress{}))
		}

		contract.Address = inputAddress

		return c.NextStep()

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
		if c.state.IsValidChainName(msg.Value) {
			return loop.Seq(
				c.msg().Messagef("Got it, will be using chain %q", c.state.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
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
			c.msg().Message(codegen.ReturnBuildMessage()).Cmd(),
			loop.Quit(nil),
		)

	case codegen.PackageDownloaded:
		return loop.Quit(nil)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

func validateContractAddress(p *Project, address string) error {
	if !strings.HasPrefix(address, "0x") && len(address) == 42 {
		return fmt.Errorf("contract address %s is invalid, it must be a 42 character hex string starting with 0x", address)
	}

	if p.Contract.Address == address {
		return fmt.Errorf("contract address %s already exists in the project", address)
	}

	return nil
}
