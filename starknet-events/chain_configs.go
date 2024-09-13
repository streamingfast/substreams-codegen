package starknet_events

import "sort"

type ChainConfig struct {
	ID             string // Public
	DisplayName    string // Public
	Network        string
	EndpointEnvVar string
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"starknet-mainnet": {
		DisplayName:    "Starknet Mainnet",
		Network:        "starknet-mainnet",
		EndpointEnvVar: "STARKNET_MAINNET_ENDPOINT",
	},
	"starknet-testnet": {
		DisplayName:    "Starknet Testnet",
		Network:        "starknet-testnet",
		EndpointEnvVar: "STARKNET_TESTNET_ENDPOINT",
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
