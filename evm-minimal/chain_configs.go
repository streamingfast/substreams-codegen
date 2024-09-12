package ethminimal

import (
	"sort"
)

type ChainConfig struct {
	ID                   string // Public
	DisplayName          string // Public
	ExplorerLink         string
	ApiEndpoint          string
	ApiEndpointDirect    bool
	FirstStreamableBlock uint64
	Network              string
	SupportsCalls        bool
	APIKeyEnvVar         string

	initialBlockCache map[string]uint64
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"mainnet": {
		DisplayName:          "Ethereum Mainnet",
		FirstStreamableBlock: 0,
		Network:              "mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"bnb": {
		DisplayName:          "BNB",
		FirstStreamableBlock: 0,
		Network:              "bsc",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"polygon": {
		DisplayName:          "Polygon",
		FirstStreamableBlock: 0,
		Network:              "polygon",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"amoy": {
		DisplayName:          "Polygon Amoy Testnet",
		FirstStreamableBlock: 0,
		Network:              "amoy",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"arbitrum": {
		DisplayName:          "Arbitrum",
		Network:              "arbitrum",
		FirstStreamableBlock: 0,
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"holesky": {
		DisplayName:          "Holesky",
		FirstStreamableBlock: 0,
		Network:              "holesky",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"sepolia": {
		DisplayName:          "Sepolia Testnet",
		FirstStreamableBlock: 0,
		Network:              "sepolia",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"optimism": {
		DisplayName:          "Optimism Mainnet",
		FirstStreamableBlock: 0,
		Network:              "optimism",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "CODEGEN_OPTIMISM_API_KEY",
	},
	"avalanche": {
		DisplayName:          "Avalanche C-chain",
		FirstStreamableBlock: 0,
		Network:              "avalanche",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
	},
	"chapel": {
		DisplayName:          "BNB Chapel Testnet",
		FirstStreamableBlock: 0,
		Network:              "chapel",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"sei-mainnet": {
		DisplayName:          "SEI Mainnet (EVM)",
		FirstStreamableBlock: 79123881,
		Network:              "sei-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"base-mainnet": {
		DisplayName:          "Base Mainnet",
		FirstStreamableBlock: 0,
		Network:              "base-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
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

// TODO: move to a `_test.go` file
func (c *ChainConfig) setTestInitialBlock(address string, initialBlock uint64) {
	c.initialBlockCache[address] = initialBlock
}
