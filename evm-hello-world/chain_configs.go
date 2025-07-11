package ethhelloworld

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
	ContractAddress      string

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
		ContractAddress:      "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
	},
	"bnb": {
		DisplayName:          "BNB",
		FirstStreamableBlock: 0,
		Network:              "bsc",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x8965349fb649a33a30cbfda057d8ec2c48abe2a2",
	},
	"polygon": {
		DisplayName:          "Polygon",
		FirstStreamableBlock: 0,
		Network:              "polygon",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
	},
	"amoy": {
		DisplayName:          "Polygon Amoy Testnet",
		FirstStreamableBlock: 0,
		Network:              "amoy",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x41E94Eb019C0762f9Bfcf9Fb1E58725BfB0e7582",
	},
	"arbitrum": {
		DisplayName:          "Arbitrum",
		Network:              "arbitrum",
		FirstStreamableBlock: 0,
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0xaf88d065e77c8cC2239327C5EDb3A432268e5831",
	},
	"holesky": {
		DisplayName:          "Holesky",
		FirstStreamableBlock: 0,
		Network:              "holesky",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x74a4a85c611679b73f402b36c0f84a7d2ccdfda3",
	},
	"sepolia": {
		DisplayName:          "Sepolia Testnet",
		FirstStreamableBlock: 0,
		Network:              "sepolia",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238",
	},
	"optimism": {
		DisplayName:          "Optimism Mainnet",
		FirstStreamableBlock: 0,
		Network:              "optimism",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "CODEGEN_OPTIMISM_API_KEY",
		ContractAddress:      "0x0b2c639c533813f4aa9d7837caf62653d097ff85",
	},
	"avalanche-mainnet": {
		DisplayName:          "Avalanche C-chain",
		FirstStreamableBlock: 0,
		Network:              "avalanche-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		ContractAddress:      "0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E",
	},
	"chapel": {
		DisplayName:          "BNB Chapel Testnet",
		FirstStreamableBlock: 0,
		Network:              "chapel",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x337610d27c682e347c9cd60bd4b3b107c9d34ddd",
	},
	"sei-mainnet": {
		DisplayName:          "SEI Mainnet (EVM)",
		FirstStreamableBlock: 79123881,
		Network:              "sei-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x3894085ef7ff0f0aedf52e2a2704928d1ec074f1",
	},
	"base-mainnet": {
		DisplayName:          "Base Mainnet",
		FirstStreamableBlock: 0,
		Network:              "base-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
	},
	"tron-evm-mainnet": {
		DisplayName:          "Tron EVM mainnet",
		FirstStreamableBlock: 0,
		Network:              "tron-evm-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		ContractAddress:      "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
	},
	"unichain": {
		DisplayName:          "Unichain Mainnet",
		FirstStreamableBlock: 0,
		Network:              "unichain",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "0x078D782b760474a361dDA0AF3839290b0EF57AD6",
	},
	"injective-evm-testnet": {
		DisplayName:          "Injective EVM testnet",
		FirstStreamableBlock: 0,
		Network:              "injective-evm-testnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		ContractAddress:      "inj1d4y0dcwh2wrgv9rcke0hx6rs2wp4w4vxds87ep",
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
