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
	defer os.RemoveAll(tempDir)

	fmt.Printf("Temporary directory created: %s\n", tempDir)

	// Set the command to execute
	cmd := exec.Command("substreams", "init", "--state-file", fmt.Sprintf("/Users/enolalvarezdeprado/Documents/projects/substreams/substreams-codegen/sol-anchor/tests/integration/generators/%s.json", generatorName)) // Replace "ls" with your desired command
	cmd.Env = append(os.Environ(), "SUBSTREAMS_CODEGEN_ENDPOINT=https://localhost:9000")
	cmd.Dir = tempDir

	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		return err
	}

	fmt.Printf("Command output:\n%s\n", output)

	// Execute build
	cmd1 := exec.Command("substreams", "build")
	cmd.Dir = tempDir

	output1, err := cmd1.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		return err
	}

	fmt.Printf("Command output:\n%s\n", output1)

	return nil
}
