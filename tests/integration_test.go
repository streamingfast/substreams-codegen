package tests

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/streamingfast/logging"

	"github.com/streamingfast/dstore"

	"github.com/streamingfast/substreams-codegen/server"
)

func TestIntegration(t *testing.T) {
	runIntegrationTests := os.Getenv("RUN_INTEGRATION_TESTS") == "true"
	if !runIntegrationTests {
		t.Skip("RUN_INTEGRATION_TESTS is not set to 'true'")
	}

	integrationTestsInDocker := os.Getenv("INTEGRATION_TESTS_IN_DOCKER") == "true"
	integrationTestsAgainstLocal := os.Getenv("INTEGRATION_TESTS_AGAINST_STAGING") != "true"

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
		//{
		//	name:      "sol-anchor-jupiter",
		//	stateFile: "./sol-anchor/generators/jupiter.json",
		//},
		{
			name:      "sol-anchor-meteora",
			stateFile: "./sol-anchor/generators/meteora.json",
		},
		{
			name:      "sol-anchor-orca",
			stateFile: "./sol-anchor/generators/orca.json",
		},
		{
			name:      "sol-anchor-pump-fun",
			stateFile: "./sol-anchor/generators/pump-fun.json",
		},
	}

	var zlog, _ = logging.RootLogger("test", "test")
	endpoint := "https://codegen-staging.substreams.dev"
	if integrationTestsAgainstLocal {
		launchLocalServer(t, ":51012", zlog)

		switch {
		case integrationTestsInDocker && os.Getenv("CI") == "":
			endpoint = "http://host.docker.internal:51012"
		case integrationTestsInDocker:
			endpoint = "http://172.17.0.1:51012"
		default:
			endpoint = "http://127.0.0.1:51012"
		}
	}

	if integrationTestsInDocker {
		runTestsInDocker(t, cases, endpoint)
		return
	}

	validateBinary(t, "substreams")
	validateBinary(t, "cargo")
	validateBinary(t, "buf")

	parallel := true
	hasSSCache := hasBinary("sscache")
	if !hasSSCache {
		zlog.Info("sscache not found, tests will not run in parallel (run `cargo install sscache` to enable parallelism)")
		parallel = false
	} else {
		os.Setenv("RUSTC_WRAPPER", "sccache")
		os.Setenv("SSCACHE_DIR", filepath.Join(os.TempDir(), "sscachedir"))
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if parallel {
				t.Parallel()
			}
			runTestLocally(t, c.stateFile)
		})
	}
}

func launchLocalServer(t *testing.T, listenAddr string, zlog *zap.Logger) {
	go func() {
		sessionStore := dstore.NewMockStore(func(base string, f io.Reader) (err error) { return nil })
		server := server.New(
			listenAddr,
			nil,
			sessionStore,
			zlog)
		server.OnTerminating(func(e error) {
			require.NoError(t, e)
		})
		server.Run()
	}()

	//Make sure server is running before, `substreams init`
	time.Sleep(2 * time.Second)
}

func runTestsInDocker(t *testing.T, cases []struct {
	name                  string
	stateFile             string
	explorerApiKeyEnvName string
	apiKeyNeeded          bool
}, endpoint string) {

	buildArgs := []string{
		"build",
		"-t",
		"substreams-test-image",
		".",
		"--platform",
		"linux/amd64",
	}

	ctx := context.Background()
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
				"SUBSTREAMS_CODEGEN_ENDPOINT=" + endpoint,
				"substreams-test-image",
			}

			runCmd := exec.CommandContext(ctx, "docker", runArgs...)
			output, err = runCmd.CombinedOutput()
			if err != nil {
				t.Error(string(output))
			}

		})
	}
}

func runTestLocally(t *testing.T, generatorPath string) {
	tempDir, err := os.MkdirTemp("", "temp")
	require.NoError(t, err)
	//defer os.RemoveAll(tempDir)

	runCommand(t, "", "cp", generatorPath, fmt.Sprintf("%s/state.json", tempDir))
	runCommand(t, tempDir, "substreams", "init", "--state-file", "state.json")
	runCommand(t, tempDir, "substreams", "build")
}

func hasBinary(bin string) bool {
	cmd := exec.Command("which", bin)
	_, err := cmd.CombinedOutput()
	return cmd.ProcessState.ExitCode() == 0 && err == nil
}

func validateBinary(t *testing.T, bin string) {
	require.True(t, hasBinary(bin), "cannot find binary %q in PATH", bin)
}

func runCommand(t *testing.T, dir string, bin string, args ...string) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "SUBSTREAMS_CODEGEN_ENDPOINT=http://127.0.0.1:51012")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "error while running %s %v. Output: %s", bin, args, output)
}
