package trontransactions

import (
	"fmt"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

var sharedFlowConfig = codegen.SharedFlowConfig{
	ValidChains: chains.TronNetworks(),
}

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
		New,
		59,
		"TRON",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
	}

	p := c.State

	if p.Filter == "" {
		return cmd(AskFilter{})
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
	case AskFilter:
		message := "Input the filter that you want to apply on the transactions. You can filter on the following fields: `contract_type`, `to`, `from`, `contract_address`.\n\n The `&&` and `||` logical operators are supported.\n\nIn the following example, you filter all the transactions of type `TransferContract` and received by `TPFduiaYgyYKPfrtGh3gN8rgbYU5XgfPP2`:\n\n (contract_type:TransferContract && to:TPFduiaYgyYKPfrtGh3gN8rgbYU5XgfPP2)"

		return c.Action(InputFilter{}).
			TextInput(message, "Submit").
			Cmd()

	case InputFilter:
		c.State.Filter = msg.Value

		return c.NextStep()
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd

type AskFilter struct{}
type InvalidFilter struct{ Err error }
type InputFilter struct{ pbconvo.UserInput_TextInput }
