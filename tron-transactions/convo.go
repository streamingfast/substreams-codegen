package trontransactions

import (
	"encoding/json"
	"fmt"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
}
func init() {
	codegen.RegisterConversation(
		"tron-transactions",
		"Creates a Substreams that outputs a transactions based on your provided parameters.",
		"Given a few parameters, you will get a project that indexes transactions.",
		codegen.ConversationFactory(New),
		59,
		"Tron",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if p.Filter == "" {
		return cmd(AskFilter{})
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

	case AskFilter:
		message := "Input the filter that you want to apply on the transactions. You can filter on the following fields: `contract_type`, `to`, `from`, `contract_address`.\n\n The `&&` and `||` logical operators are supported.\n\nIn the following example, you filter all the transactions of type `TransferContract` and received by `TPFduiaYgyYKPfrtGh3gN8rgbYU5XgfPP2`:\n\n (contract_type:TransferContract && to:TPFduiaYgyYKPfrtGh3gN8rgbYU5XgfPP2)"

		return c.Action(InputFilter{}).
			TextInput(message, "Submit").
			Cmd()

	case InputFilter:
		c.State.Filter = msg.Value

		return c.NextStep()

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd
