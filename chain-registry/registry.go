package chainregistry

import (
	"fmt"
	"strings"
	"sync"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
)

//go:embed fallback_TheGraphNetworkRegistry_0.6.59.json
var embeddedRegistryJSON []byte

// Package networks provides access to The Graph's network registry and helpers.

// NetworkRegistry is a thin wrapper around a [map[string]*registry.Network] to add some helper methods.
type NetworkRegistry map[string]*registry.Network

var (
	RegistryNetworks     NetworkRegistry
	registryNetworksOnce sync.Once
)

// FetchRegistryNetworks fetches and caches all networks from the registry (no filtering).
func FetchRegistryNetworks() NetworkRegistry {
	registryNetworksOnce.Do(func() {
		reg, err := registry.FromLatestVersion()
		if err != nil {
			// Add fallback registry?
			panic(fmt.Sprintf("Failed to load registry from both network and embedded JSON: %v", err))
		}
		m := make(map[string]*registry.Network)
		for i, net := range reg.Networks {
			m[net.ID] = &reg.Networks[i]
		}
		RegistryNetworks = m

		// Add custom Tron network until it's added to the registry
		RegistryNetworks.addCustomNetwork(&registry.Network{
			ID:        "tron",
			ShortName: "Tron",
			FullName:  "Tron Mainnet",
			Aliases:   []string{"tron-mainnet"},
			Caip2ID:   "eip155:728126428",
			GraphNode: &registry.GraphNode{
				Protocol: (*registry.Protocol)(ptr("tron")),
			},
			ExplorerUrls: []string{"https://tronscan.org"},
			RPCUrls:      []string{"https://api.trongrid.io/jsonrpc", "https://tron.drpc.org", "https://rpc.ankr.com/tron_jsonrpc"},
			APIUrls: []registry.APIURL{
				{
					URL:  "https://apilist.tronscanapi.com/api/",
					Kind: "etherscan",
				},
			},
			Services: registry.Services{
				Firehose:   []string{"mainnet.tron.streamingfast.io:443"},
				Substreams: []string{"mainnet.tron.streamingfast.io:443"},
			},
			NetworkType:     "mainnet",
			IssuanceRewards: true,
			NativeToken:     ptr("TRX"),
			DocsURL:         ptr("https://developers.tron.network/"),
			Genesis: &registry.Genesis{
				Hash:   "0x00000000000000001ebf88508a03865c71d452e25f4d51194196a1d22b6653dc",
				Height: 0,
			},
			Firehose: &registry.Firehose{
				BlockType:        "sf.tron.type.v1.Block",
				EvmExtendedModel: ptr(false),
				BufURL:           "https://buf.build/streamingfast/firehose-tron",
				BytesEncoding:    "hex",
			},
		})
	})
	return RegistryNetworks
}

// addCustomNetwork can be used to add a custom network to the registry map for testing or development.
func (r NetworkRegistry) addCustomNetwork(network *registry.Network) {
	if network == nil || network.ID == "" {
		return // Ignore invalid input
	}
	r[network.ID] = network
}

// GetSubstreamsRegistry returns only networks with Substreams endpoints.
func GetSubstreamsRegistry() NetworkRegistry {
	all := RegistryNetworks
	filtered := make(NetworkRegistry)
	for id, net := range all {
		if len(net.Services.Substreams) > 0 {
			filtered[id] = net
		}
	}
	return filtered
}

func ptr[T any](v T) *T { return &v }

// some chains have noisy '0x' prefixes, some don't, normalize it without 0x
func nox(s string) string {
	return strings.TrimPrefix(s, "0x")
}
