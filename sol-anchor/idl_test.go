package solanchor

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPumpFunIDL(t *testing.T) {
	idl := readFromFile("pumpfun")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)
	fmt.Println(len(result.Accounts))
	assert.Nil(t, err)
	assert.Equal(t, "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P", result.Metadata.Address)
	assert.Equal(t, 6, len(result.Instructions))
	assert.Equal(t, 4, len(result.Events))
	assert.Equal(t, 0, len(result.Types))
}

func TestMeteoraIDL(t *testing.T) {
	idl := readFromFile("meteora")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	assert.Nil(t, err)
	assert.Equal(t, "24Uqj9JCLxUeoC3hGfh5W3s9FM9uCHDS2SG3LYwBpyTi", result.Metadata.Address)
	assert.Equal(t, 14, len(result.Instructions))
	assert.Equal(t, 8, len(result.Events))
	assert.Equal(t, 4, len(result.Types))
}

func TestOrcaIDL(t *testing.T) {
	idl := readFromFile("orca")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	assert.Nil(t, err)
	assert.Equal(t, "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc", result.Metadata.Address)
	assert.Equal(t, 46, len(result.Instructions))
	assert.Equal(t, 0, len(result.Events))
	assert.Equal(t, 12, len(result.Types))
}

func TestJupiterV4Swap(t *testing.T) {
	idl := readFromFile("jupiter_v4_swap")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	assert.Nil(t, err)
	assert.Equal(t, "JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB", result.Metadata.Address)
	assert.True(t, result.IsTypeEnum("SwapLeg"))
}

func TestOrbitLen(t *testing.T) {
	idl := readFromFile("orbit_len")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	fmt.Println(result.IsTypeUsed("LendingAccount"))

	assert.Nil(t, err)
	assert.Equal(t, "QoB7dVkkZr3oLb95DMpSptvUF8mTygDHNjFQh5y5RAb", result.Address)

}

func TestIthaca(t *testing.T) {
	idl := readFromFile("ithaca")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	assert.Nil(t, err)
}

func TestRaydiumAMM(t *testing.T) {
	idl := readFromFile("raydium_amm")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	for _, i := range result.Instructions {
		if i.Name == "simulateInfo" {
			for _, a := range i.Args {
				if a.Type.IsOption() && a.Type.Option.Defined != nil {
					fmt.Println(*a.Type.Option.Defined)
				}

				fmt.Printf("%s\n", a.Name)
			}
		}
	}

	fmt.Println(PrintDefined("SwapInstructionBaseIn", "instruction", "instruction", result.Types, false, true))

	assert.Nil(t, err)
}


func TestProofOfPlay(t *testing.T) {
	idl := readFromFile("proof_of_play")

	result := &IDL{}
	err := json.Unmarshal(idl, &result)

	assert.Nil(t, err)
}


func readFromFile(idlName string) []byte {
	data, err := os.ReadFile("idls/" + idlName + ".json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return make([]byte, 0)
	}

	return data
}
