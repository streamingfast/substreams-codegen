package ethhelloworld

import (
	"fmt"
	"strconv"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var sharedFlowConfig = codegen.SharedFlowConfig{
	// For now, we ask in this specific conversation for the chain's name, to be refactored at some point
	ValidChains: nil,
}

func init() {
	supportedChains := make([]string, 0, len(ChainConfigs))
	for _, conf := range ChainConfigs {
		supportedChains = append(supportedChains, conf.DisplayName)
	}
	codegen.RegisterConversation(
		"evm-hello-world",
		"Creates a Substreams that extracts USDC log data from blocks",
		`Supported networks: `+strings.Join(supportedChains, ", "),
		codegen.ConversationFactory(New),
		83,
		"EVM",
	)
}

type Convo struct {
	*codegen.Conversation[*Project]
}

func New() codegen.Converser {
	c := &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}

	fmt.Println("Get state", c.GetState().GetChainName())

	return c
}

func (c *Convo) NextStep() (out loop.Cmd) {
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
	}

	p := c.State

	if p.ChainName == "" {
		return cmd(codegen.AskChainName{})
	}

	if !isValidChainName(p.ChainName) {
		return loop.Seq(cmd(codegen.MsgInvalidChainName{}), cmd(codegen.AskChainName{}))
	}

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
	}

	return c.NextPostSharedFlowStep(sharedFlowConfig)
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
}

func (c *Convo) Update(msg loop.Msg) loop.Cmd {
	if c.IsPreSharedFlowMsg(msg, sharedFlowConfig) {
		return c.UpdatePreSharedFlowMsg(msg, sharedFlowConfig, c.NextStep)
	}

	if c.IsPostSharedFlowMsg(msg, sharedFlowConfig) {
		return c.UpdatePostSharedFlowMsg(msg, sharedFlowConfig)
	}

	switch msg := msg.(type) {
	case codegen.AskChainName:
		var labels, values []string
		for _, conf := range ChainConfigs {
			labels = append(labels, conf.DisplayName)
			values = append(values, conf.ID)
		}
		return c.Action(codegen.InputChainName{}).ListSelect("Please select the chain", "chain").
			Labels(labels...).
			Values(values...).
			Cmd()

	case codegen.MsgInvalidChainName:
		return c.Msg().
			Messagef(`Hmm, %q seems like an invalid chain name. Maybe it was supported and is not anymore?`, c.State.ChainName).
			Cmd()

	case codegen.InputChainName:
		c.State.ChainName = msg.Value
		if isValidChainName(msg.Value) {
			return loop.Seq(
				c.Msg().Messagef("Got it, will be using chain %q", c.State.ChainConfig().DisplayName).Cmd(),
				c.NextStep(),
			)
		}
		return c.NextStep()

	case codegen.AskInitialStartBlockType:
		textInputMessage := "At what block do you want to start indexing data?"
		defaultValue := "0"
		return c.Action(codegen.InputAskInitialStartBlockType{}).
			TextInput(textInputMessage, "Submit").
			DefaultValue(defaultValue).
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
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd
