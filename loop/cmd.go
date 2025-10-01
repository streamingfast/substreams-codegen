package loop

import (
	"time"
)

type Cmds []Cmd

func (c Cmds) Then(cmd Cmd) Cmds {
	return append(c, cmd)
}

func (c Cmds) Seq() Cmd {
	return Seq(c...)
}

func (c Cmds) Batch() Cmd {
	return Batch(c...)
}

type Msg any

type Cmd func() Msg

type BatchMsg []Cmd

func Batch(cmds ...Cmd) Cmd {
	var validCmds []Cmd
	for _, c := range cmds {
		if c == nil {
			continue
		}
		validCmds = append(validCmds, c)
	}
	if len(validCmds) == 0 {
		return nil
	}
	return func() Msg {
		return BatchMsg(validCmds)
	}
}

type SeqMsg []Cmd

func Seq(cmds ...Cmd) Cmd {
	return func() Msg {
		return SeqMsg(cmds)
	}
}

// SeqAnys is like Seq but accepts any type and ignores nils.
func SeqAnys(inputs ...any) Cmd {
	var cmds []Cmd
	if len(inputs) > 0 {
		cmds = make([]Cmd, 0, len(inputs))
		for _, in := range inputs {
			if in != nil {
				cmds = append(cmds, func() Msg { return in })
			}
		}
	}

	return Seq(cmds...)
}

func NewQuitMsg(err error) Msg {
	return QuitMsg{err}
}

type QuitMsg struct {
	err error
}

func Quit(err error) Cmd {
	return func() Msg {
		return QuitMsg{err}
	}
}

func Tick(delay time.Duration, fn func() Msg) Cmd {
	return func() Msg {
		time.Sleep(delay)
		return fn()
	}
}
