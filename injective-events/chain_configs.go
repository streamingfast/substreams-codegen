package injective_events

import "sort"

type ChainConfig struct {
	ExplorerLink string
	ID           string // Public
	DisplayName  string // Public
	Network      string
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"injective-mainnet": {
		ExplorerLink: "https://explorer.injective.network/",
		DisplayName:  "Injective Mainnet",
		Network:      "injective-mainnet",
	},
	"injective-testnet": {
		ExplorerLink: "https://testnet.explorer.injective.network/",
		DisplayName:  "Injective Testnet",
		Network:      "injective-testnet",
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
