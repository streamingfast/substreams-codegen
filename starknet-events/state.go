package starknet_events

import (
	"embed"
	"fmt"
	"regexp"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/base"
)

//go:embed templates/*
var templatesFS embed.FS

type Project struct {
	base.ConversationState
	Contracts              []*Contract `json:"contracts"`
	ConfirmEnoughContracts bool        `json:"confirmEnoughContracts,omitempty"`

	currentContractIdx int
}

func (p *Project) Generate() codegen.ReturnGenerate {
	res := codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/events.proto.gotmpl":     "proto/events.proto",
		"src/abi/mod.rs.gotmpl":         "src/abi/mod.rs",
		"src/lib.rs.gotmpl":             "src/lib.rs",
		"build.rs.gotmpl":               "build.rs",
		"Cargo.toml.gotmpl":             "Cargo.toml",
		"rust-toolchain.toml":           "rust-toolchain.toml",
		".gitignore.gotmpl":             ".gitignore",
		"substreams.yaml.gotmpl":        "substreams.yaml",
		"README.md.gotmpl":              "README.md",
		"common-templates/buf.gen.yaml": "buf.gen.yaml",
	})
	if res.Err != nil {
		return res
	}

	for _, contract := range p.Contracts {
		res.ProjectFiles[fmt.Sprintf("abi/%s_contract.abi.json", contract.Name)] = codegen.PrettifyJSON([]byte(contract.Abi.raw))
	}

	return res
}

func (p *Project) GetRPCProviderEndpointVar() string {
	network := p.FindNetworkFromChainName()
	if network == nil {
		return ""
	}

	switch network.ID {
	case "starknet-mainnet":
		return "STARKNET_MAINNET_ENDPOINT"
	case "starknet-testnet":
		return "STARKNET_TESTNET_ENDPOINT"
	}

	return ""
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

var contractNameRegexp = regexp.MustCompile(`^([a-z][a-z0-9_]{0,63})$`)

func validateContractName(p *Project, name string) error {
	if !contractNameRegexp.MatchString(name) {
		return fmt.Errorf("contract name %s is invalid, it must match the regex ^([a-z][a-z0-9_]{0,63})$", name)
	}

	for _, contract := range p.Contracts {
		if contract.Name == name {
			return fmt.Errorf("contract with name %s already exists in the project", name)
		}
	}

	return nil
}
