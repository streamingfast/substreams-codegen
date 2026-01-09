package codegen

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	networks "github.com/streamingfast/firehose-networks"
	"github.com/streamingfast/substreams-codegen/loop"
	"github.com/streamingfast/substreams-codegen/text"
)

var (
	substreamsSinkChoices = []MapEntry{
		mapEntry(string(SubstreamsSinkChoiceSourceOnly), "Just generate the source code (no sink setup)"),
		mapEntry(string(SubstreamsSinkChoicePostgres), "To Postgres"),
		mapEntry(string(SubstreamsSinkChoiceClickhouse), "To Clickhouse"),
		mapEntry(string(SubstreamsSinkChoiceParquet), "To Parquet Files"),
		mapEntry(string(SubstreamsSinkChoiceGolang), "Write a custom sink in Go"),
		mapEntry(string(SubstreamsSinkChoiceRust), "Write a custom sink in Rust"),
		mapEntry(string(SubstreamsSinkChoiceJavascript), "Write a custom sink in JavaScript/TypeScript"),
		mapEntry(string(SubstreamsSinkChoicePython), "Write a custom sink in Python"),
	}

	substreamsSinkLabels = mapSlice(substreamsSinkChoices, mapEntryValue)
	substreamsSinkValues = mapSlice(substreamsSinkChoices, mapEntryKey)
)

func (c *Conversation[X]) IsPostSharedFlowMsg(msg loop.Msg, config SharedFlowConfig) bool {
	switch msg.(type) {
	case AskSubstreamsConsumptionChoice, InputSubstreamsConsumptionChoice,
		RunGenerate, ReturnGenerate,
		InputSourceDownloaded:
		return true
	default:
		return false
	}
}

func (c *Conversation[X]) NextPostSharedFlowStep(config SharedFlowConfig) loop.Cmd {
	if c.State.GetSubstreamsSinkChoice() == nil {
		return Cmd(AskSubstreamsConsumptionChoice{})
	}

	return Cmd(RunGenerate{})
}

func (c *Conversation[X]) UpdatePostSharedFlowMsg(msg loop.Msg, config SharedFlowConfig) loop.Cmd {
	switch msg := msg.(type) {
	case AskSubstreamsConsumptionChoice:
		return c.HandleAskSubstreamsSinkChoice()

	case InputSubstreamsConsumptionChoice:
		return c.HandleSubstreamsConsumptionChoice(msg.Value)

	case RunGenerate:
		return c.HandleRunGenerate(c.State.Generate)

	case ReturnGenerate:
		return c.HandleReturnGenerate(msg)

	case InputSourceDownloaded:
		return c.HandleSourceDownloaded(msg.Value)
	}

	return loop.Quit(fmt.Errorf("post-shared flow message not handled, have you validated it was a post-shared flow message before hand?: %T", msg))
}

func (c *Conversation[X]) HandleAskSubstreamsSinkChoice() loop.Cmd {
	return c.Action(InputSubstreamsConsumptionChoice{}).ListSelect("How would you like to consume the Substreams?", "consumption").
		Labels(substreamsSinkLabels...).
		Values(substreamsSinkValues...).
		Cmd()
}

func (c *Conversation[X]) HandleSubstreamsConsumptionChoice(value string) loop.Cmd {
	sinkChoice, err := ParseSubstreamsSinkChoice(value)
	if err != nil {
		return loop.Seq(c.Msg().Messagef("Invalid choice: %s", err).Cmd())
	}

	c.State.SetSubstreamsSinkChoice(sinkChoice)

	return Cmd(RunGenerate{})
}

func (c *Conversation[X]) HandleRunGenerate(f func() ReturnGenerate) loop.Cmd {
	return loop.Seq(
		c.Msg().Message("Generating Substreams module source code...").Cmd(),
		func() loop.Msg {
			return f()
		},
	)
}

func (c *Conversation[X]) HandleReturnGenerate(msg ReturnGenerate) loop.Cmd {
	if msg.Err != nil {
		return loop.Seq(
			c.Msg().Messagef("Code generation failed with error: %s", msg.Err).Cmd(),
			loop.Quit(msg.Err),
		)
	}

	downloadCmd := c.createDownloadFilesCommand(msg)

	if c.clientVersion >= 4 {
		return downloadCmd.Cmd()
	}

	return loop.Seq(
		downloadCmd.Cmd(),
		c.HandleSourceDownloaded("{project folder}"),
	)
}

func (c *Conversation[X]) HandleSourceDownloaded(destDir string) loop.Cmd {
	return loop.Seq(
		c.Msg().Message(c.generateFinishInstructions(destDir)).Cmd(),
		loop.Quit(nil),
	)
}

func (c *Conversation[X]) createDownloadFilesCommand(msg ReturnGenerate) *MsgWrap {
	downloadCmd := c.Action(InputSourceDownloaded{}).DownloadFiles()

	for _, fileName := range slices.Sorted(maps.Keys(msg.ProjectFiles)) {
		fileDescription := ""
		if _, ok := FileDescriptions[fileName]; ok {
			fileDescription = FileDescriptions[fileName]
		}
		downloadCmd.AddFile(fileName, msg.ProjectFiles[fileName], "text/plain", fileDescription)
	}

	if dockerContent := c.generateDockerComposeContent(); dockerContent != nil {
		downloadCmd.AddFile("docker-compose.yaml", []byte(*dockerContent), "text/yaml", "Docker Compose file to start the selected sink")
	}

	return downloadCmd
}

const (
	mdBacktick  = "`"
	mdCodeBlock = "```"
)

func (c *Conversation[X]) generateFinishInstructions(destDir string) string {
	content := &strings.Builder{}

	content.WriteString(text.Dedent(`
	  Your Substreams project is ready! Start streaming with:

	  %[1]sbash
	  cd %[2]s
	  substreams auth                   # Authenticates your CLI with your thegraph.market account

	  substreams build                  # Build your Substreams module
	  substreams gui                    # Get streaming!
	  %[1]s
	`, mdCodeBlock, destDir))

	contentSink := c.generateSinkInstructions()
	if contentSink != nil {
		content.WriteString("\n\n")
		content.WriteString(*contentSink)
	}

	content.WriteString("\n\n")
	content.WriteString(text.Dedent(`
	  Optionally, publish your Substreams to the Substreams Registry (https://substreams.dev) with:

	  %[1]sbash
	  substreams registry login         # Login to substreams.dev
	  substreams registry publish       # Publish your Substreams to substreams.dev
	  %[1]s
	`, mdCodeBlock))

	return content.String()
}

func (c *Conversation[X]) generateSinkInstructions() *string {
	sinkChoice := c.State.GetSubstreamsSinkChoice()
	if sinkChoice == nil || *sinkChoice == SubstreamsSinkChoiceNone || *sinkChoice == SubstreamsSinkChoiceSourceOnly {
		zlog.Debug("no sink choice configured or source-only selected, no sink instructions to generate")
		return nil
	}

	endpoint := networks.GetSubstreamsEndpoint(c.State.GetChainName())
	if endpoint == "" {
		// Fallback to placeholders if we can't determine the values
		endpoint = "<endpoint>"
	}

	outputModule := c.State.GetModuleName()
	if outputModule == "" {
		outputModule = "<output_module>"
	}

	switch *sinkChoice {
	case SubstreamsSinkChoicePostgres:
		return ptr(text.Dedent(`
		  Sink to Postgres:

		  1. Get the binary from https://github.com/streamingfast/substreams-sink-sql
		  1. Start Postgres via Docker %[1]sdocker compose up -d%[1]s
		  1. Run %[1]ssubstreams-sink-sql from-proto psql://dev:insecure@localhost:5432/main substreams.yaml %s%[1]s
		  1. See https://docs.substreams.dev/how-to-guides/sinks/sql/relational-mappings for more details.
		`, mdBacktick, outputModule))

	case SubstreamsSinkChoiceClickhouse:
		return ptr(text.Dedent(`
		  Sink to Clickhouse:

		  1. Get the binary from https://github.com/streamingfast/substreams-sink-sql
		  1. Start Clickhouse via Docker %[1]sdocker compose up -d%[1]s
		  1. Run %[1]ssubstreams-sink-sql from-proto clickhouse://default:@localhost:9000/default substreams.yaml %s%[1]s
		  1. See https://docs.substreams.dev/how-to-guides/sinks/sql-sink for more details.
		`, mdBacktick, outputModule))

	case SubstreamsSinkChoiceParquet:
		return ptr(text.Dedent(`
		  Sink to Parquet file:

		  1. Get the binary from https://github.com/streamingfast/substreams-sink-files/ (version 2.1.0 or above)
		  1. Run %[1]ssubstreams-sink-files run %s substreams.yaml %s ./output%[1]s
		  1. See https://github.com/streamingfast/substreams-sink-files?tab=readme-ov-file#parquet for more details.
		`, mdBacktick, endpoint, outputModule))

	case SubstreamsSinkChoiceGolang:
		return ptr(text.Dedent(`
		  Sink using Golang:

		  1. See https://docs.substreams.dev/how-to-guides/sinks/stream/go for an example.
		`))

	case SubstreamsSinkChoiceRust:
		return ptr(text.Dedent(`
		  Sink using Rust:

		  1. See https://github.com/streamingfast/substreams-sink-examples/tree/master/rust#readme for an example.
		`))

	case SubstreamsSinkChoiceJavascript:
		return ptr(text.Dedent(`
		  Sink using Javascript:

		  1. See https://docs.substreams.dev/how-to-guides/sinks/stream/javascript for an example.
		`))

	case SubstreamsSinkChoicePython:
		return ptr(text.Dedent(`
		  Sink using Python:

		  1. See https://github.com/streamingfast/substreams-sink-examples/blob/master/python/README.md for an example.
		`))

	default:
		return nil
	}
}

func (c *Conversation[X]) generateDockerComposeContent() *string {
	sinkChoice := c.State.GetSubstreamsSinkChoice()
	if sinkChoice == nil || *sinkChoice == SubstreamsSinkChoiceSourceOnly {
		zlog.Debug("no sink choice configured or source-only selected, no docker compose to generate")
		return nil
	}

	switch *sinkChoice {
	case SubstreamsSinkChoicePostgres:
		return ptr(text.Dedent(`
          services:
            postgres:
              container_name: postgres-substreams
              image: postgres:17
              ports:
                - "5432:5432"
              command: ["postgres", "-cshared_preload_libraries=pg_stat_statements"]
              environment:
                POSTGRES_USER: dev
                POSTGRES_PASSWORD: insecure
                POSTGRES_DB: main
                POSTGRES_INITDB_ARGS: "-E UTF8 --locale=C"
                POSTGRES_HOST_AUTH_METHOD: md5
              volumes:
                - ./data/postgres:/var/lib/postgresql/data
              healthcheck:
                test: ["CMD", "pg_isready"]
                interval: 30s
                timeout: 10s
                retries: 15
            pgweb:
              container_name: pgweb-substreams
              image: sosedoff/pgweb:0.16.1
              restart: on-failure
              ports:
                - "8081:8081"
              command: ["pgweb", "--bind=0.0.0.0", "--listen=8081", "--binary-codec=hex"]
              links:
                - postgres:postgres
              environment:
                - PGWEB_DATABASE_URL=postgres://dev:insecure@postgres:5432/main?sslmode=disable
              depends_on:
                - postgres
		`))

	case SubstreamsSinkChoiceClickhouse:
		return ptr(text.Dedent(`
          services:
            clickhouse:
              container_name: clickhouse-substreams
              image: clickhouse/clickhouse-server:latest
              user: "101:101"
              hostname: clickhouse
              ports:
                - "8123:8123"
                - "9000:9000"
                - "9005:9005"
		`))
	}

	return nil
}

type MapEntry struct {
	Key   string
	Value string
}

func mapEntry(k, v string) MapEntry {
	return MapEntry{
		Key:   k,
		Value: v,
	}
}

func mapEntryKey(e MapEntry) string {
	return e.Key
}

func mapEntryValue(e MapEntry) string {
	return e.Value
}

func mapSlice[T any, V any](slice []T, fn func(T) V) []V {
	out := make([]V, 0, len(slice))
	for _, item := range slice {
		out = append(out, fn(item))
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}
