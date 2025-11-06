package soltransactions

import (
	"fmt"
	"strconv"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
	"github.com/streamingfast/substreams-codegen/text"
)

var sharedFlowConfig = codegen.SharedFlowConfig{
	ValidChains: chains.SolanaNetworks(),
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
		"sol-transactions",
		"Get Solana transactions filtered by one or several Program IDs",
		"Allows you to specified a regex containing the Program IDs used to filter the Solana transactions",
		New,
		2001,
		"Solana",
	)
}

var cmd = codegen.Cmd

func (c *Convo) NextStep() loop.Cmd {
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
	}

	p := c.State

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
	}

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
	case codegen.AskInitialStartBlockType:
		return c.Action(codegen.InputAskInitialStartBlockType{}).
			TextInput(codegen.InputAskInitialStartBlockTypeTextInput(), "Submit").
			DefaultValue("0").
			Validation(codegen.InputAskInitialStartBlockTypeRegex(), codegen.InputAskInitialStartBlockTypeValidation()).
			Cmd()

	case codegen.InputAskInitialStartBlockType:
		initialBlock, err := strconv.ParseUint(msg.Value, 10, 64)
		if err != nil {
			return loop.Quit(fmt.Errorf("invalid start block input value %q, expected a number", msg.Value))
		}

		c.State.InitialBlock = initialBlock
		c.State.InitialBlockSet = true
		return c.NextStep()

	case AskFilter:
		return c.Action(InputFilter{}).
			TextInput(text.Dedent(`
				Query to filter the transaction by Program IDs and/or accounts

				Supported fields:
				- program:<PROGRAM_ID> to filter by Program IDs
				- account:<ACCOUNT_ADDRESS> to filter by accounts involved in the transaction's instructions

				Supported operators are '||' and '&&' for logical operations and '()' for grouping.

				Examples
				  # Find any transaction containing instructions from the Compute Budget program
				  'program:ComputeBudget111111111111111111111111111111

				  # Find any transaction from the Token program involving a specific account
				  program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA && account:3MQw72oGrizUDEcD9gZYMgqo1pc364y5GnnJHcGpvurK
			`), "Submit").
			DefaultValue("program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA && account:3MQw72oGrizUDEcD9gZYMgqo1pc364y5GnnJHcGpvurK").
			Cmd()

	case InputFilter:
		c.State.Filter = msg.Value
		c.State.FilterContainsAccount = strings.Contains(c.State.Filter, "account:")
		return c.NextStep()
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

type AskFilter struct{}
type InputFilter struct{ pbconvo.UserInput_TextInput }
type ShowInstructions struct{}
