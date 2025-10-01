package codegen

import (
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type SendFunc func(msg *pbconvo.SystemOutput, err error)

type ConversationFactory func() Converser

type Converser interface {
	SetClientVersion(uint32)
	NextStep() loop.Cmd
	Update(loop.Msg) loop.Cmd
	SetFactory(f *MsgWrapFactory)

	// GetState returns the current conversation state as an interface. This cannot
	// be generically typed here other when creating a registry of conversations,
	// it's impossible to cast from the specific type X to ConversationState, Go
	// doesn't allow that even though every Converser[X] is also Converser[ConversationState].
	GetState() ConversationState
}
