package starknet_events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NethermindEth/juno/core/felt"

	starknetRPC "github.com/NethermindEth/starknet.go/rpc"
)

type Alias struct {
	OldName string
	NewName string
}

type Contract struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address"`

	InitialBlock *uint64         `json:"initialBlock"`
	Aliases      []Alias         `json:"aliases"`
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
	events := c.Abi.decodedAbi.EventsBySelector

	aliases := make([]Alias, 0)
	seen := make(map[string]struct{})
	for _, eventItem := range events {

		eventName := eventItem.Name

		splitEventName := strings.Split(eventName, "::")

		lastPart := splitEventName[len(splitEventName)-1]

		if _, found := seen[lastPart]; found {
			if len(splitEventName) < 2 {
				panic("parsed event name does not contain enough parts to have an alias")
			}

			alias := Alias{
				OldName: eventName,
				NewName: splitEventName[len(splitEventName)-2] + lastPart,
			}

			aliases = append(aliases, alias)
		}

		seen[lastPart] = struct{}{}
	}

	c.Aliases = aliases
}

func (c *Contract) fetchABI(config *ChainConfig) (string, error) {
	client, err := starknetRPC.NewProvider(os.Getenv(config.EndpointEnvVar))
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

	return
}

func (c *Contract) AddressWithoutPrefix() string {
	return strings.TrimPrefix(c.Address, "0x")
}
