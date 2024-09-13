package evm_events_calls

import (
	"fmt"
	"os"
	"sort"

	"github.com/streamingfast/eth-go"
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
	ExampleContract      string

	abiCache          map[string]*ABI
	initialBlockCache map[string]uint64
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"mainnet": {
		DisplayName:          "Ethereum Mainnet",
		ExplorerLink:         "https://etherscan.io",
		ApiEndpoint:          "https://api.etherscan.io",
		ExampleContract:      "0x1f98431c8ad98523631ae4a59f267346ea31f984",
		FirstStreamableBlock: 0,
		Network:              "mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "CODEGEN_MAINNET_API_KEY",
	},
	"bnb": {
		DisplayName:          "BNB",
		ExplorerLink:         "https://bscscan.com",
		ApiEndpoint:          "https://api.bscscan.com",
		ExampleContract:      "0x2170ed0880ac9a755fd29b2688956bd959f933f8",
		FirstStreamableBlock: 0,
		Network:              "bsc",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "CODEGEN_BNB_API_KEY",
	},
	"polygon": {
		DisplayName:          "Polygon",
		ExplorerLink:         "https://polygonscan.com",
		ApiEndpoint:          "https://api.polygonscan.com",
		ExampleContract:      "0x3c499c542cef5e3811e1192ce70d8cc03d5c3359",
		FirstStreamableBlock: 0,
		Network:              "polygon",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "CODEGEN_POLYGON_API_KEY",
	},
	"amoy": {
		DisplayName:          "Polygon Amoy Testnet",
		ExplorerLink:         "https://www.okx.com/web3/explorer/amoy",
		ExampleContract:      "0x0000000071727de22e5e9d8baf0edac6f37da032",
		ApiEndpoint:          "",
		FirstStreamableBlock: 0,
		Network:              "amoy",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"arbitrum": {
		DisplayName:          "Arbitrum",
		ExplorerLink:         "https://arbiscan.io",
		ExampleContract:      "0x58318bceaa0d249b62fad57d134da7475e551b47",
		ApiEndpoint:          "https://api.arbiscan.io",
		Network:              "arbitrum",
		FirstStreamableBlock: 0,
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"holesky": {
		DisplayName:          "Holesky",
		ExplorerLink:         "https://holesky.etherscan.io/",
		ExampleContract:      "0xade8b182898240910fe9f3513db35a1c101b4748",
		ApiEndpoint:          "https://api-holesky.etherscan.io",
		FirstStreamableBlock: 0,
		Network:              "holesky",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"sepolia": {
		DisplayName:          "Sepolia Testnet",
		ExplorerLink:         "https://sepolia.etherscan.io",
		ApiEndpoint:          "https://api-sepolia.etherscan.io",
		ExampleContract:      "0x800ec0d65adb70f0b69b7db052c6bd89c2406ac4",
		FirstStreamableBlock: 0,
		Network:              "sepolia",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"optimism": {
		DisplayName:          "Optimism Mainnet",
		ExplorerLink:         "https://optimistic.etherscan.io",
		ApiEndpoint:          "https://api-optimistic.etherscan.io",
		ExampleContract:      "0x94b008aa00579c1307b0ef2c499ad98a8ce58e58",
		FirstStreamableBlock: 0,
		Network:              "optimism",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "CODEGEN_OPTIMISM_API_KEY",
	},
	"avalanche": {
		DisplayName:          "Avalanche C-chain",
		ExplorerLink:         "https://subnets.avax.network/c-chain",
		ApiEndpoint:          "",
		ExampleContract:      "0x9702230a8ea53601f5cd2dc00fdbc13d4df4a8c7",
		FirstStreamableBlock: 0,
		Network:              "avalanche",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
	},
	"chapel": {
		DisplayName:          "BNB Chapel Testnet",
		ExplorerLink:         "https://testnet.bscscan.com/",
		ApiEndpoint:          "https://api-testnet.bscscan.com",
		ExampleContract:      "0x37ffab7530fbb7e8b4bfec152132929bdcdae3f3",
		FirstStreamableBlock: 0,
		Network:              "chapel",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"sei-mainnet": {
		DisplayName:          "SEI Mainnet (EVM)",
		ExplorerLink:         "0xb75d0b03c06a926e488e2659df1a861f860bd3d1",
		ApiEndpoint:          "https://seitrace.com/pacific-1/api/v2/smart-contracts",
		ApiEndpointDirect:    true,
		FirstStreamableBlock: 79123881,
		Network:              "sei-mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"base": {
		DisplayName:          "Base Mainnet",
		ExplorerLink:         "https://basescan.org",
		ApiEndpoint:          "https://api.basescan.org",
		ExampleContract:      "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
		FirstStreamableBlock: 0,
		Network:              "base",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "CODEGEN_BASE_API_KEY",
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
func (c *ChainConfig) setTestABI(address string, abiFile string) {
	raw, err := os.ReadFile(abiFile)
	if err != nil {
		panic(fmt.Errorf("reading Abi %q: %w", abiFile, err))
	}
	abi, err := eth.ParseABIFromBytes(raw)
	if err != nil {
		panic(fmt.Errorf("parsing Abi %q: %w", abi, err))
	}
	c.abiCache[address] = &ABI{
		abi: abi,
		raw: string(raw),
	}
}

// TODO: move to a `_test.go` file
func (c *ChainConfig) setTestInitialBlock(address string, initialBlock uint64) {
	c.initialBlockCache[address] = initialBlock
}
