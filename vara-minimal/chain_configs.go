package varaminimal

import "sort"

type ChainConfig struct {
	ID           string // Public
	DisplayName  string // Public
	ExplorerLink string
	Network      string
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"vara-mainnet": {
		DisplayName:  "Vara Mainnet",
		ExplorerLink: "https://vara.subscan.io/",
		Network:      "vara-mainnet",
	},
	"vara-testnet": {
		DisplayName:  "Vara Testnet",
		ExplorerLink: "",
		Network:      "vara-testnet",
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
