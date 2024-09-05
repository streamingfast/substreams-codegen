package tests

import (
	"fmt"
	"os/exec"
	"testing"

	"golang.org/x/net/context"
)

func TestIntegration(t *testing.T) {
	cases := []struct {
		name      string
		stateFile string
	}{
		{
			name:      "evm-events-calls",
			stateFile: "./evm-events-calls/generator.json",
		},
		{
			name:      "evm-minimal",
			stateFile: "./evm-minimal/generator.json",
		},
		{
			name:      "injective-minimal",
			stateFile: "./injective-minimal/generator.json",
		},
	}

	buildArgs := []string{
		"build",
		"-t",
		"test-image",
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
		t.Run(c.name, func(t *testing.T) {
			runArgs := []string{
				"run",
				"--rm",
				"--platform",
				"linux/amd64",
				"-v",
				fmt.Sprintf("%s:/app/generator.json", c.stateFile),
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
