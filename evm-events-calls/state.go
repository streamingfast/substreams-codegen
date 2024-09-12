package evm_events_calls

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/codemodus/kace"
	"github.com/golang-cz/textcase"
	"github.com/huandu/xstrings"
	"github.com/streamingfast/eth-go"
)

type Project struct {
	Name                   string             `json:"name"`
	ChainName              string             `json:"chainName"`
	Contracts              []*Contract        `json:"contracts"`
	DynamicContracts       []*DynamicContract `json:"dynamic_contracts"`
	Compile                bool               `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download               bool               `json:"download,omitempty"`
	ConfirmEnoughContracts bool               `json:"confirm_enough_contracts,omitempty"`

	// Remote build part removed for the moment
	// confirmDownloadOnly bool
	// confirmDoCompile    bool

	currentContractIdx     int
	compilingBuild         bool
	generatedCodeCompleted bool

	buildStarted    time.Time
	forceGeneration bool
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

func (p *Project) ChainConfig() *ChainConfig { return ChainConfigByID[p.ChainName] }
func (p *Project) ChainEndpoint() string     { return ChainConfigByID[p.ChainName].FirehoseEndpoint }

func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string  { return strings.ReplaceAll(p.Name, "_", "-") }

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
	Name        string          `json:"name,omitempty"`
	TrackEvents bool            `json:"trackEvents"`
	TrackCalls  bool            `json:"trackCalls"`
	RawABI      json.RawMessage `json:"rawAbi,omitempty"`

	abiFetchedInThisSession bool
	Abi                     *ABI
	emptyABI                bool
}

func (c *BaseContract) Identifier() string { return c.Name }
func (c *BaseContract) IdentifierSnakeCase() string {
	return xstrings.ToSnakeCase(c.Name)
}
func (c *BaseContract) IdentifierPascalCase() string { return textcase.PascalCase(c.Name) }
func (c *BaseContract) IdentityCamelCase() string    { return textcase.CamelCase(c.Name) }
func (c *BaseContract) IdentifierUpper() string      { return strings.ToUpper(c.Name) }

func (c *BaseContract) EventFields(event string) ([]*eth.LogParameter, error) {
	hash, err := hex.DecodeString(event)
	if err != nil {
		return nil, fmt.Errorf("invalid event ID %q: %w", event, err)
	}
	eventDef := c.Abi.abi.FindLogByTopic(hash)
	if eventDef == nil {
		return nil, fmt.Errorf("cannot find event definition for %q", event)
	}
	return eventDef.Parameters, nil
}

func (c *BaseContract) CallModels() []codegenCall {
	calls, err := c.Abi.BuildCallModels()
	if err != nil {
		panic(err)
	}
	return calls
}

func (c *BaseContract) EventModels() []codegenEvent {
	evts, err := c.Abi.BuildEventModels()
	if err != nil {
		panic(err)
	}
	return evts
}

type Contract struct {
	BaseContract
	Address      string  `json:"address"`
	InitialBlock *uint64 `json:"initialBlock"` // for each Contract, so we discover the lowest

	TrackFactory                 *bool  `json:"trackFactory"`
	FactoryCreationEvent         string `json:"factoryCreationEvent"`
	FactoryCreationEventFieldIdx *int64 `json:"factoryCreationEventFieldIdx"`
}

func (c *Contract) PlainAddress() string { return strings.TrimPrefix(c.Address, "0x") }

func (c *Contract) FactoryCreationEventName() string {
	for _, ev := range c.EventModels() {
		if ev.Proto.MessageHash == c.FactoryCreationEvent {
			return ev.Proto.MessageName
		}
	}
	panic("not found")
}

func (c *Contract) FactoryCreationEventFieldName() string {
	for _, ev := range c.EventModels() {
		if ev.Proto.MessageHash == c.FactoryCreationEvent {
			return ev.Proto.Fields[int(*c.FactoryCreationEventFieldIdx)].Name
		}
	}
	panic("not found")
}

func (c *Contract) FetchABI(chainConfig *ChainConfig) (abi string, err error) {
	a, err := getContractABIFollowingProxy(context.Background(), c.Address, chainConfig)
	if err != nil {
		return "", err
	}
	return a.raw, nil
}

func (c *Contract) FetchInitialBlock(chainConfig *ChainConfig) (initialBlock uint64, err error) {
	return getContractInitialBlock(context.Background(), chainConfig, c.Address)
}

// That's a contract that is _created by a Factory_. It doesn't have a start block because it
// is dynamically created at some future blocks, based on its parent Factory contract, tracked
// in a "Contract" above.
type DynamicContract struct {
	BaseContract
	ParentContractName string `json:"parentContractName"`

	parentContract           *Contract
	referenceContractAddress string
}

func (d DynamicContract) FactoryInitialBlock() uint64 {
	return *d.parentContract.InitialBlock
}

func (d DynamicContract) ParentContract() *Contract   { return d.parentContract }
func (d DynamicContract) Identifier() string          { return d.Name }
func (d DynamicContract) IdentifierSnakeCase() string { return kace.Snake(d.Name) }
func (d DynamicContract) FetchABI(chainConfig *ChainConfig) (abi string, err error) {
	a, err := getContractABIFollowingProxy(context.Background(), d.referenceContractAddress, chainConfig)
	if err != nil {
		return "", err
	}
	return a.raw, nil
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
		if dynamicContract.referenceContractAddress == address {
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

		if _, found := uniqueContractAddresses[dynamicContract.referenceContractAddress]; found {
			return fmt.Errorf("contract address %s already exists in the project", dynamicContract.referenceContractAddress)
		}

		uniqueContractNames[dynamicContract.Name] = struct{}{}
		uniqueContractAddresses[dynamicContract.referenceContractAddress] = struct{}{}
	}

	return nil
}
