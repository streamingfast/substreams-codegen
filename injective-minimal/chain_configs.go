package injectiveminimal

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
	"injective-mainnet": {
		DisplayName:       "Injective Mainnet",
		ExplorerLink:      "https://explorer.injective.network/",
		FirehoseEndpoint:  "mainnet.injective.streamingfast.io:443",
		Network:           "injective-mainnet",
		initialBlockCache: make(map[string]uint64),
	},
	"injective-testnet": {
		DisplayName:       "Injective Testnet",
		ExplorerLink:      "https://testnet.explorer.injective.network/",
		FirehoseEndpoint:  "testnet.injective.streamingfast.io:443",
		Network:           "injective-testnet",
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
