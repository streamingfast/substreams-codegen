package injectivehelloworld

import (
	"fmt"
	"regexp"
	"strconv"

	networks "github.com/streamingfast/firehose-networks"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var injectiveNetworkRegexp = regexp.MustCompile(`^injective-(testnet|mainnet)`)

var sharedFlowConfig = codegen.SharedFlowConfig{
	ValidChains: networks.GetSubstreamsRegistry().Search(injectiveNetworkRegexp),
}

var InjectiveTestnetDefaultStartBlock uint64 = 37368800

type Convo struct {
	*codegen.Conversation[*Project]
}

func init() {
	codegen.RegisterConversation(
		"injective-hello-world",
		"Creates a Substreams that extracts 'transfer' Injective events from blocks.",
		"You will get a very simple project to get started with Substreams.",
		New,
		72,
		"Cosmos",
	)
}

func New() codegen.Converser {
	return &Convo{&codegen.Conversation[*Project]{
		State: &Project{},
	}}
}

func (c *Convo) NextStep() loop.Cmd {
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
	}

	p := c.State

	if !p.InitialBlockSet {
		return cmd(codegen.AskInitialStartBlockType{})
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
		textInputMessage := "At what block do you want to start indexing data?"
		defaultValue := "0"
		if c.State.IsChainTestnet() {
			defaultValue = fmt.Sprintf("%d", InjectiveTestnetDefaultStartBlock)
			textInputMessage = fmt.Sprintf("At what block do you want to start indexing data? (the first available block on %s is: %s)", c.State.ChainName, defaultValue)
		}
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
		if c.State.IsChainTestnet() && initialBlock < InjectiveTestnetDefaultStartBlock {
			initialBlock = InjectiveTestnetDefaultStartBlock
		}

		c.State.InitialBlock = initialBlock
		c.State.InitialBlockSet = true
		return c.NextStep()
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}

var cmd = codegen.Cmd
