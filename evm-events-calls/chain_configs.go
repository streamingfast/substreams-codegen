package evm_events_calls

import (
	"fmt"
	"net/url"
	"os"
	"sort"

	"github.com/streamingfast/eth-go"
)

type ChainConfig struct {
	ID                   string // Public
	DisplayName          string // Public
	ExplorerLink         string
	ApiEndpoint          string
	ApiBaseURL           string     // Base URL without query parameters to support EtherscanV2
	ApiQueryParams       url.Values // Query parameters to append (e.g. chainid=747474)
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
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"1"}},
		ExampleContract:      "0x1f98431c8ad98523631ae4a59f267346ea31f984",
		FirstStreamableBlock: 0,
		Network:              "mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"bnb": {
		DisplayName:          "BNB",
		ExplorerLink:         "https://bscscan.com",
		ApiEndpoint:          "https://api.bscscan.com",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"56"}},
		ExampleContract:      "0x2170ed0880ac9a755fd29b2688956bd959f933f8",
		FirstStreamableBlock: 0,
		Network:              "bsc",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"polygon": {
		DisplayName:          "Polygon",
		ExplorerLink:         "https://polygonscan.com",
		ApiEndpoint:          "https://api.polygonscan.com",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"137"}},
		ExampleContract:      "ETHERSCAN",
		FirstStreamableBlock: 0,
		Network:              "polygon",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"amoy": {
		DisplayName:          "Polygon Amoy Testnet",
		ExplorerLink:         "https://www.okx.com/web3/explorer/amoy",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"80002"}},
		ApiEndpoint:          "",
		ExampleContract:      "0x0000000071727de22e5e9d8baf0edac6f37da032",
		FirstStreamableBlock: 0,
		Network:              "amoy",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"arbitrum": {
		DisplayName:          "Arbitrum",
		ExplorerLink:         "https://arbiscan.io",
		ApiEndpoint:          "https://api.arbiscan.io",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"42161"}},
		Network:              "arbitrum",
		ExampleContract:      "0x58318bceaa0d249b62fad57d134da7475e551b47",
		FirstStreamableBlock: 0,
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"hoodi": {
		DisplayName:          "Hoodi",
		ExplorerLink:         "https://hoodi.etherscan.io/",
		ApiEndpoint:          "https://api-hoodi.etherscan.io",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"560048"}},
		ExampleContract:      "0x71AF2c3f12FebcCc54E62fB2A0a25968efe2701b",
		FirstStreamableBlock: 0,
		Network:              "hoodi",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"sepolia": {
		DisplayName:          "Sepolia Testnet",
		ExplorerLink:         "https://sepolia.etherscan.io",
		ApiEndpoint:          "https://api-sepolia.etherscan.io",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"11155111"}},
		ExampleContract:      "0x03C66CB1826BDB0395BF31E68Bf7E873e9564fFB",
		FirstStreamableBlock: 0,
		Network:              "sepolia",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"optimism": {
		DisplayName:          "Optimism Mainnet",
		ExplorerLink:         "https://optimistic.etherscan.io",
		ApiEndpoint:          "https://api-optimistic.etherscan.io",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"10"}},
		ExampleContract:      "0x94b008aa00579c1307b0ef2c499ad98a8ce58e58",
		FirstStreamableBlock: 0,
		Network:              "optimism",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"avalanche-mainnet": {
		DisplayName:          "Avalanche C-chain",
		ExplorerLink:         "https://subnets.avax.network/c-chain",
		ApiEndpoint:          "",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"43114"}},
		ExampleContract:      "0x9702230a8ea53601f5cd2dc00fdbc13d4df4a8c7",
		FirstStreamableBlock: 0,
		Network:              "avalanche-mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"chapel": {
		DisplayName:          "BNB Chapel Testnet",
		ExplorerLink:         "https://testnet.bscscan.com/",
		ApiEndpoint:          "https://api-testnet.bscscan.com",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"97"}},
		ExampleContract:      "0x37ffab7530fbb7e8b4bfec152132929bdcdae3f3",
		FirstStreamableBlock: 0,
		Network:              "chapel",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"sei-mainnet": {
		DisplayName:          "SEI Mainnet (EVM)",
		ApiEndpoint:          "https://seitrace.com/pacific-1/api/v2/smart-contracts",
		ApiEndpointDirect:    true,
		ExampleContract:      "0xb75d0b03c06a926e488e2659df1a861f860bd3d1",
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
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"8453"}},
		ExampleContract:      "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
		FirstStreamableBlock: 0,
		Network:              "base",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"tron-evm-mainnet": {
		DisplayName:          "Tron EVM mainnet",
		FirstStreamableBlock: 0,
		Network:              "tron-evm-mainnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
	},
	"unichain": {
		DisplayName:          "Unichain Mainnet",
		ExplorerLink:         "https://unichain.blockscout.com/",
		ApiEndpoint:          "https://unichain.blockscout.com/api",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"130"}},
		ExampleContract:      "0x1f98400000000000000000000000000000000003",
		FirstStreamableBlock: 0,
		Network:              "unichain",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"injective-evm-testnet": {
		DisplayName:          "Injective EVM testnet",
		ExplorerLink:         "https://testnet.blockscout.injective.network",
		ApiEndpoint:          "https://testnet.blockscout-api.injective.network/api",
		FirstStreamableBlock: 0,
		Network:              "injective-evm-testnet",
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
	},
	"katana-mainnet": {
		DisplayName:          "Katana Mainnet",
		ExplorerLink:         "https://katanascan.com",
		ApiEndpoint:          "https://explorer-katana.t.conduit.xyz",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"747474"}},
		ExampleContract:      "0x203A662b0BD271A6ed5a60EdFbd04bFce608FD36",
		FirstStreamableBlock: 0,
		Network:              "katana-mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        true,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"monad-mainnet": {
		DisplayName:          "Monad Mainnet",
		ExplorerLink:         "https://monadscan.com",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"143"}},
		ExampleContract:      "0xA485D7409bdaC5A504D487Ce4d0f2aF40E64d80B",
		FirstStreamableBlock: 0,
		Network:              "monad-mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
	},
	"worldchain-mainnet": {
		DisplayName:  "World Chain Mainnet",
		ExplorerLink: "https://worldscan.org",
		//		ApiEndpoint:          "https://api-optimistic.etherscan.io",
		ApiBaseURL:           "https://api.etherscan.io/v2/api",
		ApiQueryParams:       url.Values{"chainid": {"480"}},
		ExampleContract:      "0x79A02482A880bCE3F13e09Da970dC34db4CD24d1",
		FirstStreamableBlock: 0,
		Network:              "worldchain-mainnet",
		abiCache:             make(map[string]*ABI),
		initialBlockCache:    make(map[string]uint64),
		SupportsCalls:        false,
		APIKeyEnvVar:         "ETHERSCAN_API_KEY",
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
