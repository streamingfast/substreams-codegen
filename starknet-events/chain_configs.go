package starknet_events

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
	"starknet-mainnet": {
		DisplayName:       "Starknet Mainnet",
		ExplorerLink:      "https://starkscan.co/",
		FirehoseEndpoint:  "mainnet.starknet.streamingfast.io:443",
		Network:           "starknet-mainnet",
		initialBlockCache: make(map[string]uint64),
	},
	"starknet-testnet": {
		DisplayName:       "Starknet Testnet",
		ExplorerLink:      "",
		FirehoseEndpoint:  "testnet.starknet.streamingfast.io:443",
		Network:           "starknet-testnet",
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
