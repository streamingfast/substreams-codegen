package stellartransactionsoperations

import (
	"fmt"
	"regexp"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

var sharedFlowConfig = codegen.SharedFlowConfig{
	ValidChains: chains.StellarNetworks(),
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
		"stellar-transactions-operations",
		"Creates a Substreams project which filtering transactions or operations.",
		"You will get a project that indexes transactions or operations by providing a filter.",
		New,
		59,
		"Stellar",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
	}

	p := c.State

	if p.FilterType == "" {
		return codegen.Cmd(AskFilterType{})
	}

	if p.Filter == "" {
		return codegen.Cmd(AskFilter{})
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
	case AskFilterType:
		return c.Action(InputFilterType{}).ListSelect("What kind of data do you want to index?\n\n- Raw transactions: you can filter the transactions based on source account at the transactions and/or operation level.\n- Operations: you can get operatios filtered by operation name\n\n", "data_type").
			Labels("Raw transactions (filtered by source account(s))", "Operations (filtered by operation name)").
			Values("transactions", "operations").
			Cmd()
	case InputFilterType:
		c.State.FilterType = msg.Value

		return c.NextStep()

	case AskFilter:
		message := "Input the source account(s) that you want to use to filter separated by commas (,). For example:\nGADLRTGF4GCU2CNAHYPKAEBGQSBX2M3UYIZZJODZAVC5A5QCAE7AT66C,GADLWELLJ56NXB76MQGXRXSRCFT5YY2ANWXPKWVY7YP6EJIYYDFKL43W,GBWRR4M6WPOV4Y5TGLCXS2NAL76IAT4V4QSCX7CKPH4ILUJEDTQFEZM5\n"
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
