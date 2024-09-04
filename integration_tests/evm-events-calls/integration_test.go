package main

import (
	"os"
	"testing"

	evm_events_calls "github.com/streamingfast/substreams-codegen/evm-events-calls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvmEventCalls(t *testing.T) {
	p := evm_events_calls.LoadProjectFromState(t, "./testdata/bayc.state.json")

	// p.confirmDoCompile = true
	assert.Equal(t, evm_events_calls.RunDecodeContractABI{}, p.NextStep()())

	for _, contract := range p.Contracts {
		res := cmdDecodeABI(contract)().(evm_events_calls.ReturnRunDecodeContractABI)
		require.NoError(t, res.err)
		contract.abi = res.abi
	}

	projectFiles, err := p.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, len(projectFiles))
	p.projectFiles = projectFiles

	outDir := "testoutput/uniswap_v3_triggers_dynamic_datasources"
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)

	// requires a build server. Test manually by running `make package` in the bayc directory

	// artifacts, err := p.build()
	// require.NoError(t, err)
	// assert.Contains(t, artifacts.logs, "Finished release")
}
