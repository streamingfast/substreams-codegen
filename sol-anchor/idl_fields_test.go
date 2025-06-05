package solanchor

import (
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

// Test unmarshalSimple
func TestUnmarshalSimple(t *testing.T) {
	json := []byte(`"string"`)

	result := unmarshalSimple(json)

	assert.Equal(t, "string", result)
}

// ------------ OPTION
func TestUnmarshallRecursivelyOption(t *testing.T) {
	json := []byte(`{"option": {"vec": {"defined": "MyObject"}}}`)

	optionType, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, optionType.IsOption())
	assert.True(t, optionType.Option.Type.IsVec())
	assert.True(t, optionType.Option.Type.Vec.Type.IsDefined())
	assert.Equal(t, "MyObject", optionType.Option.Type.Vec.Type.Defined)
	assert.Equal(t, "Option<Vec<MyObject>>", optionType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionWithArray(t *testing.T) {
	json := []byte(`{"option": {"array": [{"defined": {"name": "EmodeEntry"}},10]}}`)

	optionType, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, optionType.IsOption())
	assert.True(t, optionType.Option.Type.IsArray())
	assert.True(t, optionType.Option.Type.Array.Type.IsDefined())
	assert.Equal(t, "EmodeEntry", optionType.Option.Type.Array.Type.Defined)
	assert.Equal(t, "Option<[EmodeEntry;10]>", optionType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionWith2DArraySimple(t *testing.T) {
	json := []byte(`{"option": {"array": [{"array": ["u64", 3]},10]}}`)

	optionType, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, optionType.IsOption())
	assert.True(t, optionType.Option.Type.IsArray())
	assert.True(t, optionType.Option.Type.Array.Type.IsArray())
	assert.Equal(t, "u64", optionType.Option.Type.Array.Type.Array.Type.Simple)
	assert.Equal(t, "Option<[[u64;3];10]>", optionType.ResolveRustType())
}

func TestUnmarshallRecursivelyOptionWith2DArrayDefined(t *testing.T) {
	json := []byte(`{"option": {"array": [{"array": [{"defined": "MyObject"}, 3]},10]}}`)

	optionType, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, optionType.IsOption())
	assert.True(t, optionType.Option.Type.IsArray())
	assert.True(t, optionType.Option.Type.Array.Type.IsArray())
	assert.Equal(t, "MyObject", optionType.Option.Type.Array.Type.Array.Type.Defined)
	assert.Equal(t, "Option<[[MyObject;3];10]>", optionType.ResolveRustType())
}

// ------------------ ARRAY
func TestUnmarshallRecursivelyArray(t *testing.T) {
	json := []byte(`{"array": [{"defined": {"name": "EmodeEntry"}},10]}`)

	array, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, array.IsArray())
	assert.Equal(t, 10, array.Array.Length)
	assert.True(t, array.Array.Type.IsDefined())
	assert.Equal(t, "EmodeEntry", array.Array.Type.Defined)
}

func TestUnmarshallRecursivelyArrayInsideArray(t *testing.T) {
	json := []byte(`{"array": [{"array": [{"defined": "MyObject"}, 3]},10]}`)

	array, err := unmarshallFieldType(json)

	assert.Nil(t, err)
	assert.True(t, array.IsArray())
	assert.Equal(t, 10, array.Array.Length)
	assert.True(t, array.Array.Type.IsArray())
	assert.Equal(t, "MyObject", array.Array.Type.Array.Type.Defined)
	assert.Equal(t, "[[MyObject;3];10]", array.ResolveRustType())
	assert.Equal(t, 3, array.Array.Type.Array.Length)
}
