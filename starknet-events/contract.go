package starknet_events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NethermindEth/juno/core/felt"
	starknetRPC "github.com/NethermindEth/starknet.go/rpc"
	registry "github.com/pinax-network/graph-networks-libs/packages/golang/lib"
)

type Alias struct {
	OldName string
	NewName string
}

func NewAlias(oldName, newName string) *Alias {
	return &Alias{
		OldName: oldName,
		NewName: newName,
	}
}

type Contract struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`

	InitialBlock *uint64         `json:"initialBlock"`
	Aliases      []*Alias        `json:"aliases"`
	RawABI       json.RawMessage `json:"rawAbi,omitempty"`

	Abi                     *ABI
	emptyABI                bool
	abiFetchedInThisSession bool
}

func (c *Contract) Identifier() string { return c.Name }
func (c *Contract) IdentifierCapitalize() string {
	if len(c.Name) == 0 {
		return c.Name
	}

	if len(c.Name) == 1 {
		return strings.ToUpper(c.Name)
	}

	return strings.ToUpper(string(c.Name[0])) + c.Name[1:]
}
func (c *Contract) SetAliases() {
	c.setAliasesForEvents()
	c.setAliasesForOtherItems()
}

// This is a bit hacky, but it works for now!
func (c *Contract) setAliasesForOtherItems() {
	otherItems := c.Abi.otherItems

	seen := make(map[string]int)
	for _, item := range otherItems {
		splittedName := strings.Split(item.Name, "::")
		lastPart := splittedName[len(splittedName)-1]

		count, found := seen[lastPart]
		if !found {
			seen[lastPart] = 1
			continue
		}

		newName := lastPart + "V" + strconv.Itoa(count)
		alias := NewAlias(item.Name, newName)
		c.Aliases = append(c.Aliases, alias)
		seen[lastPart] = count + 1
	}
}
func (c *Contract) setAliasesForEvents() {
	events := c.Abi.decodedEvents

	aliases := make([]*Alias, 0)
	seen := make(map[string]struct{})

	// Based on Starknet documentation, we assume that in each contract, it exists a Event which is an enum containing all other events... (https://docs.starknet.io/architecture-and-concepts/smart-contracts/contract-abi/)
	// Finding this "golden" event is not an easy path, as multiple enum with the same name can exist in the ABI...
	// We need to detect the Golden Event to avoid applying Alias on it...
	potentialsGoldenEvent := make(map[string]*StarknetEvent)
	for _, event := range events {
		eventName := event.Name
		lastPart, newName := eventNameInfo(eventName)

		fmt.Println(lastPart, newName)

		if lastPart == "Event" {
			// Event which are not enum, we can safely apply alias
			if event.Kind != "enum" {
				alias := NewAlias(eventName, newName)
				aliases = append(aliases, alias)

				continue
			}

			potentialsGoldenEvent[event.Name] = event
			continue
		}

		if _, found := seen[lastPart]; found {
			alias := NewAlias(eventName, newName)
			aliases = append(aliases, alias)
		}

		seen[lastPart] = struct{}{}
	}

	if len(potentialsGoldenEvent) == 1 {
		c.Aliases = aliases
		return
	}

	goldenName := detectGoldenEvent(potentialsGoldenEvent)
	if goldenName == "" {
		panic("no golden event found")
	}

	aliases = setNonGoldenAliases(potentialsGoldenEvent, goldenName, aliases)
	c.Aliases = aliases
}

func (c *Contract) fetchABI(network *registry.Network, endpointVar string) (string, error) {
	client, err := starknetRPC.NewProvider(os.Getenv(endpointVar))
	if err != nil {
		return "", fmt.Errorf("creating rpc client: %w", err)
	}

	ctx := context.Background()

	blockId := starknetRPC.BlockID{
		Tag: "latest",
	}

	emptyField := felt.Felt{}
	addressToFelt, err := emptyField.SetString(c.AddressWithoutPrefix())
	if err != nil {
		return "", fmt.Errorf("converting address to felt: %w", err)
	}

	classOutput, err := client.ClassAt(ctx, blockId, addressToFelt)
	if err != nil {
		return "", fmt.Errorf("calling class at for adderss: %s : %w", c.AddressWithoutPrefix(), err)
	}

	var contractABI string
	switch classOutput.(type) {
	case *starknetRPC.ContractClass:
		contractClass := classOutput.(*starknetRPC.ContractClass)
		contractABI = contractClass.ABI
	case *starknetRPC.DeprecatedContractClass:
		return "", fmt.Errorf("deprecated contract class not supported")
	default:
		return "", fmt.Errorf("classoutput type not supported")
	}

	return contractABI, nil
}

// In some explorers (Ex: Starkscan) the address is padded on 66 characters with the prefix 0x
// The Contract is containing both, the padded address or the raw one without leading zeros...
func (c *Contract) handleContractAddress(inputAddress string) {
	// Address padded
	if len(inputAddress) == 66 {
		c.Address = inputAddress
		return
	}

	// Address not padded
	withoutPrefix := strings.TrimPrefix(inputAddress, "0x")
	c.Address = "0x" + strings.Repeat("0", 64-len(withoutPrefix)) + withoutPrefix
}

func (c *Contract) AddressWithoutPrefix() string {
	return strings.TrimPrefix(c.Address, "0x")
}

func setNonGoldenAliases(potentialsGoldenEvent map[string]*StarknetEvent, goldenName string, aliases []*Alias) []*Alias {
	for _, event := range potentialsGoldenEvent {
		eventName := event.Name

		if eventName == goldenName {
			continue
		}

		_, newName := eventNameInfo(eventName)

		alias := NewAlias(event.Name, newName)
		aliases = append(aliases, alias)
	}

	return aliases
}

func eventNameInfo(eventName string) (lastPart, aliasName string) {
	splitEventName := strings.Split(eventName, "::")
	if len(splitEventName) < 2 {
		panic("parsed event name does not contain enough parts to have an alias")
	}

	lastPart = splitEventName[len(splitEventName)-1]
	return lastPart, splitEventName[len(splitEventName)-2] + lastPart
}

func detectGoldenEvent(potentialsGoldenEvent map[string]*StarknetEvent) string {
	for _, event := range potentialsGoldenEvent {
		seen := make(map[string]struct{})
		for _, variant := range event.Variants {
			if _, found := potentialsGoldenEvent[variant.Type]; found {
				seen[variant.Type] = struct{}{}
			}

			// Equivalent: Current Event Enum contains all other potentials "golden" events
			if len(seen) == (len(potentialsGoldenEvent) - 1) {
				return event.Name
			}
		}
	}

	return ""
}
