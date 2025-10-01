package tronhelloworld

import (
	"encoding/json"
	"fmt"
	"regexp"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/base"
	"github.com/streamingfast/substreams-codegen/loop"
)

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{
			ConversationState: base.ConversationState{
				ChainName: "tron",
			},
		},
	}}
}
func init() {
	codegen.RegisterConversation(
		"tron-hello-world",
		"Example Substreams that reads TRON blocks and extract data from `TransferContracts`.",
		"Use this example as a starting point to create your own custom Substreams, which indexes thed data you need.",
		New,
		59,
		"TRON",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	return cmd(codegen.RunGenerate{})
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	switch msg := msg.(type) {
	case codegen.MsgStart:
		c.SetClientVersion(msg.Version)
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

	case codegen.InputSubstreamsConsumptionChoice:
		return c.HandleSubstreamsConsumptionChoice(msg.Value)

	case codegen.InputSourceDownloaded:
		return c.HandleDownloaded(msg.Value)

	case codegen.AskProjectName:
		return c.CmdAskProjectName()

	case codegen.InputProjectName:
		c.State.Name = msg.Value
		return c.NextStep()

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd

var tronNetworkRegexp = regexp.MustCompile(`^tron`)

func tronNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(tronNetworkRegexp)
}
