package tronhelloworld

import (
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
)

func init() {
	codegen.RegisterConversation(
		"tron-hello-world",
		"Example Substreams that reads TRON blocks and extract data from `TransferContracts`.",
		"Use this example as a starting point to create your own custom Substreams, which indexes thed data you need.",
		New,
		59,
		"TRON",
	)
}

type Convo struct {
	*codegen.BaseBlockchainGeneratorConversation[*Project]
}

func New() codegen.Converser {
	return &Convo{codegen.NewBaseBlockchainGeneratorConversation(&Project{
		codegen.BaseConversationState{
			ChainName: "tron",
		},
	}, codegen.SharedFlowConfig{
		ValidChains: chains.TronNetworks(),
	})}
}
