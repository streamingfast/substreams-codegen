package mantra_events

import "sort"

type ChainConfig struct {
	ExplorerLink string
	ID           string // Public
	DisplayName  string // Public
	Network      string
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"mantra-mainnet": {
		ExplorerLink: "https://explorer.mantrachain.io/",
		DisplayName:  "Mantra Mainnet",
		Network:      "mantra-mainnet",
	},
	"mantra-testnet": {
		ExplorerLink: "https://testnet.mantra.explorers.guru/",
		DisplayName:  "Mantra Testnet",
		Network:      "mantra-testnet",
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
