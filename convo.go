package codegen

import (
	"fmt"

	"github.com/streamingfast/logging"
	"github.com/streamingfast/substreams-codegen/loop"
)

var zlog, tracer = logging.PackageLogger("convo", "github.com/streamingfast/substreams-codegen/codegen")

// ConversationState defines the minimal interface that a conversation state
// must implement to be used within the generic Conversation[X ConversationState] struct.
//
// Embedding [BaseConversationState] is recommended to get default implementations
// of the methods.
type ConversationState interface {
	// GetChainName should return the Network Registry chain's id ideally (or alias)
	// representing the chain the user selected. It expected to be "" before the actual
	// user made a decision
	GetChainName() string

	// GetChainDisplayName should return the Network Registry chain's full name
	// representing the chain the user selected. It expected to be "" before the actual
	// user made a decision
	GetChainDisplayName() string

	// SetChainName should set the chain name in the state.
	SetChainName(name string)

	// GetModuleName should return the module name to be used in the generated substreams.yaml
	// that the user should call in substreams run/gui/sink commands.
	GetModuleName() string

	// SetProjectName should set the project name in the state, this also derives the module name
	// so populated [GetModuleName] will return a non-empty value after this method is called.
	SetProjectName(name string)

	// GetSubstreamsSinkChoice should return the user's choice of Substreams sink.
	GetSubstreamsSinkChoice() *SubstreamsSinkChoice

	// SetSubstreamsSinkChoice should set the user's choice of Substreams sink.
	SetSubstreamsSinkChoice(choice SubstreamsSinkChoice)

	// Generate the ReturnGenerate containing the generated files or error.
	Generate() ReturnGenerate
}

type ValidatingConversationState interface {
	ConversationState
	Validate() error
}

type Conversation[X ConversationState] struct {
	// State holds the current conversation state, check [BaseConversationState] for common fields
	// every state should embedded and fullfil the interface [ConversationState].
	State X

	clientVersion uint32
	factory       *MsgWrapFactory
}

func (c *Conversation[X]) SetFactory(f *MsgWrapFactory) {
	c.factory = f
}

func (c *Conversation[X]) SetClientVersion(version uint32) {
	c.clientVersion = version
}

func (c *Conversation[X]) GetState() ConversationState {
	return c.State
}

func (c *Conversation[X]) Msg() *MsgWrap { return c.factory.NewMsg(c.State) }

func (c *Conversation[X]) Action(element any) *MsgWrap {
	return c.factory.NewInput(element, c.State)
}

func NewBaseBlockchainGeneratorConversation[X ConversationState](state X, config SharedFlowConfig) *BaseBlockchainGeneratorConversation[X] {
	return &BaseBlockchainGeneratorConversation[X]{
		Conversation:     Conversation[X]{State: state},
		sharedFlowConfig: config,
	}
}

type BaseBlockchainGeneratorConversation[X ConversationState] struct {
	Conversation[X]

	sharedFlowConfig SharedFlowConfig
}

func (c *BaseBlockchainGeneratorConversation[X]) NextStep() loop.Cmd {
	if !c.IsPreSharedFlowDone(c.sharedFlowConfig) {
		return c.NextPreSharedFlowStep(c.sharedFlowConfig)
	}

	return c.NextPostSharedFlowStep(c.sharedFlowConfig)
}

func (c *BaseBlockchainGeneratorConversation[X]) Update(msg loop.Msg) loop.Cmd {
	if c.IsPreSharedFlowMsg(msg, c.sharedFlowConfig) {
		return c.UpdatePreSharedFlowMsg(msg, c.sharedFlowConfig, c.NextStep)
	}

	if c.IsPostSharedFlowMsg(msg, c.sharedFlowConfig) {
		return c.UpdatePostSharedFlowMsg(msg, c.sharedFlowConfig)
	}

	return loop.Quit(fmt.Errorf("invalid loop message: %T", msg))
}
