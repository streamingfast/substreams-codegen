package codegen

import (
	"strings"

	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
	networks "github.com/streamingfast/firehose-networks"
)

var _ ConversationState = (*BaseConversationState)(nil)

// ConversationState implements common fields to implement the ConversationState interface
type BaseConversationState struct {
	Name string `json:"name"`
	// ChainName while named with `name` does contain actually the network registry chain ID.
	ChainName string `json:"chainName"`

	// SubstreamsSinkChoice holds the user's choice of Substreams sink.
	SubstreamsSinkChoice *SubstreamsSinkChoice `json:"substreamsSinkChoice,omitempty"`
}

func (p *BaseConversationState) GetSubstreamsSinkChoice() *SubstreamsSinkChoice {
	return p.SubstreamsSinkChoice
}

func (p *BaseConversationState) SetSubstreamsSinkChoice(choice SubstreamsSinkChoice) {
	p.SubstreamsSinkChoice = &choice
}

// GetChainName implements codegen.ConversationState.
func (p *BaseConversationState) GetChainName() string {
	return p.ChainName
}

// SetChainName implements codegen.ConversationState.
func (p *BaseConversationState) SetChainName(name string) {
	p.ChainName = name
}

// SetProjectName implements ConversationState.
func (p *BaseConversationState) SetProjectName(name string) {
	p.Name = name
}

// GetModuleName implements codegen.ConversationState.
func (p *BaseConversationState) GetModuleName() string {
	return strings.ReplaceAll(p.Name, "-", "_")
}

// Generate implements codegen.ConversationState, generate nothing by defaults for now.
func (p *BaseConversationState) Generate() ReturnGenerate {
	return ReturnGenerate{
		Err:          nil,
		ProjectFiles: make(map[string][]byte, 0),
	}
}

// GetChainNetwork is a legacy name used in templates that returns the Substreams
// value that should be put in the `network` field of the `substreams.yaml` file.
// It uses the ChainName field to lookup the network in the registry and return
// its FullName. If ChainName is empty or the config cannot be found, an empty
// string is returned.
func (p *BaseConversationState) GetChainNetwork() string {
	if network := p.FindNetworkFromChainName(); network != nil {
		return network.ID
	}
	return ""
}

// GetChainNetwork returns the chain config from the registry based on the ChainName field.
// If ChainName is empty, nil is returned right away, otherwise it attempts to find the config
// in the registry by doing `networks.GetSubstreamsRegistry().Find(p.ChainName)`.
func (p *BaseConversationState) FindNetworkFromChainName() *registry.Network {
	if p.ChainName == "" {
		return nil
	}

	return networks.GetSubstreamsRegistry().Find(p.ChainName)
}

func (p *BaseConversationState) GetChainDisplayName() string {
	return p.ChainDisplayName()
}

// ChainDisplayName returns the display name of the chain from the registry based on the ChainName field.
// If ChainName is empty or the config cannot be found, an empty string is returned.
func (p *BaseConversationState) ChainDisplayName() string {
	if network := p.FindNetworkFromChainName(); network != nil {
		return network.FullName
	}
	return ""
}

// ChainExplorerLink returns the explorer link of the chain from the registry based on the ChainName field.
// If ChainName is empty or the config cannot be found, an empty string is returned.
func (p *BaseConversationState) ChainExplorerLink() string {
	if network := p.FindNetworkFromChainName(); network != nil && len(network.ExplorerUrls) > 0 {
		return network.ExplorerUrls[0]
	}

	return ""
}

// IsChainTestnet returns true if the chain is a testnet according to the registry information.
func (p *BaseConversationState) IsChainTestnet() bool {
	if network := p.FindNetworkFromChainName(); network != nil {
		return network.NetworkType == registry.Testnet
	}
	return false
}

func (c *BaseConversationState) IsValidChainInput(input string) bool {
	return networks.GetSubstreamsRegistry().Find(input) != nil
}

// GetPackageURL generates a probable GitHub URL for the project based on the project name and chain.
// Returns a URL in the format: https://github.com/username/{project-name}-{chain}
func (p *BaseConversationState) GetPackageURL() string {
	if p.Name == "" {
		return ""
	}
	
	chainSuffix := ""
	if p.ChainName != "" {
		chainSuffix = "-" + p.ChainName
	}
	
	return "https://github.com/username/" + p.Name + chainSuffix
}

// GetPackageDescription generates a description for the package based on the project name and chain.
// Returns a description like: "Substreams module for {project-name} on {chain-display-name}"
func (p *BaseConversationState) GetPackageDescription() string {
	if p.Name == "" {
		return "Substreams module"
	}
	
	desc := "Substreams module for " + p.Name
	
	chainDisplay := p.ChainDisplayName()
	if chainDisplay != "" {
		desc += " on " + chainDisplay
	}
	
	return desc
}
