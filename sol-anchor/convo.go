package solanchor

import (
	"encoding/json"
	"fmt"
	"strconv"

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
		"sol-anchor",
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

	if p.Idl == nil {
		return cmd(AskIdl{})
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

	case AskIdl:

		return c.Action(InputIdl{}).
			TextInput(fmt.Sprintf("Input the Anchor IDL in JSON format\n"), "Submit").
			Cmd()

	case InputIdl:
		idl := &IDL{}
		err := json.Unmarshal([]byte(msg.Value), &idl)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return loop.Quit(fmt.Errorf("could not decode IDL"))
		}

		c.State.Idl = idl
		c.State.IdlString = msg.Value
		return c.NextStep()

	case codegen.RunGenerate:
		str := ""
		for _, event := range c.State.Idl.Events {
			for _, field := range event.Fields {
				fmt.Println("-----------------------------")
				fmt.Println(field.Name)

				str += fmt.Sprintf("%s", field.Type.Simple)
			}
		}
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
