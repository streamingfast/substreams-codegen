package codegen

import (
	"github.com/streamingfast/substreams-codegen/loop"
	pbconvo "github.com/streamingfast/substreams-codegen/pb/sf/codegen/conversation/v1"
)

type SendFunc func(msg *pbconvo.SystemOutput, err error)

type ConversationFactory func(*MsgWrapFactory) Conversation

type Conversation interface {
	Update(loop.Msg) loop.Cmd
}
