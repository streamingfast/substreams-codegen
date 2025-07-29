package codegen

import (
	"maps"
	"slices"

	"github.com/streamingfast/substreams-codegen/loop"
)

type Conversation[X any] struct {
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

func (c *Conversation[X]) GetState() any {
	return c.State
}

// GetNetworkAndModule attempts to extract network and output module information from the conversation state
func (c *Conversation[X]) GetNetworkAndModule() (network string, outputModule string, hasInfo bool) {
	state := c.GetState()
	
	// Try to extract information using reflection/type assertion
	// This handles the common case where state has ChainName and Name fields
	if stateMap, ok := state.(interface {
		GetChainName() string
		ModuleName() string
		TrackAnyEvents() bool
		TrackAnyCalls() bool
	}); ok {
		network = stateMap.GetChainName()
		
		// Determine output module based on what's being tracked
		if stateMap.TrackAnyEvents() && stateMap.TrackAnyCalls() {
			outputModule = "map_events_calls"
		} else if stateMap.TrackAnyEvents() {
			outputModule = "map_events"
		} else if stateMap.TrackAnyCalls() {
			outputModule = "map_calls"
		} else {
			// Default to events if we can't determine
			outputModule = "map_events"
		}
		
		return network, outputModule, true
	}
	
	// Fallback: try to access fields by name using reflection
	// This is a more generic approach for states that might have these fields
	// but don't implement the interface above
	
	return "", "", false
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
	// Try to get actual network and module information
	network, outputModule, hasInfo := c.GetNetworkAndModule()
	
	// Prepare the values to use in commands
	var endpoint, module string
	if hasInfo {
		endpoint = NetworkToEndpoint(network)
		module = outputModule
	} else {
		// Fallback to placeholders if we can't determine the values
		endpoint = "{endpoint}"
		module = "{output_module}"
	}
	
	var sinkMessage *MsgWrap
	switch value {
	case "sql":
		sinkMessage = c.Msg().Message(`Sink to SQL:
		1. Get the binary from https://github.com/streamingfast/substreams-sink-sql/ (version 4.6.1 or above)
		2. Run ` + "`substreams-sink-sql from-proto psql://db_user:db_password@db_host:5432/db_name ./substreams.yaml " + module + "`" +
			` See https://docs.substreams.dev/how-to-guides/sinks/sql-sink"`)
	case "parquet":
		sinkMessage = c.Msg().Message(`Sink to Parquet file:
			1. Get the binary from https://github.com/streamingfast/substreams-sink-files/ (version 2.1.0 or above)
			2. Run ` + "`substreams-sink-files run " + endpoint + " substreams.yaml " + module + " ./output`" +
			` See https://github.com/streamingfast/substreams-sink-files?tab=readme-ov-file#parquet"`)
	case "golang":
		sinkMessage = c.Msg().Message(`Sink using Golang

    		We provide a Substreams Golang SDK to streamline consumption of Substreams data
    		refer to https://github.com/streamingfast/substreams-sink for more details and examples.`)
	case "rust":
		sinkMessage = c.Msg().Message(`Sink using Rust

			Here is an example of a Rust sink: https://github.com/streamingfast/substreams-sink-examples/tree/master/rust#readme`)
	case "javascript":
		sinkMessage = c.Msg().Message(`Sink using Javascript

		Here is an example of a JS sink: https://github.com/streamingfast/substreams-sink-examples/blob/master/javascript/README.md`)

	case "python":
		sinkMessage = c.Msg().Message(`Sink using Python

			Here is an example of a Python sink: https://github.com/streamingfast/substreams-sink-examples/blob/master/python/README.md`)

	//case "pubsub":
	//	sinkMessage = c.Msg().Message("Stream using Pub/Sub: (Not implemented yet)")
	//case "json":
	//	sinkMessage = c.Msg().Message("Sink to JSON file: (Not implemented yet)")
	//case "csv":
	//	sinkMessage = c.Msg().Message("Sink to CSV file: (Not implemented yet)")
	default:
		sinkMessage = c.Msg().Message("Invalid choice")
	}

	return loop.Seq(
		sinkMessage.Cmd(),
		loop.Quit(nil),
	)
}

func (c *Conversation[X]) downloadedCommands(destDir string) []loop.Cmd {

	values := []string{
		"sql",
		//"csv",
		//"json",
		"parquet",
		"golang",
		"rust",
		"javascript",
		"python",
		//"pubsub",
	}
	labels := []string{
		"To SQL",
		//"To CSV Files",
		//"To JSON Files",
		"To Parquet Files",
		"Write a custom sink in Go",
		"Write a custom sink in Rust",
		"Write a custom sink in JavaScript/TypeScript",
		"Write a custom sink in Python",
		//"Stream to Pub/Sub",
	}

	act := c.Action(InputSubstreamsConsumptionChoice{}).ListSelect("How would you like to consume the Substreams?", "consumption").
		Labels(labels...).
		Values(values...)

	return []loop.Cmd{
		c.Msg().Messagef(`Your Substreams project is ready! Start streaming with:

`+"```"+`bash

cd %s
substreams build
substreams auth
substreams gui       			  # Get streaming!
`+"```"+`

Optionally, publish your Substreams to the Substreams Registry (https://substreams.dev) with:

`+"```"+`bash
substreams registry login         # Login to substreams.dev
substreams registry publish       # Publish your Substreams to substreams.dev
`+"```"+`

`, destDir).Cmd(),
		act.Cmd(),
	}
}

func (c *Conversation[X]) HandleDownloaded(destDir string) loop.Cmd {
	return loop.Seq(
		c.downloadedCommands(destDir)...,
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

	if c.clientVersion >= 4 {
		return downloadCmd.Cmd()
	}

	return loop.Seq(
		downloadCmd.Cmd(),
		c.HandleDownloaded("{project folder}"),
	)

}
