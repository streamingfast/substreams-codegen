package solhelloworld

import (
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
)

func init() {
	codegen.RegisterConversation(
		"sol-hello-world",
		"Creates a Substreams project that extracts accounts from the Pump.Fun program.",
		"You will get a very simple project to get started with Substreams.",
		New,
		2002,
		"Solana",
	)
}

type Convo struct {
	*codegen.BaseBlockchainGeneratorConversation[*Project]
}

func New() codegen.Converser {
	return &Convo{codegen.NewBaseBlockchainGeneratorConversation(&Project{}, codegen.SharedFlowConfig{
		ValidChains: chains.SolanaNetworks(),
	})}
}
