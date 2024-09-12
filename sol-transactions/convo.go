package soltransactions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
		"sol-transactions",
		"Get Solana transactions filtered by one or several Program IDs",
		"Allows you to specified a regex containing the Program IDs used to filter the Solana transactions",
		codegen.ConversationFactory(New),
		100,
	)
}

var cmd = codegen.Cmd

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
	}

	if p.Filter == "" {
		return cmd(AskFilter{})
	}

	return cmd(codegen.RunGenerate{})
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
		return c.CmdAskProjectName()

	case codegen.InputProjectName:
		c.State.Name = msg.Value
		return c.NextStep()

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
			TextInput(fmt.Sprintf("Filter the transaction by Program IDs and/or accounts.\nSupported operators are: logical or '||', logical and '&&' and parenthesis: '()'. \n\nEXAMPLE: to only consume TRANSACTIONS containing:\n   - ComputeBudget instructions\n        OR\n   - Token Instructions where the account '3MQw72oGrizUDEcD9gZYMgqo1pc364y5GnnJHcGpvurK' is included\n'program:ComputeBudget111111111111111111111111111111 || (program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA && account:3MQw72oGrizUDEcD9gZYMgqo1pc364y5GnnJHcGpvurK)'\n"), "Submit").
			DefaultValue("program:ComputeBudget111111111111111111111111111111 || (program:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA && account:3MQw72oGrizUDEcD9gZYMgqo1pc364y5GnnJHcGpvurK)").
			Cmd()

	case InputFilter:
		c.State.Filter = msg.Value
		c.State.FilterContainsAccount = strings.Contains(c.State.Filter, "account:")
		return c.NextStep()

	case codegen.RunGenerate:
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
