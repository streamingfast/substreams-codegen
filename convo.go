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

func (c *Conversation[X]) GetState() any {
	return c.State
}

func (c *Conversation[X]) Msg() *MsgWrap { return c.factory.NewMsg(c.State) }

func (c *Conversation[X]) Action(element any) *MsgWrap {
	return c.factory.NewInput(element, c.State)
}

func (c *Conversation[X]) CmdGenerate(f func() ReturnGenerate) loop.Cmd {
	return loop.Seq(
		c.Msg().Message("Generating Substreams module source code...").Cmd(),
		func() loop.Msg {
			return f()
		},
	)
}
