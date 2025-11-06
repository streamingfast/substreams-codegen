package stellarminimal

import (
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
)

func init() {
	codegen.RegisterConversation(
		"stellar-minimal",
		"Creates a Substreams project which indexes the full Stellar Block.",
		"You will get a project that indexes all the data contained in the Block.",
		New,
		59,
		"Stellar",
	)
}

type Convo struct {
	*codegen.BaseBlockchainGeneratorConversation[*Project]
}

func New() codegen.Converser {
	return &Convo{codegen.NewBaseBlockchainGeneratorConversation(&Project{}, codegen.SharedFlowConfig{
		ValidChains: chains.StellarNetworks(),
	})}
}
