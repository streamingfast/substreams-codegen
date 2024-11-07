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

func readFromFile(idlName string) []byte {
	data, err := os.ReadFile("tests/" + idlName + ".json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return make([]byte, 0)
	}

	return data
}
