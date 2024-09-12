package injective_events

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))
var InjectiveTestnetDefaultStartBlock uint64 = 37368800

type Convo struct {
	*codegen.Conversation[*Project]
}

func init() {
	codegen.RegisterConversation(
		"injective-events",
		"Stream Injective Events with specific attributes if specified",
		"Create an Injective Substreams module from specific events",
		codegen.ConversationFactory(New),
		70,
	)
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
}

func (c *Convo) contextEventDesc() *eventDesc {
	if c.State.currentEventIdx > len(c.State.EventDescs)-1 {
		return nil
	}
	return c.State.EventDescs[c.State.currentEventIdx]
}

func (c *Convo) NextStep() (out loop.Cmd) {
	p := c.State

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

	if !p.EventsComplete {
		return cmd(AskAnotherEventType{})
	}

	if !p.generatedCodeCompleted {
		return cmd(codegen.RunGenerate{})
	}

	return loop.Quit(nil)
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	switch msg := msg.(type) {
	case codegen.MsgStart:
		var msgCmd loop.Cmd
		if msg.Hydrate != nil {
			if err := json.Unmarshal([]byte(msg.Hydrate.SavedState), &c.State); err != nil {
				return loop.Quit(fmt.Errorf(`something went wrong, here's an error message to share with our devs (%s); we've notified them already`, err))
			}
			msgCmd = c.Msg().Message("Ok, I reloaded your state.").Cmd()
		} else {
			msgCmd = c.Msg().Message("Ok, let's start a new package.").Cmd()
		}
		return loop.Seq(msgCmd, c.NextStep())

	case codegen.AskProjectName:
		return c.Action(codegen.InputProjectName{}).
			TextInput(codegen.InputProjectNameTextInput(), "Submit").
			Description(codegen.InputProjectNameDescription()).
			DefaultValue("my_project").
			Validation(codegen.InputProjectNameRegex(), codegen.InputProjectNameValidation()).
			Cmd()

	case codegen.InputProjectName:
		c.State.Name = msg.Value
		return c.NextStep()

	case codegen.AskChainName:
		var labels, values []string
		for _, conf := range ChainConfigs {
			labels = append(labels, conf.DisplayName)
			values = append(values, conf.ID)
		}
		return c.Action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.Msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.State.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
		if isValidChainName(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case AskInitialStartBlockType:
		textInputMessage := "At what block do you want to start indexing data?"
		defaultValue := "0"
		if isTestnet(c.State.ChainName) {
			defaultValue = fmt.Sprintf("%d", InjectiveTestnetDefaultStartBlock)
			textInputMessage = fmt.Sprintf("At what block do you want to start indexing data? (the first available block on %s is: %s)", c.State.ChainName, defaultValue)
		}
		return c.Action(InputAskInitialStartBlockType{}).
			TextInput(textInputMessage, "Submit").
			DefaultValue(defaultValue).
			Validation(`^\d+$`, "The start block cannot be empty and must be a number").
			Cmd()

	case InputAskInitialStartBlockType:
		initialBlock, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid start block input value %q, expected a number", msg.Value))
		}
		if isTestnet(c.State.ChainName) && initialBlock < InjectiveTestnetDefaultStartBlock {
			initialBlock = InjectiveTestnetDefaultStartBlock
		}

		c.State.InitialBlock = initialBlock
		c.State.InitialBlockSet = true
		return c.NextStep()

	case AskDataType:
		labels := []string{
			"Specific events",
			"All events in transactions where at least one event matches your query",
		}
		values := []string{EVENTS_DATA_TYPE, EVENT_GROUPS_DATA_TYPE}
		return c.Action(InputDataType{}).
			ListSelect(fmt.Sprintf("This codegen will build a substreams that filters data based on events.\n" +
				"Do you want to target:")).
			Labels(labels...).
			Values(values...).
			Cmd()

	case InputDataType:
		c.State.DataType = msg.Value
		return c.NextStep()

	case AskEventType:
		var cmds []loop.Cmd

		if len(c.State.EventDescs) == 0 {
			cmds = append(cmds, c.Msg().Message("Let's start by filtering event types").Cmd())
		}

		cmds = append(cmds, c.Action(InputEventType{}).
			TextInput(fmt.Sprintf("Please enter the type of Event that you want to track.\n\nYou can usually find them under the transaction details in the explorer: %s.\nExamples: message, injective.exchange.v1beta1.EventCancelDerivativeOrder, wasm ...", c.State.ChainConfig().ExplorerLink), "Submit").
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
		return c.Action(InputEventAttribute{}).
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
			c.Msg().Messagef("Current filtering event types %q", c.State.GetEventsQuery()).Cmd(),
			c.Action(InputAskAnotherEventType{}).
				ListSelect("Do you want to add another event type").
				Labels("Yes", "No").
				Values("yes", "no").Cmd(),
		)

	case InputAskAnotherEventType:
		switch msg.Value {
		case "yes":
			c.State.EventDescs = append(c.State.EventDescs, &eventDesc{Incomplete: true})
			return c.NextStep()
		case "no":
			c.State.EventsComplete = true
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
			return c.Msg().Messagef("Ok, now let's talk about event %q",
				evt.EventType,
			).Cmd()
		}
		return nil

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		if msg.Err != nil {
			return loop.Seq(
				c.Msg().Message("Build failed!").Cmd(),
				c.Msg().Messagef("The build failed with error: %s", msg.Err).Cmd(),
				loop.Quit(msg.Err),
			)
		}

		c.State.generatedCodeCompleted = true

		downloadCmd := c.Action(codegen.InputSourceDownloaded{}).DownloadFiles()

		for fileName, fileContent := range msg.ProjectFiles {
			fileDescription := ""
			if _, ok := codegen.FileDescriptions[fileName]; ok {
				fileDescription = codegen.FileDescriptions[fileName]
			}

			downloadCmd.AddFile(fileName, fileContent, "text/plain", fileDescription)
		}

		return loop.Seq(c.Msg().Messagef("Code generation complete!").Cmd(), downloadCmd.Cmd())

	case codegen.InputSourceDownloaded:
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

var cmd = codegen.Cmd
