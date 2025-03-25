package stellartransactionsoperations

import (
	"fmt"
	"strings"
)

type Project struct {
	Name       string `json:"name"`
	ChainName  string `json:"chainName"`
	Compile    bool   `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download   bool   `json:"download,omitempty"`
	FilterType string `json:"filterType,omitempty"`
	Filter     string `json:"filter,omitempty"`
}

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

func (p *Project) ChainConfig() *ChainConfig          { return ChainConfigByID[p.ChainName] }
func (p *Project) ChainNetwork() string               { return ChainConfigByID[p.ChainName].Network }
func (p *Project) IsValidChainName(input string) bool { return ChainConfigByID[input] != nil }

func (p *Project) ComposeFilter() string {
	keyword := "source_account"
	if p.FilterType == "operations" {
		keyword = "operation"
	}

	var stringBuilder string

	splittedFilter := strings.Split(p.Filter, ",")
	for i, filterPiece := range splittedFilter {
		stringBuilder += fmt.Sprintf("%s:%s", keyword, filterPiece)

		if i < (len(splittedFilter) - 1) {
			stringBuilder += " || "
		}
	}

	return stringBuilder
}
