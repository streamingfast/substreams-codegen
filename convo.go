package codegen

import (
	"maps"
	"slices"

	"github.com/streamingfast/substreams-codegen/loop"
)

type Conversation[X any] struct {
	State X

	factory *MsgWrapFactory
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

func (c *Conversation[X]) CmdAskProjectName() loop.Cmd {
	return c.Action(InputProjectName{}).
		TextInput("Please enter the project name", "Submit").
		Description("Identifier with only lowercase letters, numbers and underscores, up to 64 characters.").
		DefaultValue("my_project").
		Validation("^([a-z][a-z0-9_]{0,63})$", "The project name must be a valid identifier with only lowercase letters, numbers and underscores, up to 64 characters.").
		Cmd()
}

func (c *Conversation[X]) CmdDownloadFiles(msg ReturnGenerate) loop.Cmd {
	if msg.Err != nil {
		return loop.Seq(
			c.Msg().Messagef("Code generation failed with error: %s", msg.Err).Cmd(),
			loop.Quit(msg.Err),
		)
	}

	downloadCmd := c.Action(InputSourceDownloaded{}).DownloadFiles()

	for _, fileName := range slices.Sorted(maps.Keys(msg.ProjectFiles)) {
		fileDescription := ""
		if _, ok := FileDescriptions[fileName]; ok {
			fileDescription = FileDescriptions[fileName]
		}
		downloadCmd.AddFile(fileName, msg.ProjectFiles[fileName], "text/plain", fileDescription)
	}

	return loop.Seq(
		c.Msg().Messagef(`Your Substreams project is ready! Start streaming with:

`+"```"+`bash
substreams build
substreams auth
substreams gui       # Get streaming!
`+"```"+`

Build Subgraphs and other sinks with:

`+"```"+`bash
substreams codegen subgraph
substreams codegen sql
`+"```"+`
`).Cmd(), downloadCmd.Cmd(),
		loop.Quit(nil),
	)
}
