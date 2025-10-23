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

// Test unmarshalVec
func TestUnmarshalVecSimple(t *testing.T) {
	json := []byte(`{"vec": "string"}`)

	kind, kindType := unmarshalVec(json)

	assert.Equal(t, "simple", kind)
	assert.Equal(t, "string", kindType)
}

func TestUnmarshalVecDefined(t *testing.T) {
	json := []byte(`{"vec": {"defined": "string"}}`)

	kind, kindType := unmarshalVec(json)

	assert.Equal(t, "defined", kind)
	assert.Equal(t, "string", kindType)
}

// Test unmarshalOption
func TestUnmarshalOptionSimple(t *testing.T) {
	json := []byte(`{"option": "string"}`)

	optionType, result := unmarshalOption(json)

	assert.Equal(t, "simple", optionType)
	assert.Equal(t, "string", result)
}

func TestUnmarshalOptionDefined(t *testing.T) {
	json := []byte(`{"option": {"defined": "string"}}`)

	optionType, result := unmarshalOption(json)

	assert.Equal(t, "defined", optionType)
	assert.Equal(t, "string", result)
}

func TestUnmarshalOptionDefinedWithName(t *testing.T) {
	json := []byte(`{"option": {"defined": {"name": "string"}}}`)

	optionType, result := unmarshalOption(json)

	assert.Equal(t, "defined", optionType)
	assert.Equal(t, "string", result)
}

func TestUnmarshalOptionVecSimple(t *testing.T) {
	json := []byte(`{"option": {"vec": "string"}}`)

	optionType, result := unmarshalOption(json)

	assert.Equal(t, "vecSimple", optionType)
	assert.Equal(t, "string", result)
}

func TestUnmarshalOptionVecDefined(t *testing.T) {
	json := []byte(`{"option": {"vec": {"defined": "string"}}}`)

	optionType, result := unmarshalOption(json)

	assert.Equal(t, "vecDefined", optionType)
	assert.Equal(t, "string", result)
}
