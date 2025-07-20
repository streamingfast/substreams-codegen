package solanchor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-cz/textcase"
)

// --- TYPES

type Type struct {
	Name string      `json:"name"`
	Type TypeDetails `json:"type"`
}

func (f *Type) PrintRustStruct(allTypes []Type) string {
	return PrintDefinedTree(f.Name, "idlType", allTypes)
}

func (t *Type) SnakeCaseName() string {
	return toSnakeCase(t.Name, true)
}

func (t *Type) SnakeCaseNameUpperCase() string {
	return strings.ToUpper(t.SnakeCaseName())
}

func (t *Type) PascalCaseName() string {
	return ToRustPascalCase(textcase.PascalCase(t.Name))
}

func (t *Type) LowerCaseCapitalizedName() string {
	return toLowerCaseCapitalized(t.Name)
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

	return fmt.Errorf("failed to unmarshal TypeDetails: %s", string(data))
}

type TypeStruct struct {
	Kind   string  `json:"kind"`
	Fields []Field `json:"fields"`
}

/* ENUM */
type TypeEnum struct {
	Kind     string            `json:"kind"`
	Variants []TypeEnumVariant `json:"variants"`
}

/*
func (f *TypeEnum) PrintNecessaryProtobufMessages() string {
	return strings.ToUpper(toSnakeCase(f.Name, true))
}

func (f *TypeEnum) PrintNecessaryRustStructs() string {
	return strings.ToUpper(toSnakeCase(f.Name, true))
}

func (f *TypeEnum) ResolveRustType() string {
	return f.T
}*/

type TypeEnumVariant struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

func (f *TypeEnumVariant) SnakeCaseName() string {
	return strings.ToUpper(toSnakeCase(f.Name, true))
}
