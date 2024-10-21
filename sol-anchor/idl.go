package solanchor

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type IDL struct {
	Events   []Event  `json:"events"`
	Metadata Metadata `json:"metadata"`
	Types    []Type   `json:"types"`
}

type Metadata struct {
	Address string `json:"address"`
}

// --- EVENTS

type Event struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

func (e *Event) SnakeCaseName() string {
	return toSnakeCase(e.Name)
}

// --- FIELDS

type Field struct {
	Name  string    `json:"name"`
	Type  FieldType `json:"type"`
	Index bool      `json:"index"`
}

func (f *Field) SnakeCaseName() string {
	return toSnakeCase(f.Name)
}

type FieldType struct {
	Simple  string
	Defined string
	Array   string
}

func (t *FieldType) IsSimple() bool {
	return t.Simple != ""
}

func (t *FieldType) IsSimplePubKey() bool {
	return t.Simple == "publicKey"
}

func (t *FieldType) IsDefined() bool {
	return t.Defined != ""
}

func (t *FieldType) IsArray() bool {
	return t.Array != ""
}

func (t *FieldType) Resolve() string {
	if t.IsSimplePubKey() {
		return "PubKey"
	}

	if t.IsSimple() {
		return t.Simple
	}

	if t.IsDefined() {
		return t.Defined
	}

	if t.IsArray() {
		return t.Array
	}

	return ""
}

func (t *FieldType) ResolveProtobufType() string {
	return ToProtobufType(t.Resolve())
}

func ToProtobufType(rustType string) string {
	switch rustType {
	case "u8":
		return "uint64"
	case "u64":
		return "uint64"
	case "i64":
		return "int64"
	case "f64":
		return "double"
	case "f32":
		return "float"
	case "i32":
		return "int32"
	case "u32":
		return "uint32"
	case "PubKey":
		return "string"
	}

	return rustType
}

func (t *FieldType) UnmarshalJSON(data []byte) error {
	var simpleType string
	if err := json.Unmarshal(data, &simpleType); err == nil {
		t.Simple = simpleType
		return nil
	}

	var definedType struct {
		Defined string `json:"defined"`
	}
	if err := json.Unmarshal(data, &definedType); err == nil && definedType.Defined != "" {
		t.Defined = definedType.Defined
		return nil
	}

	var arrayType struct {
		Array []interface{} `json:"array"`
	}
	if err := json.Unmarshal(data, &arrayType); err == nil && len(arrayType.Array) > 0 {
		stringType, ok1 := arrayType.Array[0].(string)
		//_, _ := arrayType.Array[1].(int)
		if ok1 {
			t.Array = stringType
		}
		return nil
	}

	return fmt.Errorf("failed to unmarshal Type: %s", string(data))
}

// --- TYPES

type Type struct {
	Name string      `json:"name"`
	Type TypeDetails `json:"type"`
}

func (t *Type) SnakeCaseName() string {
	return toSnakeCase(t.Name)
}

type TypeDetails struct {
	Kind   string
	Struct *TypeStruct
	Enum   *TypeEnum
}

func (t *TypeDetails) IsStruct() bool {
	return t.Kind == "struct"
}

func (t *TypeDetails) IsEnum() bool {
	return t.Kind == "enum"
}

func (t *TypeDetails) UnmarshalJSON(data []byte) error {
	var kindType struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &kindType); err == nil {
		switch kindType.Kind {
		case "enum":
			var typeEnum TypeEnum
			if err := json.Unmarshal(data, &typeEnum); err == nil {
				t.Kind = "enum"
				t.Enum = &typeEnum
				return nil
			}
		case "struct":
			var typeStruct TypeStruct
			if err := json.Unmarshal(data, &typeStruct); err == nil {
				t.Kind = "struct"
				t.Struct = &typeStruct
				return nil
			}
		}
		return nil
	}

	return fmt.Errorf("failed to unmarshal Type: %s", string(data))
}

type TypeStruct struct {
	Kind   string            `json:"kind"`
	Fields []TypeStructField `json:"fields"`
}

type TypeStructField struct {
	Name string    `json:"name"`
	Type FieldType `json:"type"`
}

func (f *TypeStructField) SnakeCaseName() string {
	return toSnakeCase(f.Name)
}

type TypeEnum struct {
	Kind     string            `json:"kind"`
	Variants []TypeEnumVariant `json:"variants"`
}

type TypeEnumVariant struct {
	Name string `json:"name"`
}

func (f *TypeEnumVariant) SnakeCaseName() string {
	return strings.ToUpper(toSnakeCase(f.Name))
}

// --- UTILS

func toSnakeCase(str string) string {
	var result []rune

	for i, r := range str {
		// Check if the character is uppercase
		if unicode.IsUpper(r) {
			// Add an underscore before the uppercase letter if it's not the first character
			if i > 0 {
				result = append(result, '_')
			}
			// Convert the uppercase letter to lowercase
			result = append(result, unicode.ToLower(r))
		} else {
			// Just add the character as is
			result = append(result, r)
		}
	}

	return string(result)
}
