package mantra_events

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var QuitInvalidContext = loop.Quit(fmt.Errorf("invalid state context: no current contract"))
var MantraTestnetDefaultStartBlock uint64 = 1

var mantraNetworkRegexp = regexp.MustCompile(`^mantra`)

var sharedFlowConfig = codegen.SharedFlowConfig{
	ValidChains: networks.GetSubstreamsRegistry().Search(mantraNetworkRegexp),
}

type Convo struct {
	*codegen.Conversation[*Project]
}

func init() {
	codegen.RegisterConversation(
		"mantra-events",
		"Stream Mantra Events with specific attributes if specified",
		"Create an Mantra Substreams module from specific events",
		New,
		70,
		"Cosmos",
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
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
	}

	p := c.State
	p.DataType = "events"

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

	return c.NextPostSharedFlowStep(sharedFlowConfig)
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	if c.IsPreSharedFlowMsg(msg, sharedFlowConfig) {
		return c.UpdatePreSharedFlowMsg(msg, sharedFlowConfig, c.NextStep)
	}

	if c.IsPostSharedFlowMsg(msg, sharedFlowConfig) {
		return c.UpdatePostSharedFlowMsg(msg, sharedFlowConfig)
	}

	switch msg := msg.(type) {
	case AskInitialStartBlockType:
		textInputMessage := "At what block do you want to start indexing data?"
		defaultValue := "0"
		if c.State.IsChainTestnet() {
			defaultValue = fmt.Sprintf("%d", MantraTestnetDefaultStartBlock)
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
		if c.State.IsChainTestnet() && initialBlock < MantraTestnetDefaultStartBlock {
			initialBlock = MantraTestnetDefaultStartBlock
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
			ListSelect(fmt.Sprintf("This codegen will build a substreams that filters data based on events.\n"+
				"Do you want to target:"), "target_type").
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
			TextInput(fmt.Sprintf("Please enter the type of Event that you want to track.\n\nYou can usually find them under the transaction details in the explorer: %s.\nExamples: message, wasm ...", c.State.ChainExplorerLink()), "Submit").
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
				ListSelect("Do you want to add another event type", "other_event_type").
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
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd

func mantraNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(mantraNetworkRegexp)
}
