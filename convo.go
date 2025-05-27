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

func (c *Conversation[X]) HandleSubstreamsConsumptionChoice(value string) loop.Cmd {

	var sinkMessage *MsgWrap
	switch value {
	case "sql":
		sinkMessage = c.Msg().Message("Sink to SQL: See https://docs.substreams.dev/how-to-guides/sinks/sql-sink")
	case "csv":
		sinkMessage = c.Msg().Message("Sink to CSV file: See https://docs.substreams.dev/how-to-guides/sinks/community-sinks/files")
	case "json":
		sinkMessage = c.Msg().Message("Sink to JSON file: See https://docs.substreams.dev/how-to-guides/sinks/community-sinks/files")
	case "parquet":
		sinkMessage = c.Msg().Message("Sink to Parquet file: See https://docs.substreams.dev/how-to-guides/sinks/community-sinks/files")
	case "golang":
		sinkMessage = c.Msg().Message("Stream using Go: https://docs.substreams.dev/how-to-guides/sinks/stream/go")
	case "rust":
		sinkMessage = c.Msg().Message("Stream using Rust: https://github.com/streamingfast/substreams-sink-examples/tree/master/rust#readme")
	case "javascript":
		sinkMessage = c.Msg().Message("Stream using JavaScript: https://docs.substreams.dev/how-to-guides/sinks/stream/javascript")
	case "pubsub":
		sinkMessage = c.Msg().Message("Stream using Pub/Sub: https://docs.substreams.dev/how-to-guides/sinks/pubsub")
	default:
		sinkMessage = c.Msg().Message("Invalid choice")
	}

	return loop.Seq(
		sinkMessage.Cmd(),
		loop.Quit(nil),
	)
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
	values := []string{"sql", "csv", "json", "parquet", "golang", "rust", "javascript", "pubsub"}
	labels := []string{"To SQL", "To CSV Files", "To JSON Files", "To Parquet Files", "Stream using Golang", "Stream using Rust", "Stream using JavaScript/TypeScript", "Stream to Pub/Sub"}

	act := c.Action(InputSubstreamsConsumptionChoice{}).ListSelect("How would you like to consume the Substreams?", "consumption").
		Labels(labels...).
		Values(values...)

	return loop.Seq(
		downloadCmd.Cmd(),
		c.Msg().Messagef(`Your Substreams project is ready! Start streaming with:

`+"```"+`bash
substreams build
substreams auth
substreams gui       			  # Get streaming!
`+"```"+`

Build Subgraphs and other sinks with:

`+"```"+`bash
substreams codegen subgraph
substreams codegen sql
`+"```"+`

Optionally, publish your Substreams to the Substreams Registry (https://substreams.dev) with:

`+"```"+`bash
substreams registry login         # Login to substreams.dev
substreams registry publish       # Publish your Substreams to substreams.dev
`+"```"+`

`).Cmd(),
		act.Cmd(),
	)
}
