package tests

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/streamingfast/dstore"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/substreams-codegen/server"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
			name:      "evm-hello-world",
			stateFile: "./evm-hello-world/generator.json",
		},
		{
			name:      "injective-hello-world",
			stateFile: "./injective-hello-world/generator.json",
		},
		{
			name:      "sol-hello-world",
			stateFile: "./sol-hello-world/generator.json",
		},
		{
			name:      "starknet-hello-world",
			stateFile: "./starknet-hello-world/generator.json",
		},
		{
			name:      "near-hello-world",
			stateFile: "./near-hello-world/generator.json",
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
		{
			name:      "sol-anchor-meteora",
			stateFile: "./sol-anchor/meteora.json",
		},
		{
			name:      "sol-anchor-orca",
			stateFile: "./sol-anchor/orca.json",
		},
		{
			name:      "sol-anchor-pump-fun",
			stateFile: "./sol-anchor/pump-fun.json",
		},
		{
			name:      "sol-anchor-jupiter-governance",
			stateFile: "./sol-anchor/jupiter-governance.json",
		},
		{
			name:      "sol-anchor-jupiter-staking",
			stateFile: "./sol-anchor/jupiter-staking.json",
		},
		{
			name:      "sol-anchor-raydium-cp-swap.json",
			stateFile: "./sol-anchor/raydium-cp-swap.json",
		},
		{
			name:      "sol-anchor-oasis.json",
			stateFile: "./sol-anchor/oasis.json",
		},
		{
			name:      "sol-anchor-lifinity.json",
			stateFile: "./sol-anchor/lifinity.json",
		},
		{
			name:      "sol-anchor-bonkswap.json",
			stateFile: "./sol-anchor/bonkswap.json",
		},
		{
			name:      "sol-anchor-sanctum.json",
			stateFile: "./sol-anchor/sanctum.json",
		},
	}

	var zlog, _ = logging.RootLogger("test", "test")
	endpoint := "https://codegen-staging.substreams.dev"
	if integrationTestsAgainstLocal {
		fmt.Println("Starting local server... :51012")
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

	// Determine the correct build context based on current working directory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	
	var buildContext string
	if strings.HasSuffix(cwd, "/tests") {
		// Running from tests directory, build context is current directory
		buildContext = "."
	} else {
		// Running from root directory, build context is tests subdirectory
		buildContext = "./tests"
	}

	buildArgs := []string{
		"build",
		"-t",
		"substreams-test-image",
		buildContext,
	}

	// Add timeout for Docker build to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	fmt.Printf("Building Docker image with command: docker %s\n", strings.Join(buildArgs, " "))
	buildCmd := exec.CommandContext(ctx, "docker", buildArgs...)

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build Docker image: %v\nOutput: %s", err, string(output))
	}
	fmt.Println("Docker image built successfully")

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			runArgs := []string{
				"run",
				"--rm",
				"-t",
				"--name",
				c.name,
				"-v",
				fmt.Sprintf("%s:/app/generator.json", c.stateFile),
				"-e",
				"SUBSTREAMS_CODEGEN_ENDPOINT=" + endpoint,
				"-e",
				"BUF_TOKEN=" + os.Getenv("BUF_TOKEN"),
				"substreams-test-image",
			}

			// Add timeout for Docker run to prevent hanging
			runCtx, runCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer runCancel()
			
			fmt.Printf("Running Docker container for test %s\n", c.name)
			runCmd := exec.CommandContext(runCtx, "docker", runArgs...)
			output, err = runCmd.CombinedOutput()
			if err != nil {
				t.Errorf("Docker run failed for test %s: %v\nOutput: %s", c.name, err, string(output))
			} else {
				fmt.Printf("Test %s completed successfully\n", c.name)
			}

		})
	}
}

func runTestLocally(t *testing.T, generatorPath string) {
	tempDir, err := os.MkdirTemp("", "temp")
	require.NoError(t, err)
	fmt.Println("tempdir:", tempDir)
	//defer os.RemoveAll(tempDir)

	runCommand(t, "", "cp", generatorPath, fmt.Sprintf("%s/state.json", tempDir))
	runCommand(t, tempDir, "substreams", "init", "--state-file", "state.json", "--force-download-cwd")
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
