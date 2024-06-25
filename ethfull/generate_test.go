package ethfull

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/saracen/fastzip"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	tpls, err := codegen.ParseFS(nil, templatesFS, "**/*.gotmpl")
	require.NoError(t, err)
	_ = tpls
}

func Test_Generate(t *testing.T) {
	cases := []struct {
		name          string
		generatorFile string
		outputType    outputType
	}{
		{
			name:          "uniswap_factory_track_calls",
			generatorFile: "./testdata/uniswap_factory_track_calls.json",
			outputType:    outputTypeSubgraph,
		},
		{
			name:          "uniswap_factory_track_calls_events",
			generatorFile: "./testdata/uniswap_factory_track_calls_events.json",
			outputType:    outputTypeSubgraph,
		},
		{
			name:          "uniswap_factory_events_dynamic_calls",
			generatorFile: "./testdata/uniswap_factory_events_dynamic_calls.json",
			outputType:    outputTypeSubgraph,
		},
		{
			name:          "multiple_contract_with_factory",
			generatorFile: "./testdata/multiple_contract_with_factory.json",
			outputType:    outputTypeSubgraph,
		},
		{
			name:          "multiple_factories",
			generatorFile: "./testdata/multiple_factories.json",
			outputType:    outputTypeSubgraph,
		},
		{
			name:          "track events postgres sql",
			generatorFile: "./testdata/uniswap_track_events_sql.json",
			outputType:    outputTypeSQL,
		},
		{
			name:          "track events clickhouse",
			generatorFile: "./testdata/uniswap_track_events_clickhouse.json",
			outputType:    outputTypeSQL,
		},
		{
			name:          "track calls postgres sql",
			generatorFile: "./testdata/uniswap_track_calls_sql.json",
			outputType:    outputTypeSQL,
		},
		{
			name:          "track calls clickhouse",
			generatorFile: "./testdata/uniswap_track_calls_clickhouse.json",
			outputType:    outputTypeSQL,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := testProjectFromState(t, c.generatorFile)

			assert.Equal(t, RunDecodeContractABI{}, p.NextStep()())
			for _, contract := range p.Contracts {
				res := cmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
				require.NoError(t, res.err)
				contract.abi = res.abi
			}

			for _, dynamicContract := range p.DynamicContracts {
				res := cmdDecodeDynamicABI(dynamicContract)().(ReturnRunDecodeDynamicContractABI)
				require.NoError(t, res.err)
				dynamicContract.abi = res.abi

				for _, contract := range p.Contracts {
					if contract.Name == dynamicContract.ParentContractName {
						dynamicContract.parentContract = contract
					}
				}
			}

			p.outputType = c.outputType

			srcZip, projZip, err := p.generate(c.outputType)
			require.NoError(t, err)
			assert.NotEmpty(t, srcZip)
			assert.NotEmpty(t, projZip)

			sourceTmpDir, err := os.MkdirTemp(os.TempDir(), "test_output_src.zip")
			require.NoError(t, err)

			projectTmpDir, err := os.MkdirTemp(os.TempDir(), "test_output_project.zip")
			require.NoError(t, err)

			if os.Getenv("TEST_KEEP_TEST_OUTPUT") != "true" {
				defer func() {
					_ = os.RemoveAll(sourceTmpDir)
					_ = os.RemoveAll(projectTmpDir)
				}()
			}

			extractor, err := fastzip.NewExtractorFromReader(bytes.NewReader(projZip), int64(len(projZip)), projectTmpDir)
			require.NoError(t, err)
			require.NoError(t, extractor.Extract(context.Background()))

			extractor, err = fastzip.NewExtractorFromReader(bytes.NewReader(srcZip), int64(len(srcZip)), sourceTmpDir)
			require.NoError(t, err)
			require.NoError(t, extractor.Extract(context.Background()))

		})
	}
}

func TestUniFactory(t *testing.T) {
	p := testProjectFromState(t, "./testdata/uniswap_factory_v3.json")

	p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, p.NextStep()())

	for _, contract := range p.Contracts {
		res := cmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.err)
		contract.abi = res.abi
	}

	srcZip, projZip, err := p.generate(outputTypeSubgraph)
	require.NoError(t, err)
	assert.NotEmpty(t, srcZip)
	assert.NotEmpty(t, projZip)
	p.sourceZip = srcZip
	p.projectZip = projZip

	outDir := "testoutput/uniswap_factory_v3"
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)
	extractor, err := fastzip.NewExtractorFromReader(bytes.NewReader(projZip), int64(len(projZip)), outDir)
	require.NoError(t, err)
	require.NoError(t, extractor.Extract(context.Background()))

	extractor, err = fastzip.NewExtractorFromReader(bytes.NewReader(srcZip), int64(len(srcZip)), outDir)
	require.NoError(t, err)
	require.NoError(t, extractor.Extract(context.Background()))

	// requires a build server. Test manually by running `make all` in the unifactory directory

	// artifacts, err := p.build()
	// require.NoError(t, err)
	// assert.Contains(t, artifacts.logs, "Finished release")
}

func TestBaycSQL(t *testing.T) {
	p := testProjectFromState(t, "./testdata/bayc.state.json")

	p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, p.NextStep()())

	for _, contract := range p.Contracts {
		res := cmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.err)
		contract.abi = res.abi
	}

	srcZip, projZip, err := p.generate(outputTypeSubgraph)
	require.NoError(t, err)
	assert.NotEmpty(t, srcZip)
	assert.NotEmpty(t, projZip)
}

func Test_Uniswapv3riggersDynamicDatasources(t *testing.T) {
	p := testProjectFromState(t, "./testdata/uniswap_v3_dynamic_datasources.state.json")

	p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, p.NextStep()())

	for _, contract := range p.Contracts {
		res := cmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.err)
		contract.abi = res.abi
	}

	for _, contract := range p.DynamicContracts {
		res := cmdDecodeDynamicABI(contract)().(ReturnRunDecodeDynamicContractABI)
		require.NoError(t, res.err)
		contract.abi = res.abi
		contract.parentContract = p.Contracts[0]
	}

	srcZip, projZip, err := p.generate(outputTypeSubgraph)
	require.NoError(t, err)
	assert.NotEmpty(t, srcZip)
	assert.NotEmpty(t, projZip)
	p.sourceZip = srcZip
	p.projectZip = projZip

	outDir := "testoutput/uniswap_v3_triggers_dynamic_datasources"
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)
	extractor, err := fastzip.NewExtractorFromReader(bytes.NewReader(projZip), int64(len(projZip)), outDir)
	require.NoError(t, err)
	require.NoError(t, extractor.Extract(context.Background()))

	extractor, err = fastzip.NewExtractorFromReader(bytes.NewReader(srcZip), int64(len(srcZip)), outDir)
	require.NoError(t, err)
	require.NoError(t, extractor.Extract(context.Background()))

	// requires a build server. Test manually by running `make package` in the bayc directory

	// artifacts, err := p.build()
	// require.NoError(t, err)
	// assert.Contains(t, artifacts.logs, "Finished release")
}

func Test_BaycTriggers(t *testing.T) {
	p := testProjectFromState(t, "./testdata/bayc.state.json")

	p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, p.NextStep()())

	for _, contract := range p.Contracts {
		res := cmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.err)
		contract.abi = res.abi
	}

	srcZip, projZip, err := p.generate(outputTypeSubgraph)
	require.NoError(t, err)
	assert.NotEmpty(t, srcZip)
	assert.NotEmpty(t, projZip)
	p.sourceZip = srcZip
	p.projectZip = projZip

	outDir := "testoutput/bayc_triggers"
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)
	extractor, err := fastzip.NewExtractorFromReader(bytes.NewReader(projZip), int64(len(projZip)), outDir)
	require.NoError(t, err)
	require.NoError(t, extractor.Extract(context.Background()))

	extractor, err = fastzip.NewExtractorFromReader(bytes.NewReader(srcZip), int64(len(srcZip)), outDir)
	require.NoError(t, err)
	require.NoError(t, extractor.Extract(context.Background()))

	// requires a build server. Test manually by running `make package` in the bayc directory

	// artifacts, err := p.build()
	// require.NoError(t, err)
	// assert.Contains(t, artifacts.logs, "Finished release")
}

func TestProtoFieldName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no starting underscore",
			input:    "tokenId",
			expected: "tokenId",
		},
		{
			name:     "input starting with an underscore",
			input:    "_tokenId",
			expected: "u_tokenId",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, codegen.SanitizeProtoFieldName(test.input))
		})
	}
}

func testProjectFromState(t *testing.T, stateFile string) *Project {
	cnt, err := os.ReadFile(stateFile)
	require.NoError(t, err)

	p := &Project{}
	require.NoError(t, json.Unmarshal(cnt, p))

	return p
}
