package starknet_events

import (
	"fmt"
	"strings"
)

type Project struct {
	Name                 string    `json:"name"`
	ChainName            string    `json:"chainName"`
	Compile              bool      `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download             bool      `json:"download,omitempty"`
	InitialBlock         uint64    `json:"initialBlock,omitempty"`
	InitialBlockSet      bool      `json:"initialBlockSet,omitempty"`
	Contract             *Contract `json:"contract"`
	EventsTrackCompleted bool
}

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

func (p *Project) ChainConfig() *ChainConfig          { return ChainConfigByID[p.ChainName] }
func (p *Project) ChainNetwork() string               { return ChainConfigByID[p.ChainName].Network }
func (p *Project) IsValidChainName(input string) bool { return ChainConfigByID[input] != nil }
func (p *Project) IsTestnet(input string) bool {
	return ChainConfigByID[input].Network == "starknet-testnet"
}

type Contract struct {
	Address       string   `json:"address"`
	InitialBlock  *uint64  `json:"initialBlock"`
	TrackedEvents []string `json:"trackedEvents"`
}

func (p *Project) GetEventsQuery() string {
	query := fmt.Sprintf("tx:%s", p.Contract.Address)
	for _, event := range p.Contract.TrackedEvents {
		query = query + fmt.Sprintf(" && ev:%s", event)
	}
	return query
}
