package evm_events_calls_raw

import (
	"sort"

	evm_events_calls "github.com/streamingfast/substreams-codegen/evm-events-calls"
)

var ChainConfigs []*evm_events_calls.ChainConfig

var ChainConfigByID = evm_events_calls.ChainConfigByID

func init() {
	for k, v := range ChainConfigByID {
		v.ID = k
		ChainConfigs = append(ChainConfigs, v)
	}
	sort.Slice(ChainConfigs, func(i, j int) bool {
		return ChainConfigs[i].DisplayName < ChainConfigs[j].DisplayName
	})
}

//// TODO: move to a `_test.go` file
//func (c *ChainConfig) setTestABI(address string, abiFile string) {
//	raw, err := os.ReadFile(abiFile)
//	if err != nil {
//		panic(fmt.Errorf("reading Abi %q: %w", abiFile, err))
//	}
//	abi, err := eth.ParseABIFromBytes(raw)
//	if err != nil {
//		panic(fmt.Errorf("parsing Abi %q: %w", abi, err))
//	}
//}
//
//// TODO: move to a `_test.go` file
//func (c *ChainConfig) setTestInitialBlock(address string, initialBlock uint64) {
//	c.initialBlockCache[address] = initialBlock
//}
//
