package solanchor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
		"solana",
	)
}

var cmd = codegen.Cmd
var IdlFilepathPrefix = "file://"

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if p.IdlFormat == "" {
		return cmd(AskIDLFormat{})
	}

	if p.idl == nil {
		switch p.IdlFormat {
		case "string":
			return cmd(AskIDLJSON{})
		case "file":
			return cmd(AskIDLFile{})
		}
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

			idl, err := createIDLFromJSON(c.State.IdlString)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				return loop.Quit(fmt.Errorf("could not decode IDL"))
			}
			c.State.idl = idl

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

	case AskIDLFormat:
		return c.Action(InputIDLFormat{}).
			ListSelect("How do you want to provide the JSON IDL?").
			Labels("JSON string", "JSON in a local file").
			Values("string", "file").
			DefaultValue("string").
			Cmd()

	case InputIDLFormat:
		c.State.IdlFormat = msg.Value
		return c.NextStep()

	case AskIDLJSON:
		return c.Action(InputIDLJSON{}).
			TextInput("Paste the Anchor IDL in JSON format\n", "Submit").
			Cmd()

	case InputIDLJSON:
		return inputIDLStep(c, msg.Value)

	case AskIDLFile:
		return c.Action(InputIDLFile{}).
			LocalFile("Input the full path of your JSON IDL in your filesystem (e.g. PATH_TO_MY_IDL/MY_IDL.json)\n", "Submit").
			Cmd()

	case InputIDLFile:
		return inputIDLStep(c, string(msg.Value))

	case AskConfirmIDL:
		return c.Action(InputConfirmIDL{}).
			Confirm("Do you want to proceed with this IDL?", "Yes", "No").
			DefaultAccept().
			Cmd()

	case InputConfirmIDL:
		returnStep := func() loop.Cmd {
			if c.State.IdlFormat == "string" {
				return cmd(AskIDLFile{})
			}
			return cmd(AskIDLJSON{})
		}

		if msg.Affirmative {
			return c.NextStep()
		}
		c.State.idl = nil
		c.State.IdlString = ""
		return loop.Seq(c.Msg().Message("Modify your JSON IDL and try again...").Cmd(), returnStep())

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

		c.State.idl.Metadata.Address = strings.TrimSpace(msg.Value)
		b, err := base58.Decode(msg.Value)
		if err != nil || len(b) != 32 {
			return loop.Seq(
				c.Msg().Message("This address is not a valid base58-encoded solana address").Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case codegen.RunGenerate:
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

func inputIDLStep(c *Convo, msgValue string) loop.Cmd {
	idl, err := createIDLFromJSON(msgValue)
	if err != nil {
		return loop.Quit(fmt.Errorf("could not decode IDL"))
	}
	if idl.Metadata.Name == "" {
		idl.Metadata.Name = c.State.Name // we need a name so anchor can compile
	}

	c.State.idl = idl
	c.State.IdlString = msgValue

	descString := descriptionFromIDL(idl)

	peekIDL := c.Msg().Message(descString).Cmd()
	return loop.Seq(peekIDL, cmd(AskConfirmIDL{}))
}

func createIDLFromJSON(text string) (*IDL, error) {
	idl := &IDL{}
	err := json.Unmarshal([]byte(text), &idl)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil, err
	}
	idl.MoveEventsIfNecessary()

	return idl, nil
}

func descriptionFromIDL(idl *IDL) string {
	descString := "# Instructions\n\n"
	for _, inst := range idl.Instructions {
		descString += fmt.Sprintf("## %s (%s)\n", inst.Name, arrayToHex(inst.Discriminator))
		if len(inst.Args) != 0 {
			argNames := make([]string, len(inst.Args))
			for i, field := range inst.Args {
				argNames[i] = field.Name
			}
			descString += fmt.Sprintf("* Args: (%s)\n", strings.Join(argNames, ", "))
		}
		if len(inst.Accounts) != 0 {
			accNames := make([]string, len(inst.Accounts))
			for i, field := range inst.Accounts {
				accNames[i] = field.Name
				if field.Address != "" {
					accNames[i] = "_" + field.Name + "_"
				}
			}
			descString += fmt.Sprintf("* Accounts: (%s)\n", strings.Join(accNames, ", "))
		}
		descString += "\n"
	}

	if len(idl.Events) != 0 {
		descString += fmt.Sprintf("# %s\n", "Events")
		for _, evt := range idl.Events {
			fieldNames := make([]string, len(evt.Fields))
			for i, field := range evt.Fields {
				fieldNames[i] = field.Name
			}
			descString += fmt.Sprintf("* %s (%s)\n", evt.Name, strings.Join(fieldNames, ", "))
			descString += "\n"
		}
	}

	return descString
}
