package nearhelloworld

import (
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/chains"
)

func init() {
	codegen.RegisterConversation(
		"near-hello-world",
		"Example Substreams that reads NEAR blocks and extract data from transactions and receipts.",
		"Use this example as a starting point to create your own custom Substreams, which indexes the data you need from the NEAR blockchain.",
		New,
		60,
		"NEAR",
	)
}

type Convo struct {
	*codegen.BaseBlockchainGeneratorConversation[*Project]
}

func New() codegen.Converser {
	return &Convo{codegen.NewBaseBlockchainGeneratorConversation(&Project{
		codegen.BaseConversationState{
			ChainName: "near",
		},
	}, codegen.SharedFlowConfig{
		ValidChains: chains.NearNetworks(),
	})}
}
