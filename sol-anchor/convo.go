package solanchor

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mr-tron/base58"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"github.com/tidwall/sjson"
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
		"sol-anchor-beta",
		"Given an Anchor JSON IDL, create a Substreams that decodes instructions and events",
		"Allows you to decode data based on an Anchor JSON IDL",
		codegen.ConversationFactory(New),
		2000,
	)
}

var cmd = codegen.Cmd

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if p.idl == nil {
		return cmd(AskIdl{})
	}

	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if p.idl.ProgramID() == "" {
		return cmd(AskProgramID{})
	}

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
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

	case codegen.AskChainName:
		labels := []string{"Solana Mainnet", "Solana Devnet"}
		values := []string{"solana-mainnet", "solana-devnet"}
		return c.Action(codegen.InputChainName{}).ListSelect("Please select the chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
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
			TextInput("Input the Anchor IDL in JSON format\n", "Submit").
			Cmd()

	case InputIdl:
		idl := &IDL{}
		err := json.Unmarshal([]byte(msg.Value), &idl)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return loop.Quit(fmt.Errorf("could not decode IDL"))
		}
		if idl.Metadata.Name == "" {
			idl.Metadata.Name = c.State.Name // we need a name so anchor can compile
		}
		c.State.idl = idl
		c.State.IdlString = msg.Value

		descString := "Here are the instructions that you will get from this IDL:\n\n"
		for _, inst := range idl.Instructions {
			descString += fmt.Sprintf("# %s (%s)\n", inst.Name, arrayToHex(inst.Discriminator))
			if len(inst.Args) == 0 {
				descString += "## Args (none)\n"
			} else {
				descString += "## Args\n"
				for _, arg := range inst.Args {
					descString += fmt.Sprintf("* %s\n", arg.Name)
				}
			}
			if len(inst.Accounts) == 0 {
				descString += "\n## Accounts (none)\n"
			} else {
				descString += "\n## Accounts\n"
				for _, acc := range inst.Accounts {
					if acc.Address == "" {
						descString += fmt.Sprintf("- %s\n", acc.Name)
					} else {
						descString += fmt.Sprintf("- _%s (ignored static: %s)_\n", acc.Name, acc.Address)
					}
				}
			}
			descString += "\n"
		}

		peekIDL := c.Msg().Message(descString).Cmd()
		return loop.Seq(peekIDL, cmd(AskConfirmIDL{}))

	case AskConfirmIDL:
		return c.Action(InputConfirmIDL{}).
			Confirm("Do you want to proceed with this IDL?", "Yes", "No").
			DefaultAccept().
			Cmd()

	case InputConfirmIDL:
		if msg.Affirmative {
			return c.NextStep()
		}
		c.State.idl = nil
		c.State.IdlString = ""
		return loop.Seq(c.Msg().Message("Modify your JSON IDL and try again...").Cmd(), cmd(AskIdl{}))

	case AskProgramID:
		return c.Action(InputProgramID{}).
			TextInput("Cannot get the ProgramID from the IDL. Please input the Program ID to match.\n", "Submit").
			Cmd()

	case InputProgramID:
		newIDLString, err := sjson.Set(c.State.IdlString, "metadata.address", msg.Value)
		if err != nil {
			return loop.Quit(fmt.Errorf("could not set ProgramID in IDL: %w", err))
		}
		c.State.IdlString = newIDLString

		c.State.idl.Metadata.Address = msg.Value
		b, err := base58.Decode(msg.Value)
		if err != nil || len(b) != 32 {
			return loop.Seq(
				c.Msg().Message("This address is not a valid base58-encoded solana address").Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case codegen.RunGenerate:
		str := ""
		for _, event := range c.State.idl.Events {
			for _, field := range event.Fields {
				fmt.Println("-----------------------------")
				fmt.Println(field.Name)

				str += field.Type.Simple
			}
		}
		return c.CmdGenerate(c.State.Generate)

	case codegen.ReturnGenerate:
		return c.CmdDownloadFiles(msg)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

func arrayToHex(arr []uint8) (out string) {
	for i, v := range arr {
		if i == 0 {
			out = fmt.Sprintf("%02x", v)
			continue
		}
		out = fmt.Sprintf("%s %02x", out, v)
	}
	return out
}
