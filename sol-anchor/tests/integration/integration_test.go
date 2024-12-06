package solanchor

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestIntegrationPumpFun(t *testing.T) {
	err := runTest("pump-fun")
	if err != nil {
		t.Fail()
	}
}

func TestIntegrationMeteora(t *testing.T) {
	err := runTest("meteora")
	if err != nil {
		t.Fail()
	}
}

func TestIntegrationOrca(t *testing.T) {
	err := runTest("orca")
	if err != nil {
		t.Fail()
	}
}

func runTest(generatorName string) error {
	tempDir, err := os.MkdirTemp("", "temp")
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		return err
	}
	fmt.Printf("Temporary directory created: %s\n", tempDir)

	defer os.RemoveAll(tempDir)

	// Move generator to temp directory
	cpCmd := exec.Command("cp", fmt.Sprintf("generators/%s.json", generatorName), fmt.Sprintf("%s/%s.json", tempDir, generatorName))
	output, err := cpCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error moving files: %v\n", err)
		return err
	}
	fmt.Printf("Output %s\n", output)

	// Set the command to execute
	cmd := exec.Command("substreams", "init", "--state-file", fmt.Sprintf("%s.json", generatorName))
	cmd.Env = append(os.Environ(), "SUBSTREAMS_CODEGEN_ENDPOINT=http://localhost:9000")
	cmd.Dir = tempDir

	// Run the command and capture the output
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		return err
	}

	fmt.Printf("Command output:\n%s\n", output)

	// Execute build
	cmd1 := exec.Command("substreams", "build")
	cmd1.Dir = tempDir

	output1, err := cmd1.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		return err
	}

	fmt.Printf("Command output:\n%s\n", output1)

	return nil
}
