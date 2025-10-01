package stellartransactionsoperations

import (
	"encoding/json"
	"fmt"
	"regexp"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
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
		New,
		59,
		"Stellar",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	p := c.State
	if p.Name == "" {
		return codegen.Cmd(codegen.AskProjectName{})
	}

	if p.ChainName == "" {
		return codegen.Cmd(codegen.AskChainName{})
	}

	if !networks.GetSubstreamsRegistry().Has(p.ChainName) {
		return loop.SeqAnys(codegen.MsgInvalidChainName{}, codegen.AskChainName{})
	}

	if p.FilterType == "" {
		return codegen.Cmd(AskFilterType{})
	}

	if p.Filter == "" {
		return codegen.Cmd(AskFilter{})
	}

	return codegen.Cmd(codegen.RunGenerate{})
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

	case codegen.AskProjectName:
		return c.CmdAskProjectName()

	case codegen.InputProjectName:
		c.State.Name = msg.Value
		return c.NextStep()

	case codegen.AskChainName:
		var labels, values []string
		for _, network := range stellarNetworks() {
			labels = append(labels, network.FullName)
			values = append(values, network.ID)
		}
		return c.Action(codegen.InputChainName{}).ListSelect("Please select the chain", "chain").
			Labels(labels...).
			Values(values...).
			Cmd()
	case AskFilterType:
		return c.Action(InputFilterType{}).ListSelect("What kind of data do you want to index?\n\n- Raw transactions: you can filter the transactions based on source account at the transactions and/or operation level.\n- Operations: you can get operatios filtered by operation name\n\n", "data_type").
			Labels("Raw transactions (filtered by source account(s))", "Operations (filtered by operation name)").
			Values("transactions", "operations").
			Cmd()
	case InputFilterType:
		c.State.FilterType = msg.Value

		return c.NextStep()

	case AskFilter:
		message := "Input the source account(s) that you want to use to filter separated by commas (,). For example:\nGADLRTGF4GCU2CNAHYPKAEBGQSBX2M3UYIZZJODZAVC5A5QCAE7AT66C,GADLWELLJ56NXB76MQGXRXSRCFT5YY2ANWXPKWVY7YP6EJIYYDFKL43W\n"
		if c.State.FilterType == "operations" {
			message = "Input the operation names that you want to use to filter separated by commas (,) For example:\npayment,create_account\n"
		}

		return c.Action(InputFilter{}).
			TextInput(message, "Submit").
			Cmd()

	case InputFilter:
		if !isFilterCorrect(msg.Value) {
			return loop.SeqAnys(InvalidFilter{fmt.Errorf("ERROR: The specified filter does not have a correct format: %s", msg.Value)}, AskFilter{})
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

	case codegen.InputSubstreamsConsumptionChoice:
		return c.HandleSubstreamsConsumptionChoice(msg.Value)

	case codegen.InputSourceDownloaded:
		return c.HandleDownloaded(msg.Value)

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
		if networks.GetSubstreamsRegistry().Has(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainDisplayName()).Cmd(),
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

type AskFilterType struct{}
type InputFilterType struct{ pbconvo.UserInput_Selection }

type AskFilter struct{}
type InvalidFilter struct{ Err error }
type InputFilter struct{ pbconvo.UserInput_TextInput }

// Regular expression: Allows letters, numbers, and underscores, separated by commas
var filterRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+(,[a-zA-Z0-9_]+)*$`)

func isFilterCorrect(s string) bool {
	return filterRegexp.MatchString(s)
}

var stellarNetworkRegexp = regexp.MustCompile(`^stellar`)

func stellarNetworks() []*registry.Network {
	return networks.GetSubstreamsRegistry().Search(stellarNetworkRegexp)
}
