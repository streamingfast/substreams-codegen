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

func (i *IDL) IsTypeUsed(typeName string) bool {
	for _, instruction := range i.Instructions {
		for _, arg := range instruction.Args {
			if arg.Type.IsTypeUsed(typeName) == true {
				return true
			}
		}
	}

	for _, event := range i.Events {
		for _, field := range event.Fields {
			if field.Type.IsTypeUsed(typeName) == true {
				return true
			}
		}
	}

	return false
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

// FIELD TYPE

type FieldTypeArray struct {
	Type   string
	Length string
}

type FieldTypeVec struct {
	Type      string
	IsDefined bool
}

type FieldTypeOption struct {
	Simple  *string
	Defined *string
	Vec     *FieldTypeVec
	Array   *FieldTypeArray
}

func (f *FieldTypeOption) IsSimple() bool {
	return f.Simple != nil
}

func (f *FieldTypeOption) IsDefined() bool {
	return f.Defined != nil
}

func (f *FieldTypeOption) IsVec() bool {
	return f.Vec != nil
}

func (f *FieldTypeOption) IsArray() bool {
	return f.Array != nil
}

type FieldType struct {
	Simple  string
	Defined string
	Array   *FieldTypeArray
	Vec     *FieldTypeVec
	Option  *FieldTypeOption
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
	return t.Array != nil
}

func (t *FieldType) IsVec() bool {
	return t.Vec != nil
}

func (t *FieldType) IsOption() bool {
	return t.Option != nil
}

func (t *FieldType) IsTypeUsed(typeName string) bool {
	if t.IsDefined() && t.Defined == typeName {
		return true
	}

	if t.IsVec() && t.Vec.Type == typeName {
		return true
	}

	if t.IsOption() && *t.Option.Defined == typeName {
		return true
	}

	return false
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
		return t.Array.Type
	}

	if t.IsVec() {
		return t.Vec.Type
	}

	if t.IsOption() {
		if t.Option.IsSimple() {
			return *t.Option.Simple
		}

		if t.Option.IsDefined() {
			return *t.Option.Defined
		}

		if t.Option.IsVec() {
			return t.Option.Vec.Type
		}

		if t.Option.IsArray() {
			return t.Array.Type
		}
	}

	return ""
}

func (t *FieldType) ResolveProtobufType() string {
	return ToProtobufType(t.Resolve())
}

func PrintSimple(simpleType string, fieldName string, variableName string, isOptional bool) string {
	fieldNameSnakeCase := toSnakeCase(fieldName, true)
	fieldNameSnakeCaseWithoutInitialUnderscore := toSnakeCase(fieldName, false)

	// If cast is needed
	cast := CastInRustIfNeeded(simpleType)
	if cast != "" {
		if isOptional {
			return fmt.Sprintf("%s: Some(%s.%s as %s),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase, cast)
		}
		return fmt.Sprintf("%s: %s.%s as %s,", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase, cast)
	}

	// If cast is NOT needed
	if isOptional {
		return fmt.Sprintf("%s: Some(%s.%s),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
	}
	return fmt.Sprintf("%s: %s.%s,", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
}

func PrintDefined(typeName string, fieldName string, variableName string, types []Type, isVec bool, isOption bool) string {
	for _, t := range types {
		if t.Name == typeName {
			if t.Type.IsEnum() {
				return fmt.Sprintf("%s: map_enum_%s(%s.%s),", toSnakeCase(fieldName, false), t.SnakeCaseName(), toSnakeCase(variableName, true), toSnakeCase(fieldName, true))
			} else {
				var fieldsInString strings.Builder
				for _, structField := range t.Type.Struct.Fields {
					fieldString := fmt.Sprintf("%s.%s", toSnakeCase(variableName, true), toSnakeCase(fieldName, true))
					if variableName == "" {
						fieldString = toSnakeCase(fieldName, true)
					}

					if isOption {
						fieldString = fmt.Sprintf("%s: map_option_%s(%s.%s),", toSnakeCase(fieldName, true), toSnakeCase(fieldName, true), toSnakeCase(variableName, true), toSnakeCase(fieldName, true))
						fieldsInString.WriteString(fieldString)
						continue
					}

					fieldsInString.WriteString(structField.Type.Print(structField.Name, fieldString, types))
				}

				if isVec {
					return fmt.Sprintf(`%s {
						%s
					}`, t.Name, fieldsInString.String())
				}

				if isOption {
					return fieldsInString.String()
				}

				return fmt.Sprintf(`%s: Some(%s {
						%s
					}),`, toSnakeCase(fieldName, false), t.Name, fieldsInString.String())
			}
		}
	}

	return ""
}

func (f *FieldType) Print(fieldName string, variableName string, types []Type) string {
	fieldNameSnakeCase := toSnakeCase(fieldName, true)
	fieldNameSnakeCaseWithoutInitialUnderscore := toSnakeCase(fieldName, false)

	if f.IsSimplePubKey() {
		return fmt.Sprintf("%s: %s.%s.to_string(),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
	}

	if f.IsSimple() {
		return PrintSimple(f.Simple, fieldName, variableName, false)
	}

	if f.IsDefined() {
		return PrintDefined(f.Defined, fieldName, variableName, types, false, false)
	}

	if f.IsArray() {
		cast := CastInRustIfNeeded(f.Array.Type)
		if cast == "" {
			return fmt.Sprintf("%s: %s.%s.to_vec(),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
		}

		return fmt.Sprintf("%s: %s.%s.into_iter().map(|f| f as %s).collect(),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase, cast)
	}

	if f.IsVec() && f.Vec.IsDefined {
		return fmt.Sprintf("%s: %s.%s.into_iter().map(|%s| %s).collect(),", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase, fieldNameSnakeCase, PrintDefined(f.Vec.Type, fieldNameSnakeCase, "", types, true, false))
	}

	if f.IsVec() && !f.Vec.IsDefined {
		return fmt.Sprintf("%s: %s.%s,", fieldNameSnakeCaseWithoutInitialUnderscore, variableName, fieldNameSnakeCase)
	}

	if f.IsOption() {
		if f.Option.IsSimple() {
			return PrintSimple(*f.Option.Simple, fieldName, variableName, true)
		}

		if f.Option.IsDefined() {
			return PrintDefined(*f.Option.Defined, fieldName, variableName, types, false, true)
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
  - optional (simple, defined or vec)
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

/*
Return types:
  - option type (simple, defined or vec)
  - type
*/
func unmarshalOption(data []byte) (string, string) {
	// Try simple
	var optionSimpleType struct {
		Option string `json:"option"`
	}
	if err := json.Unmarshal(data, &optionSimpleType); err == nil && optionSimpleType.Option != "" {
		return "simple", optionSimpleType.Option
	}

	// Try defined
	type DefinedType struct {
		Defined string `json:"defined"`
	}
	var optionDefinedType struct {
		Option DefinedType `json:"option"`
	}
	if err := json.Unmarshal(data, &optionDefinedType); err == nil && optionDefinedType.Option.Defined != "" {
		return "defined", optionDefinedType.Option.Defined
	}

	// Try vec simple
	type VecTypeSimple struct {
		Vec string `json:"vec"`
	}
	var optionVecSimpleType struct {
		Option VecTypeSimple `json:"option"`
	}
	if err := json.Unmarshal(data, &optionVecSimpleType); err == nil && optionVecSimpleType.Option.Vec != "" {
		return "vecSimple", optionVecSimpleType.Option.Vec
	}

	// Try vec defined
	type VecTypeDefined struct {
		Vec DefinedType `json:"vec"`
	}
	var optionVecDefinedType struct {
		Option VecTypeDefined `json:"option"`
	}
	if err := json.Unmarshal(data, &optionVecDefinedType); err == nil && optionVecDefinedType.Option.Vec.Defined != "" {
		return "vecDefined", optionVecDefinedType.Option.Vec.Defined
	}

	// TODO: Try array?

	return "", ""
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
		stringType, stringOk := arrayType.Array[0].(string)
		// TODO: The length should be in position [1] of the array, but it's not decoded correctly. Why?
		//intType, intOk := arrayType.Array[1].(int)

		if stringOk {
			t.Array = &FieldTypeArray{
				Type:   stringType,
				Length: string(0),
			}
		}
		return nil
	}

	isDefined, result := unmarshalVec(data)
	if result != "" {
		t.Vec = &FieldTypeVec{
			Type:      result,
			IsDefined: isDefined,
		}

		return nil
	}

	fieldType, result := unmarshalOption(data)
	if fieldType != "" && result != "" {
		var fieldTypeOption FieldTypeOption

		switch fieldType {
		case "simple":
			fieldTypeOption = FieldTypeOption{Simple: &result}
		case "defined":
			fieldTypeOption = FieldTypeOption{Defined: &result}
		case "vecSimple":
			fieldTypeOption = FieldTypeOption{Vec: &FieldTypeVec{Type: result, IsDefined: false}}
		case "vecDefined":
			fieldTypeOption = FieldTypeOption{Vec: &FieldTypeVec{Type: result, IsDefined: true}}
		}

		t.Option = &fieldTypeOption

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

func (t *Type) PascalCaseName() string {
	return textcase.PascalCase(t.Name)
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
	case "u16":
		return "uint64"
	case "u128":
		return "uint64"
	case "i128":
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
	case "u16":
		return "u64"
	case "u128":
		return "u64"
	case "i128":
		return "u64"
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
