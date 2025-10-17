package evm_events_calls_raw

import (
	"embed"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/codemodus/kace"
	"github.com/golang-cz/textcase"
	"github.com/huandu/xstrings"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/base"
	evm_events_calls "github.com/streamingfast/substreams-codegen/evm-events-calls"
)

//go:embed templates/*
var templatesFS embed.FS

type Project struct {
	base.ConversationState
	Contracts              []*Contract        `json:"contracts"`
	DynamicContracts       []*DynamicContract `json:"dynamic_contracts"`
	Compile                bool               `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download               bool               `json:"download,omitempty"`
	ConfirmEnoughContracts bool               `json:"confirm_enough_contracts,omitempty"`

	currentContractIdx int
}

func (p *Project) Generate() codegen.ReturnGenerate {
	res := codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/contract.proto.gotmpl":   "proto/contract.proto",
		"src/pb/mod.rs.gotmpl":          "src/pb/mod.rs",
		"src/lib.rs.gotmpl":             "src/lib.rs",
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

	return res
}

func dynamicContractNames(contracts []*DynamicContract) (out []string) {
	for _, contract := range contracts {
		out = append(out, contract.Name)
	}
	return
}

func contractNames(contracts []*Contract) (out []string) {
	for _, contract := range contracts {
		out = append(out, contract.Name)
	}
	return
}

func (p *Project) ChainConfig() *evm_events_calls.ChainConfig { return ChainConfigByID[p.ChainName] }

func (p *Project) GetContractByName(contractName string) *Contract {
	for _, contract := range p.Contracts {
		if contract.Name == contractName {
			return contract
		}
	}
	return nil
}

func (p *Project) dynamicContractOf(contractName string) (out *DynamicContract) {
	for _, dynContract := range p.DynamicContracts {
		if dynContract.ParentContractName == contractName {
			out = dynContract
			break
		}
	}
	if out == nil {
		out = &DynamicContract{
			ParentContractName: contractName,
		}
		p.DynamicContracts = append(p.DynamicContracts, out)
	}
	return
}

func isValidChainName(input string) bool {
	return ChainConfigByID[input] != nil
}

func (p *Project) ApplyEventsBlockFilter() bool {
	for _, dcontract := range p.DynamicContracts {
		if dcontract.TrackEvents {
			return false
		}
	}

	for _, contract := range p.Contracts {
		if contract.TrackEvents {
			return true
		}
	}

	return false
}

func (p *Project) ApplyCallsBlockFilter() bool {
	for _, dcontract := range p.DynamicContracts {
		if dcontract.TrackCalls {
			return false
		}
	}

	for _, contract := range p.Contracts {
		if contract.TrackCalls {
			return true
		}
	}

	return false
}

func (p *Project) GenerateEventsBlockFilterQuery() string {
	var query string
	for _, contract := range p.Contracts {
		if !contract.TrackEvents {
			continue
		}

		if query == "" {
			query = fmt.Sprintf("evt_addr:%s", strings.ToLower(contract.Address))
			continue
		}

		query += fmt.Sprintf(" || evt_addr:%s", strings.ToLower(contract.Address))
	}

	return query
}

func (p *Project) GenerateCallsBlockFilterQuery() string {
	var query string
	for _, contract := range p.Contracts {
		if !contract.TrackCalls {
			continue
		}

		if query == "" {
			query = fmt.Sprintf("call_to:%s", strings.ToLower(contract.Address))
			continue
		}

		query += fmt.Sprintf(" || call_to:%s", strings.ToLower(contract.Address))
	}

	return query
}

func (p *Project) TrackAnyCalls() bool {
	for _, contract := range p.Contracts {
		if contract.TrackCalls {
			return true
		}
	}

	for _, dynamicContract := range p.DynamicContracts {
		if dynamicContract.TrackCalls {
			return true
		}
	}

	return false
}

func (p *Project) TrackAnyEvents() bool {
	for _, contract := range p.Contracts {
		if contract.TrackEvents {
			return true
		}
	}

	for _, dynamicContract := range p.DynamicContracts {
		if dynamicContract.TrackEvents {
			return true
		}
	}

	return false
}

func (p *Project) TrackOnlyCalls() bool {
	for _, contract := range p.Contracts {
		if contract.TrackEvents {
			return false
		}
	}

	for _, dynamicContract := range p.DynamicContracts {
		if dynamicContract.TrackEvents {
			return false
		}
	}

	return true
}

func (p *Project) TrackOnlyEvents() bool {
	for _, contract := range p.Contracts {
		if contract.TrackCalls {
			return false
		}
	}

	for _, dynamicContract := range p.DynamicContracts {
		if dynamicContract.TrackCalls {
			return false
		}
	}

	return true
}

func (p *Project) MustLowestStartBlock() (out uint64) {
	out = math.MaxUint64
	for _, contract := range p.Contracts {
		out = min(out, *contract.InitialBlock)
	}
	return
}

// was .hasDDS
func (p *Project) HasFactoryTrackers() bool {
	for _, contract := range p.Contracts {
		if *contract.TrackFactory {
			return true
		}
	}
	return false
}

func (p *Project) AllContracts() []*BaseContract {
	out := make([]*BaseContract, len(p.Contracts)+len(p.DynamicContracts))
	for i, contract := range p.Contracts {
		out[i] = &contract.BaseContract
	}

	offset := len(p.Contracts)
	for i, dynamicContract := range p.DynamicContracts {
		out[i+offset] = &dynamicContract.BaseContract
	}

	return out
}

type BaseContract struct {
	Name        string `json:"name,omitempty"`
	TrackEvents bool   `json:"trackEvents"`
	TrackCalls  bool   `json:"trackCalls"`
}

func (c *BaseContract) Identifier() string { return c.Name }
func (c *BaseContract) IdentifierSnakeCase() string {
	return xstrings.ToSnakeCase(c.Name)
}
func (c *BaseContract) IdentifierPascalCase() string { return textcase.PascalCase(c.Name) }
func (c *BaseContract) IdentityCamelCase() string    { return textcase.CamelCase(c.Name) }
func (c *BaseContract) IdentifierUpper() string      { return strings.ToUpper(c.Name) }

type Contract struct {
	BaseContract
	Address      string  `json:"address"`
	InitialBlock *uint64 `json:"initialBlock"` // for each Contract, so we discover the lowest

	TrackFactory                 *bool  `json:"trackFactory"`
	FactoryCreationEvent         string `json:"factoryCreationEvent"`
	FactoryCreationEventFieldIdx *int64 `json:"factoryCreationEventFieldIdx"`
}

func (c *Contract) PlainAddress() string { return strings.TrimPrefix(c.Address, "0x") }

// That's a contract that is _created by a Factory_. It doesn't have a start block because it
// is dynamically created at some future blocks, based on its parent Factory contract, tracked
// in a "Contract" above.
type DynamicContract struct {
	BaseContract
	ParentContractName string `json:"parentContractName"`

	parentContract           *Contract
	ReferenceContractAddress string `json:"referenceContractAddress"`
}

func (d DynamicContract) FactoryInitialBlock() uint64 {
	return *d.parentContract.InitialBlock
}
func (d DynamicContract) GenerateStoreQuery() string {
	return fmt.Sprintf("evt_addr:%s && evt_sig:%s", strings.ToLower(d.parentContract.Address), "0x"+d.parentContract.FactoryCreationEvent)
}
func (d DynamicContract) ParentContract() *Contract   { return d.parentContract }
func (d DynamicContract) Identifier() string          { return d.Name }
func (d DynamicContract) IdentifierSnakeCase() string { return kace.Snake(d.Name) }

func validateContractName(p *Project, name string) error {
	if !regexp.MustCompile(`^([a-z][a-z0-9_]{0,63})$`).MatchString(name) {
		return fmt.Errorf("contract name %s is invalid, it must match the regex ^([a-z][a-z0-9_]{0,63})$", name)
	}

	for _, contract := range p.Contracts {
		if contract.Name == name {
			return fmt.Errorf("contract with name %s already exists in the project", name)
		}
	}

	for _, dynamiContract := range p.DynamicContracts {
		if dynamiContract.Name == name {
			return fmt.Errorf("contract with name %s already exists in the project", name)
		}
	}
	return nil
}

func validateContractAddress(p *Project, address string) error {
	if !strings.HasPrefix(address, "0x") && len(address) == 42 {
		return fmt.Errorf("contract address %s is invalid, it must be a 42 character hex string starting with 0x", address)
	}

	for _, contract := range p.Contracts {
		if contract.Address == address {
			return fmt.Errorf("contract address %s already exists in the project", address)
		}
	}

	for _, dynamicContract := range p.DynamicContracts {
		if dynamicContract.ReferenceContractAddress == address {
			return fmt.Errorf("contract address %s already exists in the project", address)
		}
	}

	return nil
}

func validateIncomingState(p *Project) error {
	uniqueContractNames := map[string]struct{}{}
	uniqueContractAddresses := map[string]struct{}{}

	for _, contract := range p.Contracts {
		if _, found := uniqueContractNames[contract.Name]; found {
			return fmt.Errorf("contract with name %s already exists in the project", contract.Name)
		}

		if _, found := uniqueContractAddresses[contract.Address]; found {
			return fmt.Errorf("contract address %s already exists in the project", contract.Address)
		}

		uniqueContractNames[contract.Name] = struct{}{}
		uniqueContractAddresses[contract.Address] = struct{}{}
	}

	for _, dynamicContract := range p.DynamicContracts {
		if _, found := uniqueContractNames[dynamicContract.Name]; found {
			return fmt.Errorf("contract with name %s already exists in the project", dynamicContract.Name)
		}

		if _, found := uniqueContractAddresses[dynamicContract.ReferenceContractAddress]; found {
			return fmt.Errorf("contract address %s already exists in the project", dynamicContract.ReferenceContractAddress)
		}

		uniqueContractNames[dynamicContract.Name] = struct{}{}
		uniqueContractAddresses[dynamicContract.ReferenceContractAddress] = struct{}{}
	}

	return nil
}
