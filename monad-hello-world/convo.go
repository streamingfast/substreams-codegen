package monadhelloworld

import (
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
)

func init() {
	codegen.RegisterConversation(
		"monad-hello-world",
		"Example Substreams that reads Monad blocks and extract data from transactions and receipts.",
		"Use this example as a starting point to create your own custom Substreams, which indexes the data you need from the Monad blockchain.",
		New,
		85, // Higher weight than EVM (83) since it's a specific chain
		"Monad",
	)
}

type Convo struct {
	*codegen.BaseBlockchainGeneratorConversation[*Project]
}

func New() codegen.Converser {
	return &Convo{codegen.NewBaseBlockchainGeneratorConversation(&Project{
		codegen.BaseConversationState{
			ChainName: "monad-mainnet",
		},
	}, codegen.SharedFlowConfig{
		ValidChains: chains.MonadNetworks(),
	})}
}
