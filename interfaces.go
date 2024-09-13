package codegen

import (
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type SendFunc func(msg *pbconvo.SystemOutput, err error)

type ConversationFactory func() Converser

type Converser interface {
	// Functions provided by the Conversation instance

	NextStep() loop.Cmd
	Update(loop.Msg) loop.Cmd

	// Functions provided by the *Conversation type

	SetFactory(f *MsgWrapFactory)
	GetState() any
}
