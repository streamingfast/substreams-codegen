package codegen

import (
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type SendFunc func(msg *pbconvo.SystemOutput, err error)

type ConversationFactory func() Converser

type Converser interface {
	NextStep() loop.Cmd
	Update(loop.Msg) loop.Cmd

	SetFactory(f *MsgWrapFactory)
}
