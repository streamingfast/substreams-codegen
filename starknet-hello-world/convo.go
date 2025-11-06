package starknethelloworld

import (
	"fmt"
	"regexp"

	networks "github.com/streamingfast/firehose-networks"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

var starknetNetworkRegexp = regexp.MustCompile(`^starknet`)

var sharedFlowConfig = codegen.SharedFlowConfig{
	ValidChains: networks.GetSubstreamsRegistry().Search(starknetNetworkRegexp),
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
		"starknet-hello-world",
		"Creates a Substreams that indexes Starknet events from the Starknet Token Contract",
		"You will get a very simple project to get started with Substreams.",
		New,
		59,
		"Starknet",
	)
}

func (c *Convo) NextStep() loop.Cmd {
	if !c.IsPreSharedFlowDone(sharedFlowConfig) {
		return c.NextPreSharedFlowStep(sharedFlowConfig)
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

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
