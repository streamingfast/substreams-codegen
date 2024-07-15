package main

import (
	"fmt"
	_ "net/http/pprof"
	"regexp"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/substreams-codegen/server"
	"go.uber.org/zap"
)

// Commit sha1 value, injected via go build `ldflags` at build time
var commit = ""

// Version value, injected via go build `ldflags` at build time
var version = "dev"

// Date value, injected via go build `ldflags` at build time
var date = ""

var zlog, tracer = logging.RootLogger("substreams-codegen", "github.com/streamingfast/substreams-codegen/cmd/substreams-codegen")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zap.InfoLevel))
}

func main() {
	Run("substreams-codegen", "Substreams Code Generation API",
		Execute(apiE),
		ConfigureViper("CODEGEN"),
		ConfigureVersion("dev"),

		PersistentFlags(
			func(flags *pflag.FlagSet) {
				flags.Duration("delay-before-start", 0, "[OPERATOR] Amount of time to wait before starting any internal processes, can be used to perform to maintenance on the pod before actually letting it starts")
				flags.String("metrics-listen-addr", "localhost:9102", "[OPERATOR] If non-empty, the process will listen on this address for Prometheus metrics request(s)")
				flags.String("pprof-listen-addr", "localhost:6060", "[OPERATOR] If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")
				flags.String("log-format", "text", "Format for logging to stdout. Either 'text' or 'stackdriver'. When 'text', if the standard output is detected to be interactive, colored text is output, otherwise non-colored text.")
				flags.String("session-store-url", "", "Optional store to save session information (ex: file://./sessions or gs://bucket/sessions)")
				flags.String("http-listen-addr", ":9000", "http listen address")
				flags.String("cors-host-regex-allow", "^localhost", "Regex to allow CORS origin requests from, defaults to localhost only")
			},
		),
		AfterAllHook(func(cmd *cobra.Command) {
			cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
				if err := setupCmd(cmd); err != nil {
					return err
				}
				return nil
			}
		}),
	)
}

func apiE(cmd *cobra.Command, args []string) error {
	app := NewApplication(cmd.Context())

	httpListenAddr := sflags.MustGetString(cmd, "http-listen-addr")
	corsHostRegexAllow := sflags.MustGetString(cmd, "cors-host-regex-allow")
	sessionStoreURL := sflags.MustGetString(cmd, "session-store-url")

	sessionStore, err := dstore.NewStore(sessionStoreURL, "", "", false)
	if err != nil {
		return fmt.Errorf("failed to create session store: %w", err)
	}

	zlog.Info("starting substreams-codegen api",
		zap.String("http_listen_addr", httpListenAddr),
		zap.String("cors_host_regex_allow", corsHostRegexAllow),
		zap.String("session_store_url", sessionStoreURL),
	)

	var cors *regexp.Regexp
	if corsHostRegexAllow != "" {
		hostRegex, err := regexp.Compile(corsHostRegexAllow)
		if err != nil {
			return fmt.Errorf("failed to compile cors host regex: %w", err)
		}
		cors = hostRegex
	}

	server := server.New(
		httpListenAddr,
		cors,
		sessionStore,
		zlog)

	app.SuperviseAndStart(server)
	return app.WaitForTermination(zlog, 0, 0)
}
