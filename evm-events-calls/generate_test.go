package evm_events_calls

import (
	"encoding/json"
	"os"
	"testing"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	tpls, err := codegen.ParseFS(templatesFS, "**/*.gotmpl")
	require.NoError(t, err)
	_ = tpls
}

type match struct {
	file     string
	contains string
}

func Test_Generate(t *testing.T) {
	cases := []struct {
		name          string
		generatorFile string
		contains      []match
	}{
		{
			name:          "uniswap_factory_track_calls",
			generatorFile: "./testdata/uniswap_factory_track_calls.json",
			contains: []match{
				{"README.md", "## Usage\n\n```bash\nsubstreams build\n"},
			},
		},
		{
			name:          "uniswap_factory_track_calls_events",
			generatorFile: "./testdata/uniswap_factory_track_calls_events.json",
		},
		{
			name:          "uniswap_factory_events_dynamic_calls",
			generatorFile: "./testdata/uniswap_factory_events_dynamic_calls.json",
		},
		{
			name:          "multiple_contract_with_factory",
			generatorFile: "./testdata/multiple_contract_with_factory.json",
			contains: []match{
				{"README.md", "_unifac_ at **0x1f98431c8ad98523631ae4a59f267346ea31f984**\n- _bayc_ at **0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d**"},
			},
		},
		{
			name:          "multiple_factories",
			generatorFile: "./testdata/multiple_factories.json",
		},
		{
			name:          "track events postgres sql",
			generatorFile: "./testdata/uniswap_track_events_sql.json",
		},
		{
			name:          "track events clickhouse",
			generatorFile: "./testdata/uniswap_track_events_clickhouse.json",
		},
		{
			name:          "track calls postgres sql",
			generatorFile: "./testdata/uniswap_track_calls_sql.json",
		},
		{
			name:          "track calls clickhouse",
			generatorFile: "./testdata/uniswap_track_calls_clickhouse.json",
		},
		{
			name:          "Complex Abi with digits and specific character",
			generatorFile: "./testdata/complex_abi.json",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			convo := loadProjectFromState(t, c.generatorFile)
			p := convo.State
			assert.Equal(t, RunDecodeContractABI{}, convo.NextStep()(), c.generatorFile)
			for _, contract := range p.Contracts {
				res := CmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
				require.NoError(t, res.Err)
				contract.Abi = res.Abi
			}

			for _, dynamicContract := range p.DynamicContracts {
				res := cmdDecodeDynamicABI(dynamicContract)().(ReturnRunDecodeDynamicContractABI)
				require.NoError(t, res.err)
				dynamicContract.Abi = res.abi

				for _, contract := range p.Contracts {
					if contract.Name == dynamicContract.ParentContractName {
						dynamicContract.parentContract = contract
					}
				}
			}

			res := p.Generate()
			require.NoError(t, res.Err)
			assert.NotEmpty(t, len(res.ProjectFiles))

			for _, cont := range c.contains {
				assert.Contains(t, res.ProjectFiles, cont.file)
				assert.Contains(t, string(res.ProjectFiles[cont.file]), cont.contains)
			}
		})
	}
}

//func TestGoldenImage(t *testing.T) {
//	cases := []struct {
//		name           string
//		generatorFile  string
//		expectedOutput string
//	}{
//		{
//			name:           "complex_abi",
//			generatorFile:  "./testdata/complex_abi.json",
//			expectedOutput: "./testoutput/complex_abi",
//		},
//	}
//
//	for _, c := range cases {
//		t.Run(c.name, func(t *testing.T) {
//			p := LoadProjectFromState(t, c.generatorFile)
//
//			for _, contract := range p.Contracts {
//				res := CmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
//				require.NoError(t, res.Err)
//				contract.Abi = res.Abi
//			}
//
//			for _, dynamicContract := range p.DynamicContracts {
//				res := cmdDecodeDynamicABI(dynamicContract)().(ReturnRunDecodeDynamicContractABI)
//				require.NoError(t, res.Err)
//				dynamicContract.Abi = res.Abi
//
//				for _, contract := range p.Contracts {
//					if contract.Name == dynamicContract.ParentContractName {
//						dynamicContract.parentContract = contract
//					}
//				}
//			}
//
//			p.outputType = outputTypeSubgraph
//
//			sourceFiles, projectFiles, Err := p.Generate(outputTypeSubgraph)
//			require.NoError(t, Err)
//			assert.NotEmpty(t, len(sourceFiles))
//			assert.NotEmpty(t, len(projectFiles))
//
//			for fileName, fileContent := range projectFiles {
//				goldenFileName := c.expectedOutput + "/" + strings.TrimPrefix(fileName, "substreams/")
//				goldenContent, Err := os.ReadFile(goldenFileName)
//				require.NoError(t, Err)
//
//				require.Equal(t, goldenContent, fileContent)
//			}
//		})
//	}
//
//}

func TestUniFactory(t *testing.T) {
	convo := loadProjectFromState(t, "./testdata/uniswap_factory_v3.json")
	p := convo.State

	// p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, convo.NextStep()())

	for _, contract := range p.Contracts {
		res := CmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.Err)
		contract.Abi = res.Abi
	}

	res := p.Generate()
	require.NoError(t, res.Err)
	assert.NotEmpty(t, len(res.ProjectFiles))

	// requires a build server. Test manually by running `make all` in the unifactory directory

	// artifacts, Err := p.build()
	// require.NoError(t, Err)
	// assert.Contains(t, artifacts.logs, "Finished release")
}

func TestBaycSQL(t *testing.T) {
	convo := loadProjectFromState(t, "./testdata/bayc.state.json")
	p := convo.State

	// p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, convo.NextStep()())

	for _, contract := range p.Contracts {
		res := CmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.Err)
		contract.Abi = res.Abi
	}

	res := p.Generate()
	require.NoError(t, res.Err)
	assert.NotEmpty(t, len(res.ProjectFiles))
}

func Test_Uniswapv3riggersDynamicDatasources(t *testing.T) {
	convo := loadProjectFromState(t, "./testdata/uniswap_v3_dynamic_datasources.state.json")
	p := convo.State

	// p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, convo.NextStep()())

	for _, contract := range p.Contracts {
		res := CmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.Err)
		contract.Abi = res.Abi
	}

	for _, contract := range p.DynamicContracts {
		res := cmdDecodeDynamicABI(contract)().(ReturnRunDecodeDynamicContractABI)
		require.NoError(t, res.err)
		contract.Abi = res.abi
		contract.parentContract = p.Contracts[0]
	}

	res := p.Generate()
	require.NoError(t, res.Err)
	assert.NotEmpty(t, len(res.ProjectFiles))

	outDir := "testoutput/uniswap_v3_triggers_dynamic_datasources"
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)
}
func Test_BaycTriggers(t *testing.T) {
	convo := loadProjectFromState(t, "./testdata/bayc.state.json")
	p := convo.State

	// p.confirmDoCompile = true
	assert.Equal(t, RunDecodeContractABI{}, convo.NextStep()())

	for _, contract := range p.Contracts {
		res := CmdDecodeABI(contract)().(ReturnRunDecodeContractABI)
		require.NoError(t, res.Err)
		contract.Abi = res.Abi
	}

	res := p.Generate()
	require.NoError(t, res.Err)
	assert.NotEmpty(t, len(res.ProjectFiles))

	outDir := "testoutput/uniswap_v3_triggers_dynamic_datasources"
	os.RemoveAll(outDir)
	os.MkdirAll(outDir, 0755)
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

func loadProjectFromState(t *testing.T, stateFile string) *Convo {
	cnt, err := os.ReadFile(stateFile)
	require.NoError(t, err)

	convo := New().(*Convo)
	p := convo.State
	p.currentContractIdx = 0
	require.NoError(t, json.Unmarshal(cnt, p))

	return convo
}
