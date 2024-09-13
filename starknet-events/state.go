package starknet_events

import (
	"fmt"
	"regexp"
	"strings"
)

type Project struct {
	Name                   string      `json:"name"`
	ChainName              string      `json:"chainName"`
	Contracts              []*Contract `json:"contracts"`
	ConfirmEnoughContracts bool        `json:"confirmEnoughContracts,omitempty"`

	currentContractIdx     int
	generatedCodeCompleted bool
	projectFiles           map[string][]byte
}

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

func (p *Project) ChainConfig() *ChainConfig          { return ChainConfigByID[p.ChainName] }
func (p *Project) ChainNetwork() string               { return ChainConfigByID[p.ChainName].Network }
func (p *Project) IsValidChainName(input string) bool { return ChainConfigByID[input] != nil }
func (p *Project) IsTestnet(input string) bool {
	return ChainConfigByID[input].Network == "starknet-testnet"
}

func (p *Project) GetEventsQuery() string {
	var query string
	for i, contract := range p.Contracts {
		if i == 0 {
			query = fmt.Sprintf("ev:from_address:%s", contract.Address)
			continue
		}
		query = query + fmt.Sprintf(" || ev:from_address:%s", contract.Address)
	}

	return query
}

func contractNames(contracts []*Contract) (out []string) {
	for _, contract := range contracts {
		out = append(out, contract.Name)
	}
	return
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
}

func validateContractName(p *Project, name string) error {
	if !regexp.MustCompile(`^([a-z][a-z0-9_]{0,63})$`).MatchString(name) {
		return fmt.Errorf("contract name %s is invalid, it must match the regex ^([a-z][a-z0-9_]{0,63})$", name)
	}

	for _, contract := range p.Contracts {
		if contract.Name == name {
			return fmt.Errorf("contract with name %s already exists in the project", name)
		}
	}

	return nil
}
