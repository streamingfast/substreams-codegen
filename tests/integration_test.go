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

	"github.com/streamingfast/dstore"
	"github.com/streamingfast/substreams-codegen/server"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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
	ctx := context.Background()

	// Build the Docker image once using Docker CLI (more efficient for parallel tests)
	imageName := "substreams-test-image:latest"
	fmt.Printf("Building Docker image %s for all tests...\n", imageName)

	// Debug Docker environment
	fmt.Printf("Docker version info:\n")
	if versionCmd := exec.Command("docker", "version"); versionCmd != nil {
		if versionOutput, versionErr := versionCmd.CombinedOutput(); versionErr == nil {
			fmt.Printf("%s\n", string(versionOutput))
		} else {
			fmt.Printf("Failed to get Docker version: %v\n", versionErr)
		}
	}

	buildCtx, buildCancel := context.WithTimeout(ctx, 15*time.Minute)
	defer buildCancel()

	// Retry Docker build to handle transient issues
	const maxBuildRetries = 3
	var buildOutput []byte
	var err error

	for attempt := 1; attempt <= maxBuildRetries; attempt++ {
		fmt.Printf("Docker build attempt %d/%d...\n", attempt, maxBuildRetries)

		// Tests are always run in the package folder (here "tests"), so "." refers to "tests" here
		// Use legacy builder to avoid buildkit mount issues in CI
		// Try with cache first, then without cache on retry
		var buildArgs []string
		if attempt == 1 {
			buildArgs = []string{"build", "-t", imageName, "."}
		} else {
			buildArgs = []string{"build", "--no-cache", "-t", imageName, "."}
		}
		buildCmd := exec.CommandContext(buildCtx, "docker", buildArgs...)
		buildCmd.Env = append(os.Environ(), "DOCKER_BUILDKIT=0")
		buildOutput, err = buildCmd.CombinedOutput()

		if err == nil {
			fmt.Printf("Docker image %s built successfully on attempt %d\n", imageName, attempt)
			break
		}

		if attempt < maxBuildRetries {
			fmt.Printf("Docker build attempt %d failed: %v\nOutput: %s\nRetrying in 10 seconds...\n",
				attempt, err, string(buildOutput))
			time.Sleep(10 * time.Second)
		}
	}

	require.NoError(t, err, "Failed to build Docker image after %d attempts: %v\nFinal output: %s",
		maxBuildRetries, err, string(buildOutput))

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// Resolve absolute path for the state file
			absStateFile, err := filepath.Abs(c.stateFile)
			require.NoError(t, err)

			// Create and start container using the pre-built image with retry logic
			fmt.Printf("Starting container for test %s\n", c.name)

			container, err := runContainerWithRetry(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: testcontainers.ContainerRequest{
					Image: imageName,
					Env: map[string]string{
						"SUBSTREAMS_CODEGEN_ENDPOINT": endpoint,
						"BUF_TOKEN":                   os.Getenv("BUF_TOKEN"),
					},
					Files: []testcontainers.ContainerFile{
						{
							HostFilePath:      absStateFile,
							ContainerFilePath: "/app/generator.json",
							FileMode:          0644,
						},
					},
					// The entrypoint will run the test automatically, so this is essentially
					// how long we allow the container to take to complete.
					WaitingFor: wait.ForExit().WithExitTimeout(5 * time.Minute),
				},
				Started: true,
			}, c.name)
			require.NoError(t, err)

			defer func() {
				printContainerLogs(ctx, container, t.Name())
				if err := container.Terminate(ctx); err != nil {
					t.Logf("Failed to terminate container for test %s: %v", c.name, err)
				}
			}()

			// The WaitingFor strategy handles waiting for exit, so the container
			// should already be done when we reach here. Just check the exit code.
			state, err := container.State(ctx)
			require.NoError(t, err, "Failed to get container state for test %s", c.name)

			if state.ExitCode != 0 {
				t.Errorf("Container exited with non-zero code %d for test %s", state.ExitCode, c.name)
			} else {
				fmt.Printf("Test %s completed successfully\n", c.name)
			}
		})
	}
}

// runContainerWithRetry runs a container with retry logic to handle transient failures. The mere
// fact on starting the container will kick in the entrypoint.sh bash script which does all the work.
// So if the container finishes, it means the test is done.
func runContainerWithRetry(ctx context.Context, req testcontainers.GenericContainerRequest, testName string) (testcontainers.Container, error) {
	const maxRetries = 3

	var container testcontainers.Container
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		container, err = testcontainers.GenericContainer(ctx, req)
		if err == nil {
			return container, nil
		}

		// Retry on any error during container startup
		if attempt < maxRetries {
			fmt.Printf("Container startup attempt %d/%d failed for test %s: %v, retrying...\n",
				attempt, maxRetries, testName, err)
			time.Sleep(time.Second) // Brief pause before retry
		}
	}

	return nil, fmt.Errorf("failed to start container after %d attempts: %w", maxRetries, err)
}

func printContainerLogs(ctx context.Context, container testcontainers.Container, testName string) {
	if logs, logErr := container.Logs(ctx); logErr == nil {
		defer logs.Close()
		logBytes := make([]byte, 0, 4096)
		buf := make([]byte, 1024)
		for {
			n, readErr := logs.Read(buf)
			if n > 0 {
				logBytes = append(logBytes, buf[:n]...)
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				break
			}
		}
		if len(logBytes) > 0 {
			fmt.Printf("Container logs for test %s:\n%s\n", testName, string(logBytes))
		}
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
