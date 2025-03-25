package stellartransactionsoperations

import (
	"encoding/json"
	"errors"
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
		"stellar-transactions-operations",
		"Creates a Substreams project which filtering transactions or operations.",
		"You will get a project that indexes transactions or operations by providing a filter.",
		codegen.ConversationFactory(New),
		59,
		"stellar",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return cmd(codegen.AskProjectName{})
	}

	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if !p.IsValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if p.FilterType == "" {
		return cmd(AskFilterType{})
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
	case AskFilterType:
		return c.Action(InputFilterType{}).ListSelect("What kind of data do you want to index?\n\n- Raw transactions: you can filter the transactions based on source account at the transactions and/or operation level.\n- Operations: you can get operatios filtered by operation name\n\n").
			Labels("Raw transactions (filtered by source account(s))", "Operations (filtered by operation name)").
			Values("transactions", "operations").
			Cmd()
	case InputFilterType:
		c.State.FilterType = msg.Value

		return c.NextStep()

	case AskFilter:
		message := "Input the source account(s) that you want to use to filter separated by commas (,). For example:\nGADLRTGF4GCU2CNAHYPKAEBGQSBX2M3UYIZZJODZAVC5A5QCAE7AT66C,GADLWELLJ56NXB76MQGXRXSRCFT5YY2ANWXPKWVY7YP6EJIYYDFKL43W\n"
		if c.State.FilterType == "operations" {
			message = "Input the operation names that you want to use to filter separated by commas (,) For eample:\npayment,create_account\n"
		}

		return c.Action(InputFilter{}).
			TextInput(message, "Submit").
			Cmd()

	case InputFilter:
		if isValid := isFilterCorrect(msg.Value); isValid == false {
			return loop.Seq(cmd(InvalidFilter{errors.New(fmt.Sprintf("ERROR: The specified filter does not have a correct format: %s", msg.Value))}), cmd(AskFilter{}))
		}

		c.State.Filter = msg.Value

		return c.NextStep()

	case InvalidFilter:
		return c.Msg().
			Messagef("%s", msg.Err).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.Msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.State.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
		if c.State.IsValidChainName(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainConfig().DisplayName).Cmd(),
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

var cmd = codegen.Cmd
