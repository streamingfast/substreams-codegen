package varaextrinsics

import "sort"

type ChainConfig struct {
	ID               string // Public
	DisplayName      string // Public
	ExplorerLink     string
	FirehoseEndpoint string
	Network          string

	initialBlockCache map[string]uint64
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"vara-mainnet": {
		DisplayName:       "Vara Mainnet",
		ExplorerLink:      "https://vara.subscan.io/",
		FirehoseEndpoint:  "mainnet.vara.streamingfast.io:443",
		Network:           "vara-mainnet",
		initialBlockCache: make(map[string]uint64),
	},
	"vara-testnet": {
		DisplayName:       "Vara Testnet",
		ExplorerLink:      "",
		FirehoseEndpoint:  "testnet.vara.streamingfast.io:443",
		Network:           "vara-testnet",
		initialBlockCache: make(map[string]uint64),
	},
}

func init() {
	for k, v := range ChainConfigByID {
		v.ID = k
		ChainConfigs = append(ChainConfigs, v)
	}
	sort.Slice(ChainConfigs, func(i, j int) bool {
		return ChainConfigs[i].DisplayName < ChainConfigs[j].DisplayName
	})
}
