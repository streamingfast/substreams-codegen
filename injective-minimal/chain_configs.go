package injectiveminimal

import "sort"

type ChainConfig struct {
	ID          string // Public
	DisplayName string // Public
	Network     string
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"injective-mainnet": {
		DisplayName: "Injective Mainnet",
		Network:     "injective-mainnet",
	},
	"injective-testnet": {
		DisplayName: "Injective Testnet",
		Network:     "injective-testnet",
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
