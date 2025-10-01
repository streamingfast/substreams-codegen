package base

import (
	"strings"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
	codegen "github.com/streamingfast/substreams-codegen"
)

var _ codegen.ConversationState = (*ConversationState)(nil)

// ConversationState implements common fields to implement the ConversationState interface
type ConversationState struct {
	Name string `json:"name"`
	// ChainName while named with `name` does contain actually the network registry chain ID.
	ChainName string `json:"chainName"`
}

func (p *ConversationState) IsValidChainInput(input string) bool {
	return networks.GetSubstreamsRegistry().Has(input)
}

// GetChainName implements codegen.ConversationState.
func (p *ConversationState) GetChainName() string {
	return p.ChainName
}

// GetModuleName implements codegen.ConversationState.
func (p *ConversationState) GetModuleName() string {
	return strings.ReplaceAll(p.Name, "-", "_")
}

// GetChainNetwork is a legacy name used in templates that returns the Substreams
// value that should be put in the `network` field of the `substreams.yaml` file.
// It uses the ChainName field to lookup the network in the registry and return
// its FullName. If ChainName is empty or the config cannot be found, an empty
// string is returned.
func (p *ConversationState) GetChainNetwork() string {
	if network := p.FindNetworkFromChainName(); network != nil {
		return network.ID
	}
	return ""
}

// GetChainNetwork returns the chain config from the registry based on the ChainName field.
// If ChainName is empty, nil is returned right away, otherwise it attempts to find the config
// in the registry by doing `networks.GetSubstreamsRegistry().Find(p.ChainName)`.
func (p *ConversationState) FindNetworkFromChainName() *registry.Network {
	if p.ChainName == "" {
		return nil
	}

	return networks.GetSubstreamsRegistry().Find(p.ChainName)
}

// ChainDisplayName returns the display name of the chain from the registry based on the ChainName field.
// If ChainName is empty or the config cannot be found, an empty string is returned.
func (p *ConversationState) ChainDisplayName() string {
	if network := p.FindNetworkFromChainName(); network != nil {
		return network.FullName
	}
	return ""
}

// ChainExplorerLink returns the explorer link of the chain from the registry based on the ChainName field.
// If ChainName is empty or the config cannot be found, an empty string is returned.
func (p *ConversationState) ChainExplorerLink() string {
	if network := p.FindNetworkFromChainName(); network != nil && len(network.ExplorerUrls) > 0 {
		return network.ExplorerUrls[0]
	}

	return ""
}

// IsChainTestnet returns true if the chain is a testnet according to the registry information.
func (p *ConversationState) IsChainTestnet() bool {
	if network := p.FindNetworkFromChainName(); network != nil {
		return network.NetworkType == registry.Testnet
	}
	return false
}
