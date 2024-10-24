package solanchor

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/golang-cz/textcase"
)

type IDL struct {
	Events       []Event       `json:"events"`
	Instructions []Instruction `json:"instructions"`
	Metadata     Metadata      `json:"metadata"`
	Types        []Type        `json:"types"`
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
	return toSnakeCase(e.Name, true)
}

// --- INSTRUCTIONS
type Instruction struct {
	Name string  `json:"name"`
	Args []Field `json:"args"`
}

func (i *Instruction) PascalCaseName() string {
	return textcase.PascalCase(i.Name)
}

func (i *Instruction) SnakeCaseName() string {
	return toSnakeCase(i.Name, true)
}

// --- FIELDS

type Field struct {
	Name string    `json:"name"`
	Type FieldType `json:"type"`
	//Index bool      `json:"index"`
}

func (f *Field) SnakeCaseName() string {
	return toSnakeCase(f.Name, true)
}

func (f *Field) SnakeCaseNameWithoutInitialUnderscore() string {
	return toSnakeCase(f.Name, false)
}

type FieldType struct {
	Simple     string
	Defined    string
	Array      string
	VecSimple  string
	VecDefined string
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

func (t *FieldType) IsVecDefined() bool {
	return t.VecDefined != ""
}

func (t *FieldType) IsVecSimple() bool {
	return t.VecSimple != ""
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

func (f *FieldType) Print(fieldName string, variableName string, types []Type) string {
	fieldNameSnakeCase := toSnakeCase(fieldName, true)
	fieldNameSnakeCaseWithoutInitialUnderscore := toSnakeCase(fieldName, false)

	if f.IsSimplePubKey() {
		return fmt.Sprintf("%s: %s.%s.toString(),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
	}

	if f.IsSimple() {
		cast := CastInRustIfNeeded(f.Simple)
		if cast != "" {
			return fmt.Sprintf("%s: %s.%s as %s,", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldName, cast)
		}

		return fmt.Sprintf("%s: %s.%s,", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
	}

	if f.IsDefined() {
		for _, t := range types {
			if t.Name == f.Defined {
				if t.Type.IsEnum() {
					return fmt.Sprintf("%s: map_enum_%s(%s.%s),", toSnakeCase(fieldName, false), t.SnakeCaseName(), variableName, toSnakeCase(fieldName, true))
				} else {
					var fieldsInString strings.Builder
					for _, structField := range t.Type.Struct.Fields {
						fieldsInString.WriteString(structField.Type.Print(structField.Name, variableName, types))
					}
					return fmt.Sprintf(`%s: %s {
							%s
						},`, toSnakeCase(fieldName, false), t.Name, fieldsInString.String())
				}
			}
		}
	}

	return ""
}

/*
A field could be of type:
    - simple (e.g. "string")
    - defined
    - array
    - vec
*/

func unmarshalDefined(data []byte) string {
	var definedType struct {
		Defined string `json:"defined"`
	}
	if err := json.Unmarshal(data, &definedType); err == nil && definedType.Defined != "" {
		return definedType.Defined
	}

	return ""
}

func unmarshalSimple(data []byte) string {
	var simpleType string
	if err := json.Unmarshal(data, &simpleType); err == nil && simpleType != "" {
		return simpleType
	}

	return ""
}

/*
Return types:

	bool: isDefined (true/false)
	string: type (simple/defined)
*/
func unmarshalVec(data []byte) (bool, string) {
	// Try simple
	var vecSimpleType struct {
		Vec string `json:"vec"`
	}
	err := json.Unmarshal(data, &vecSimpleType)
	if err == nil {
		return false, vecSimpleType.Vec
	}

	// Try defined
	type DefinedType struct {
		Defined string `json:"defined"`
	}
	var vecDefinedType struct {
		Vec DefinedType `json:"vec"`
	}
	err = json.Unmarshal(data, &vecDefinedType)
	//fmt.Printf("%s", string(vecDefinedType.Vec.Defined))
	if err == nil {
		return true, vecDefinedType.Vec.Defined
	}

	return false, ""
}

func (t *FieldType) UnmarshalJSON(data []byte) error {
	if result := unmarshalSimple(data); result != "" {
		t.Simple = result
		return nil
	}

	if result := unmarshalDefined(data); result != "" {
		t.Defined = result
		return nil
	}

	var arrayType struct {
		Array []interface{} `json:"array"`
	}
	if err := json.Unmarshal(data, &arrayType); err == nil && len(arrayType.Array) > 0 {
		stringType, ok1 := arrayType.Array[0].(string)
		if ok1 {
			t.Array = stringType
		}
		return nil
	}

	isDefined, result := unmarshalVec(data)
	if result != "" {
		if isDefined {
			t.VecDefined = result
		} else {
			t.VecSimple = result
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
	return toSnakeCase(t.Name, true)
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

func (t *TypeStructField) SnakeCaseName() string {
	return toSnakeCase(t.Name, true)
}

func (t *TypeStructField) SnakeCaseNameWithoutInitialUnderscore() string {
	return toSnakeCase(t.Name, false)
}

type TypeEnum struct {
	Kind     string            `json:"kind"`
	Variants []TypeEnumVariant `json:"variants"`
}

type TypeEnumVariant struct {
	Name string `json:"name"`
}

func (f *TypeEnumVariant) SnakeCaseName() string {
	return strings.ToUpper(toSnakeCase(f.Name, true))
}

// --- UTILS

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

func CastInRustIfNeeded(rustType string) string {
	switch rustType {
	case "u8":
		return "u64"
	}

	return ""
}

func PrintFieldTypeRecursively(fieldName string, fieldType FieldType, types []Type) string {
	if fieldType.IsSimple() {
		return fmt.Sprintf("%s: %s", toSnakeCase(fieldName, false), fieldName)
	}

	if fieldType.IsDefined() {
		for _, t := range types {
			if t.Type.IsEnum() {

			} else {
				var fieldsInString strings.Builder
				for _, f := range t.Type.Struct.Fields {
					fieldsInString.WriteString(PrintFieldTypeRecursively(f.SnakeCaseName(), f.Type, types))
				}
				return fmt.Sprintf(`%s {
					%s
				}`, t.Name, fieldsInString.String())
			}
		}
	}

	return ""
}

func toSnakeCase(str string, initialUnderscore bool) string {
	var result []rune

	for i, r := range str {
		if !initialUnderscore && r == '_' && i == 0 {
			continue
		}
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
