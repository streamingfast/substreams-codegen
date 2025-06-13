package solanchor

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test unmarshalDefined
func TestUnmarshalDefinedStandard(t *testing.T) {
	json := []byte(`{"defined":"MyType"}`)

	result := unmarshalDefined(json)

	assert.Equal(t, "MyType", result)
}

func TestUnmarshalDefinedName(t *testing.T) {
	json := []byte(`{"defined":{"name": "MyType"}}`)

	result := unmarshalDefined(json)

	assert.Equal(t, "MyType", result)
}

func TestPrintDefined(t *testing.T) {
	// Generate types from IDL
	idlString := readFromFile("orca")

	idl := &IDL{}
	err := json.Unmarshal(idlString, &idl)

	jsonBytes := []byte(`{"defined":{"name": "OpenPositionWithMetadataBumps"}}`)

	result, err := unmarshallFieldType(jsonBytes)
	assert.Nil(t, err)

	resolvedFieldType, err := result.GetResolvedFieldType()
	assert.Nil(t, err)

	mappings := resolvedFieldType.PrintRustMappings("my_type", "inst", idl.Types)
	fmt.Println(mappings)
}

func TestPrintDefinedWithNestedDefined(t *testing.T) {
	// Generate types from IDL
	idlString := readFromFile("orca")

	idl := &IDL{}
	err := json.Unmarshal(idlString, &idl)
	assert.Nil(t, err)

	jsonBytes := []byte(`{"defined":{"name": "RemainingAccountsSlice"}}`)

	result, err := unmarshallFieldType(jsonBytes)
	assert.Nil(t, err)

	resolvedFieldType, err := result.GetResolvedFieldType()
	assert.Nil(t, err)

	mappings := resolvedFieldType.PrintRustMappings("my_type", "inst", idl.Types)
	fmt.Println(mappings)
}

func TestPrintDefinedRustStructures(t *testing.T) {
	// Generate types from IDL
	idlString := readFromFile("orca")

	idl := &IDL{}
	err := json.Unmarshal(idlString, &idl)
	assert.Nil(t, err)

	jsonBytes := []byte(`{"defined":{"name": "WhirlpoolRewardInfo"}}`)

	result, err := unmarshallFieldType(jsonBytes)
	assert.Nil(t, err)

	resolvedFieldType, err := result.GetResolvedFieldType()
	assert.Nil(t, err)

	mappings := resolvedFieldType.PrintNecessaryRustStructs("my_type", idl.Types)
	fmt.Println(mappings)
}

// vec
func TestPrintVecDefined(t *testing.T) {
	// Generate types from IDL
	idlString := readFromFile("orca")

	idl := &IDL{}
	err := json.Unmarshal(idlString, &idl)
	assert.Nil(t, err)

	jsonBytes := []byte(`{"vec": {"defined":{"name": "RemainingAccountsSlice"}}}`)

	result, err := unmarshallFieldType(jsonBytes)
	assert.Nil(t, err)

	resolvedFieldType, err := result.GetResolvedFieldType()
	assert.Nil(t, err)

	mappings := resolvedFieldType.PrintRustMappings("my_type", "inst", idl.Types)
	fmt.Println(mappings)
}

// Test unmarshalSimple
func TestUnmarshalSimple(t *testing.T) {
	json := []byte(`"string"`)

	result := unmarshalSimple(json)

	assert.Equal(t, "string", result)
}

// ------------ OPTION
// vec
func TestPrintOptionSimple(t *testing.T) {
	// Generate types from IDL
	jsonBytes := []byte(`{"option": "u64"}`)

	result, err := unmarshallFieldType(jsonBytes)
	assert.Nil(t, err)

	resolvedFieldType, err := result.GetResolvedFieldType()
	assert.Nil(t, err)

	assert.Equal(t, "Option<u64>", resolvedFieldType.ResolveRustType())
}
func TestPrintOptionDefined(t *testing.T) {
	// Generate types from IDL
	idlString := readFromFile("orca")

	idl := &IDL{}
	err := json.Unmarshal(idlString, &idl)
	assert.Nil(t, err)

	jsonBytes := []byte(`{"option": {"defined": {"name": "RemainingAccountsSlice"}}}`)

	result, err := unmarshallFieldType(jsonBytes)
	assert.Nil(t, err)

	resolvedFieldType, err := result.GetResolvedFieldType()
	assert.Nil(t, err)

	mappings := resolvedFieldType.PrintRustMappings("my_type", "inst", idl.Types)
	fmt.Println(mappings)
	fmt.Println("-------------------------")
	fmt.Println(resolvedFieldType.PrintNecessaryRustStructs("idl", idl.Types))
}

func TestUnmarshallRecursivelyOptionDefined(t *testing.T) {
	json := []byte(`{"option": {"defined": "MyObject"}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	assert.Nil(t, err)
	assert.True(t, optionType.IsOptionRecursive())
	assert.True(t, optionType.OptionRecursive.Type.IsVecRecursive())
	assert.True(t, optionType.OptionRecursive.Type.VecRecursive.Type.IsDefined())
	assert.Equal(t, "MyObject", optionType.OptionRecursive.Type.VecRecursive.Type.Defined.Type)
	assert.Equal(t, "Option<Vec<MyObject>>", resolvedType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionVecSimple(t *testing.T) {
	json := []byte(`{"option": {"vec": "u8"}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	assert.Nil(t, err)
	assert.Equal(t, "Option<Vec<u64>>", resolvedType.ResolveRustType())

	fmt.Println(resolvedType.PrintNecessaryRustStructs("inst", []Type{}))
}

func TestUnmarshallRecursivelyOptionVecDefined(t *testing.T) {
	json := []byte(`{"option": {"vec": {"defined": "MyObject"}}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	assert.Nil(t, err)
	assert.True(t, optionType.IsOptionRecursive())
	assert.True(t, optionType.OptionRecursive.Type.IsVecRecursive())
	assert.True(t, optionType.OptionRecursive.Type.VecRecursive.Type.IsDefined())
	assert.Equal(t, "MyObject", optionType.OptionRecursive.Type.VecRecursive.Type.Defined.Type)
	assert.Equal(t, "Option<Vec<MyObject>>", resolvedType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionWithArray(t *testing.T) {
	json := []byte(`{"option": {"array": [{"defined": {"name": "EmodeEntry"}},10]}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	assert.Nil(t, err)
	assert.True(t, optionType.IsOptionRecursive())
	assert.True(t, optionType.OptionRecursive.Type.IsArrayRecursive())
	assert.True(t, optionType.OptionRecursive.Type.ArrayRecursive.Type.IsDefined())
	assert.Equal(t, "EmodeEntry", optionType.OptionRecursive.Type.ArrayRecursive.Type.Defined.Type)
	assert.Equal(t, "Option<[EmodeEntry;10]>", resolvedType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionWith2DArraySimple(t *testing.T) {
	json := []byte(`{"option": {"array": [{"array": ["u64", 3]},10]}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	fmt.Println(resolvedType.PrintNecessaryProtobufMessages())
	assert.Nil(t, err)
	assert.True(t, optionType.IsOptionRecursive())
	assert.True(t, optionType.OptionRecursive.Type.IsArrayRecursive())
	assert.True(t, optionType.OptionRecursive.Type.ArrayRecursive.Type.IsArrayRecursive())
	assert.Equal(t, "u64", optionType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Type.Simple.Type)
	assert.Equal(t, 3, optionType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Length)
	assert.Equal(t, "Option<[[u64;3];10]>", resolvedType.ResolveRustType())

	fmt.Println(resolvedType.PrintNecessaryRustStructs("inst", []Type{}))
}

func TestUnmarshallRecursivelyOptionWith2DArrayDefined(t *testing.T) {
	json := []byte(`{"option": {"array": [{"array": [{"defined": "MyObject"}, 3]},10]}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	assert.Nil(t, err)
	assert.True(t, optionType.IsOptionRecursive())
	assert.True(t, optionType.OptionRecursive.Type.IsArrayRecursive())
	assert.True(t, optionType.OptionRecursive.Type.ArrayRecursive.Type.IsArrayRecursive())
	assert.Equal(t, "MyObject", optionType.OptionRecursive.Type.ArrayRecursive.Type.ArrayRecursive.Type.Defined)
	assert.Equal(t, "Option<[[MyObject;3];10]>", resolvedType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionWithVec(t *testing.T) {
	json := []byte(`{"option": {"vec": "MyObject"}}`)

	optionType, err := unmarshallFieldType(json)
	resolvedType, err := optionType.GetResolvedFieldType()

	fmt.Println(resolvedType.PrintNecessaryProtobufMessages())
	assert.Nil(t, err)
	assert.Equal(t, "Option<Vec<MyObject>>", resolvedType.ResolveRustType())
}

// ------------------ ARRAY
func TestUnmarshallRecursivelyArray(t *testing.T) {
	json := []byte(`{"array": [{"defined": {"name": "EmodeEntry"}},10]}`)

	array, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, array.IsArrayRecursive())
	assert.Equal(t, 10, array.ArrayRecursive.Length)
	assert.True(t, array.ArrayRecursive.Type.IsDefined())
	assert.Equal(t, "EmodeEntry", array.ArrayRecursive.Type.Defined.Type)
}

func TestUnmarshallRecursivelyArrayInsideArray(t *testing.T) {
	json := []byte(`{"array": [{"array": [{"defined": "MyObject"}, 3]},10]}`)

	array, err := unmarshallFieldType(json)
	resolvedType, err := array.GetResolvedFieldType()

	assert.Nil(t, err)
	assert.True(t, array.IsArrayRecursive())
	assert.Equal(t, 10, array.ArrayRecursive.Length)
	assert.True(t, array.ArrayRecursive.Type.IsArrayRecursive())
	assert.Equal(t, "MyObject", array.ArrayRecursive.Type.ArrayRecursive.Type.Defined.Type)
	assert.Equal(t, "[[MyObject;3];10]", resolvedType.ResolveRustType())
	assert.Equal(t, 3, array.ArrayRecursive.Type.ArrayRecursive.Length)
}
