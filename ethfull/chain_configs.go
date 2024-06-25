package ethfull

import (
	"fmt"
	"os"
	"sort"

	"github.com/streamingfast/eth-go"
)

type ChainConfig struct {
	ID               string // Public
	DisplayName      string // Public
	ExplorerLink     string
	ApiEndpoint      string
	FirehoseEndpoint string
	Network          string
	SupportsCalls    bool

	abiCache          map[string]*ABI
	initialBlockCache map[string]uint64
}

var ChainConfigs []*ChainConfig

var ChainConfigByID = map[string]*ChainConfig{
	"mainnet": {
		DisplayName:       "Ethereum Mainnet",
		ExplorerLink:      "https://etherscan.io",
		ApiEndpoint:       "https://api.etherscan.io",
		FirehoseEndpoint:  "mainnet.eth.streamingfast.io:443",
		Network:           "mainnet",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"bnb": {
		DisplayName:       "BNB",
		ExplorerLink:      "https://bscscan.com",
		ApiEndpoint:       "https://api.bscscan.com",
		FirehoseEndpoint:  "bnb.streamingfast.io:443",
		Network:           "bsc",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"polygon": {
		DisplayName:       "Polygon",
		ExplorerLink:      "https://polygonscan.com",
		ApiEndpoint:       "https://api.polygonscan.com",
		FirehoseEndpoint:  "polygon.streamingfast.io:443",
		Network:           "polygon",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"amoy": {
		DisplayName:       "Polygon Amoy Testnet",
		ExplorerLink:      "https://www.okx.com/web3/explorer/amoy",
		ApiEndpoint:       "",
		FirehoseEndpoint:  "amoy.substreams.pinax.network:443",
		Network:           "amoy",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"arbitrum": {
		DisplayName:       "Arbitrum",
		ExplorerLink:      "https://arbiscan.io",
		ApiEndpoint:       "https://api.arbiscan.io",
		FirehoseEndpoint:  "arb-one.streamingfast.io:443",
		Network:           "arbitrum",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"holesky": {
		DisplayName:       "Holesky",
		ExplorerLink:      "https://holesky.etherscan.io/",
		ApiEndpoint:       "https://api-holesky.etherscan.io",
		FirehoseEndpoint:  "holesky.eth.streamingfast.io:443",
		Network:           "holesky",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"sepolia": {
		DisplayName:       "Sepolia Testnet",
		ExplorerLink:      "https://sepolia.etherscan.io",
		ApiEndpoint:       "https://api-sepolia.etherscan.io",
		FirehoseEndpoint:  "sepolia.streamingfast.io:443",
		Network:           "sepolia",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
	},
	"optimism": {
		DisplayName:       "Optimism Mainnet",
		ExplorerLink:      "https://optimistic.etherscan.io",
		ApiEndpoint:       "https://api-optimistic.etherscan.io",
		FirehoseEndpoint:  "opt-mainnet.streamingfast.io:443",
		Network:           "optimism",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     false,
	},
	"avalanche": {
		DisplayName:       "Avalanche C-chain",
		ExplorerLink:      "https://subnets.avax.network/c-chain",
		ApiEndpoint:       "",
		FirehoseEndpoint:  "avalanche-mainnet.streamingfast.io:443",
		Network:           "avalanche",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     false,
	},
	"chapel": {
		DisplayName:       "BNB Chapel Testnet",
		ExplorerLink:      "https://testnet.bscscan.com/",
		ApiEndpoint:       "",
		FirehoseEndpoint:  "chapel.substreams.pinax.network:443",
		Network:           "chapel",
		abiCache:          make(map[string]*ABI),
		initialBlockCache: make(map[string]uint64),
		SupportsCalls:     true,
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
		panic(fmt.Errorf("reading abi %q: %w", abiFile, err))
	}
	abi, err := eth.ParseABIFromBytes(raw)
	if err != nil {
		panic(fmt.Errorf("parsing abi %q: %w", abi, err))
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
