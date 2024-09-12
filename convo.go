package codegen

import (
	"github.com/streamingfast/substreams-codegen/loop"
)

type Conversation[X any] struct {
	State X

	factory    *MsgWrapFactory
	updateFunc func(Conversation[X], loop.Msg) loop.Cmd
}

func (c *Conversation[X]) SetFactory(f *MsgWrapFactory) {
	c.factory = f
}

func (c *Conversation[X]) Msg() *MsgWrap { return c.factory.NewMsg(c.State) }

func (c *Conversation[X]) Action(element any) *MsgWrap {
	return c.factory.NewInput(element, c.State)
}
