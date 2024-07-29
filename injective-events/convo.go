package injective_events

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))

type outputType string

const outputTypeSQL = "sql"
const outputTypeSubgraph = "subgraph"

type InjectiveConvo struct {
	factory    *codegen.MsgWrapFactory
	state      *Project
	outputType outputType

	remoteBuildState *codegen.RemoteBuildState
}

func init() {
	codegen.RegisterConversation(
		"injective-subgraph",
		"Insert Injective events into a Graph-Node subgraph",
		"Create an Injective Substreams module from specific events that can feed a subgraph.",
		codegen.ConversationFactory(NewWithSubgraph),
		40,
	)
	codegen.RegisterConversation(
		"injective-sql",
		"Insert Injective events into PostgreSQL or Clickhouse",
		"Given a list of events, generate the SQL schema and the Substreams module to insert them into a SQL database",
		codegen.ConversationFactory(NewWithSQL),
		20,
	)
}

func NewWithSubgraph(factory *codegen.MsgWrapFactory) codegen.Conversation {
	c := &InjectiveConvo{
		factory:          factory,
		state:            &Project{},
		remoteBuildState: &codegen.RemoteBuildState{},
		outputType:       outputTypeSubgraph,
	}
	return c
}

func NewWithSQL(factory *codegen.MsgWrapFactory) codegen.Conversation {
	c := &InjectiveConvo{
		factory:          factory,
		state:            &Project{},
		remoteBuildState: &codegen.RemoteBuildState{},
		outputType:       outputTypeSQL,
	}
	return c
}

func (c *InjectiveConvo) msg() *codegen.MsgWrap { return c.factory.NewMsg(c.state) }

func (c *InjectiveConvo) action(element any) *codegen.MsgWrap {
	return c.factory.NewInput(element, c.state)
}

func (c *InjectiveConvo) validate() error {
	if _, err := json.Marshal(c.state); err != nil {
		return fmt.Errorf("validating state format: %w", err)
	}

	switch c.outputType {
	case outputTypeSQL:
		if c.state.SubgraphOutputFlavor != "" {
			return fmt.Errorf("cannot have SubgraphOutputFlavor set on this code generator")
		}
	case outputTypeSubgraph:
		if c.state.SqlOutputFlavor != "" {
			return fmt.Errorf("cannot have SqlOutputFlavor set on this code generator")
		}
	default:
		return fmt.Errorf("invalid output type %q (should not happen, this is a bug)", c.outputType)
	}
	c.state.outputType = c.outputType
	return nil
}

func cmd(msg any) loop.Cmd {
	return func() loop.Msg {
		return msg
	}
}

func (c *InjectiveConvo) contextEventDesc() *eventDesc {
	if c.state.currentEventIdx > len(c.state.EventDescs)-1 {
		return nil
	}
	return c.state.EventDescs[c.state.currentEventIdx]
}

func (c *InjectiveConvo) NextStep() loop.Cmd {
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

	if !isValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if !p.InitialBlockSet {
		return cmd(AskInitialStartBlockType{})
	}

	switch p.DataType {
	case "":
		return cmd(AskDataType{})
	case "events", "event_groups":
	case "transactions":
		if p.outputType == outputTypeSQL {
			return loop.Quit(fmt.Errorf("transactions data type is not supported for SQL output"))
		}
	default:
		return loop.Quit(fmt.Errorf("invalid data type %q", p.DataType))
	}

	if len(p.EventDescs) == 0 {
		p.currentEventIdx = -1
		p.EventDescs = append(p.EventDescs, &eventDesc{Incomplete: true})
	}

	previousEventIdx := p.currentEventIdx
	for idx, evt := range p.EventDescs {
		p.currentEventIdx = idx
		notifyContext := func(next loop.Cmd) loop.Cmd {
			if previousEventIdx != p.currentEventIdx {
				return loop.Seq(cmd(MsgEventSwitch{}), next)
			}
			return next
		}
		if evt.EventType == "" {
			return notifyContext(cmd(AskEventType{}))
		}

		if evt.Incomplete {
			return notifyContext(cmd(AskEventAttribute{}))
		}
	}

	if !p.eventsComplete {
		return cmd(AskAnotherEventType{})
	}

	if p.outputType == outputTypeSQL && p.SqlOutputFlavor == "" {
		return cmd(codegen.AskSqlOutputFlavor{})
	}

	// Hacky way of setting the trigger, change this once we add the entities support
	if p.outputType == outputTypeSubgraph && p.SubgraphOutputFlavor == "" {
		p.SubgraphOutputFlavor = "trigger"
	}

	return cmd(codegen.RunGenerate{})
}

func (c *InjectiveConvo) Update(msg loop.Msg) loop.Cmd {
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

	case AskInitialStartBlockType:
		textInputMessage := "At what block do you want to start indexing data?"
		defaultValue := "0"
		if isTestnet(c.state.ChainName) {
			defaultValue = "27751658"
			textInputMessage = fmt.Sprintf("At what block do you want to start indexing data? (the first available block on %s is: %s)", c.state.ChainName, defaultValue)
		}
		return c.action(InputAskInitialStartBlockType{}).
			TextInput(textInputMessage, "Submit").
			DefaultValue(defaultValue).
			Validation(`^\d+$`, "The start block cannot be empty and must be a number").
			Cmd()

	case InputAskInitialStartBlockType:
		initialBlock, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid start block input value %q, expected a number", msg.Value))
		}
		if isTestnet(c.state.ChainName) && initialBlock < 27751658 {
			initialBlock = 27751658
		}

		c.state.InitialBlock = initialBlock
		c.state.InitialBlockSet = true
		return c.NextStep()

	case AskDataType:
		labels := []string{
			"Specific events",
			"All events in transactions where at least one event matches your query",
		}
		values := []string{EVENTS_DATA_TYPE, EVENT_GROUPS_DATA_TYPE}
		if c.outputType == outputTypeSubgraph {
			labels = append(labels, "Full transactions where at least one event matches your query")
			values = append(values, TRXS_DATA_TYPE)
		}
		return c.action(InputDataType{}).
			ListSelect(fmt.Sprintf("This codegen will build a substreams that filters data based on events.\n" +
				"Do you want to target:")).
			Labels(labels...).
			Values(values...).
			Cmd()

	case InputDataType:
		c.state.DataType = msg.Value
		return c.NextStep()

	case AskEventType:
		var cmds []loop.Cmd

		if len(c.state.EventDescs) == 0 {
			cmds = append(cmds, c.msg().Message("Let's start by filtering event types").Cmd())
		}

		cmds = append(cmds, c.action(InputEventType{}).
			TextInput(fmt.Sprintf("Please enter the type of Event that you want to track.\n\nYou can usually find them under the transaction details in the explorer: %s.\nExamples: message, injective.exchange.v1beta1.EventCancelDerivativeOrder, wasm ...", c.state.ChainConfig().ExplorerLink), "Submit").
			Validation(`(.|\s)*\S(.|\s)*`, "The event type cannot be empty").
			Cmd(),
		)

		return loop.Seq(cmds...)

	case InputEventType:
		evt := c.contextEventDesc()
		if evt == nil {
			return QuitInvalidContext
		}
		evt.EventType = strings.TrimSpace(msg.Value)
		return c.NextStep()

	case AskEventAttribute:
		evt := c.contextEventDesc()
		if evt == nil {
			return QuitInvalidContext
		}
		textInput := fmt.Sprintf("Do you want the substreams to match only %q events that contain specific attributes ?\n"+
			"Enter either {attribute_key} or {attribute_key}:{attribute_value} to add such a constraint, or leave empty to skip.", evt.EventType)

		if len(evt.Attributes) > 0 {
			textInput = fmt.Sprintf("Do you want to add another attribute constraint to matching the %q event ? (current conditions: %q)\n"+
				"All conditions must be met for an event to match. You can define additional events of the same type to match different conditions.\n"+
				"Enter either {attribute_key} or {attribute_key}:{attribute_value} to add such a constraint, or leave empty to skip.", evt.EventType, evt.GetEventQuery())
		}
		return c.action(InputEventAttribute{}).
			TextInput(textInput,
				"Submit").
			Cmd()

	case InputEventAttribute:
		evt := c.contextEventDesc()
		if evt == nil {
			return QuitInvalidContext
		}
		if msg.Value == "" {
			evt.Incomplete = false
			return c.NextStep()
		}
		if evt.Attributes == nil {
			evt.Attributes = make(map[string]string)
		}
		kv := strings.SplitN(msg.Value, ":", 2)
		k := kv[0]
		v := ""
		if len(kv) == 2 {
			v = kv[1]
		}
		evt.Attributes[k] = v
		return c.NextStep()

	case AskAnotherEventType:
		return loop.Seq(
			c.msg().Messagef("Current filtering event types %q", c.state.GetEventsQuery()).Cmd(),
			c.action(InputAskAnotherEventType{}).
				ListSelect("Do you want to add another event type").
				Labels("Yes", "No").
				Values("yes", "no").Cmd(),
		)

	case InputAskAnotherEventType:
		switch msg.Value {
		case "yes":
			c.state.EventDescs = append(c.state.EventDescs, &eventDesc{Incomplete: true})
			return c.NextStep()
		case "no":
			c.state.eventsComplete = true
			return c.NextStep()
		default:
			return loop.Quit(fmt.Errorf("invalid selection input value %q, expected 'yes', 'more' or 'no'", msg.Value))
		}

	case MsgEventSwitch:
		evt := c.contextEventDesc()
		if evt == nil {
			return QuitInvalidContext
		}
		if evt.EventType != "" {
			return c.msg().Messagef("Ok, now let's talk about event %q",
				evt.EventType,
			).Cmd()
		}
		return nil

	case codegen.AskSqlOutputFlavor:
		return c.action(codegen.InputSQLOutputFlavor{}).ListSelect("Please select the type of SQL output").
			Labels("PostgreSQL", "Clickhouse").
			Values("sql", "clickhouse").
			Cmd()

	case codegen.InputSQLOutputFlavor:
		c.state.SqlOutputFlavor = msg.Value
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

		c.state.projectZip = msg.ProjectZip
		c.state.sourceZip = msg.SubstreamsSourceZip

		var cmds []loop.Cmd
		cmds = append(cmds, c.msg().Message("Code generation complete!").Cmd())
		cmds = append(cmds, c.action(codegen.RunBuild{}).DownloadFiles().
			AddFile("project.zip", msg.ProjectZip, "application/x-zip+extract", "\nProject files, schemas, dev environment...").
			AddFile("substreams_src.zip", msg.SubstreamsSourceZip, "application/x-zip+extract", "").
			Cmd())

		c.state.generatedCodeCompleted = true
		return loop.Seq(cmds...)

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

	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
}

func isTestnet(input string) bool {
	return ChainConfigByID[input].Network == "injective-testnet"
}
