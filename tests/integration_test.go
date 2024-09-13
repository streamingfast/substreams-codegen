package tests

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/streamingfast/logging"

	"github.com/streamingfast/dstore"

	"github.com/streamingfast/substreams-codegen/server"

	"golang.org/x/net/context"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip()
	}

	cases := []struct {
		name                  string
		stateFile             string
		explorerApiKeyEnvName string
		apiKeyNeeded          bool
	}{
		{
			name:                  "evm-events-calls",
			stateFile:             "./evm-events-calls/generator.json",
			explorerApiKeyEnvName: "CODEGEN_MAINNET_API_KEY",
			apiKeyNeeded:          true,
		},
		{
			name:      "evm-minimal",
			stateFile: "./evm-minimal/generator.json",
		},
		{
			name:      "injective-minimal",
			stateFile: "./injective-minimal/generator.json",
		},
		{
			name:      "vara-minimal",
			stateFile: "./vara-minimal/generator.json",
		},
		{
			name:      "sol-minimal",
			stateFile: "./sol-minimal/generator.json",
		},
		{
			name:      "starknet-minimal",
			stateFile: "./starknet-minimal/generator.json",
		},
		{
			name:      "injective-events",
			stateFile: "./injective-events/generator.json",
		},
		{
			name:      "sol-transactions",
			stateFile: "./sol-transactions/generator.json",
		},
		{
			name:      "starknet-events",
			stateFile: "./starknet-events/generator.json",
		},
	}

	ctx := context.Background()

	buildArgs := []string{
		"build",
		"-t",
		"test-image",
		".",
		"--platform",
		"linux/amd64",
	}

	if os.Getenv("TEST_LOCAL_CODEGEN") == "true" {
		go func() {
			var cors *regexp.Regexp
			hostRegex, err := regexp.Compile("^localhost")
			require.NoError(t, err)
			cors = hostRegex

			sessionStore, err := dstore.NewStore("", "", "", false)
			require.NoError(t, err)

			var zlog, _ = logging.RootLogger("test", "test")

			server := server.New(
				":9000",
				cors,
				sessionStore,
				zlog)

			server.Run()
		}()

		//Make sure server is running before, `substreams init`
		time.Sleep(2 * time.Second)
	}

	buildCmd := exec.CommandContext(ctx, "docker", buildArgs...)
	buildCmd.Dir = "./"

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Error(string(output))
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if os.Getenv("TEST_LOCAL_CODEGEN") == "true" {
				explorerApiKey := os.Getenv(c.explorerApiKeyEnvName)
				if explorerApiKey == "" && c.apiKeyNeeded {
					fmt.Printf("NO %s has been provided, please make sure to provide it to enable code generation...", c.explorerApiKeyEnvName)
				}
			}

			runArgs := []string{
				"run",
				"--rm",
				"--name",
				c.name,
				"--platform",
				"linux/amd64",
				"-v",
				fmt.Sprintf("%s:/app/generator.json", c.stateFile),
				"-e",
				fmt.Sprintf("TEST_LOCAL_CODEGEN=%s", os.Getenv("TEST_LOCAL_CODEGEN")),
				"test-image",
			}

			runCmd := exec.CommandContext(ctx, "docker", runArgs...)
			output, err = runCmd.CombinedOutput()
			if err != nil {
				t.Error(string(output))
			}

		})
	}
}
