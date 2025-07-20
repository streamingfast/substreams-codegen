package solanchor

import (
	"strings"
	"unicode"
)

type IDL struct {
	Address      string        `json:"address"` // seen in the 'secret' program
	Events       []Event       `json:"events"`
	Instructions []Instruction `json:"instructions"`
	Accounts     []Type        `json:"accounts"`
	Metadata     Metadata      `json:"metadata"`
	Types        []Type        `json:"types"`
}

/*
	Old IDLs do not contain the "discriminator" in the JSON, so the only way is to generate them through the Anchor library.
	We must support both.
*/
func (i *IDL) IsOldIDLFormat() bool {
	if len(i.Instructions) > 0 {
		return len(i.Instructions[0].Discriminator) == 0
	}

	if len(i.Events) > 0 {
		return len(i.Events[0].Discriminator) == 0
	}

	return false
}

func (i *IDL) AccountsAndTypes() []Type {
	return append(i.Types, i.Accounts...)
}

func (i *IDL) GetFieldsFromInstructionsEventsTypesAndAccounts() []Field {
	fieldList := make([]Field, 0)
	for _, inst := range i.Instructions {
		fieldList = append(fieldList, inst.Args...)
	}

	for _, evt := range i.Events {
		fieldList = append(fieldList, evt.Fields...)
	}

	for _, t := range i.Types {
		if t.Type.IsStruct() {
			fieldList = append(fieldList, t.Type.Struct.Fields...)
		} else {
			for _, variant := range t.Type.Enum.Variants {
				fieldList = append(fieldList, variant.Fields...)
			}
		}
	}

	return fieldList
}

/*
	Generate all the necessary Protobuf message (from instructions, events, types and accounts)
*/
func (i *IDL) PrintNecessaryProtobufMessages() string {
	allFields := i.GetFieldsFromInstructionsEventsTypesAndAccounts()
	seen := make(map[string]struct{})
	outputs := []string{}

	for _, f := range allFields {
		resolvedType, err := f.Type.GetResolvedFieldType()
		if err != nil {
			continue
		}

		code := strings.TrimSpace(resolvedType.PrintNecessaryProtobufMessages())
		if _, exists := seen[code]; !exists && code != "" {
			seen[code] = struct{}{}
			outputs = append(outputs, code)
		}
	}

	return strings.Join(outputs, "\n\n")
}

/*
	Generate all the necessary Rust structs (from instructions, events, types and accounts)
*/
func (i *IDL) PrintNecessaryRustStructs() string {
	allFields := i.GetFieldsFromInstructionsEventsTypesAndAccounts()
	seen := make(map[string]struct{})
	outputs := []string{}

	for _, f := range allFields {
		resolvedType, err := f.Type.GetResolvedFieldType()
		if err != nil || f.Type.IsDefined() {
			continue
		}

		code := strings.TrimSpace(resolvedType.PrintNecessaryRustStructs(f.SnakeCaseName(), i.Types))
		if _, exists := seen[code]; !exists && code != "" {
			seen[code] = struct{}{}
			outputs = append(outputs, code)
		}
	}

	for _, t := range i.Types {
		code := strings.TrimSpace(t.PrintRustStruct(i.Types))
		if _, exists := seen[code]; !exists && code != "" {
			seen[code] = struct{}{}
			outputs = append(outputs, code)
		}
	}

	return strings.Join(outputs, "\n\n")
}

func (i *IDL) ProgramID() string {
	if i.Metadata.Address != "" {
		return i.Metadata.Address
	}
	return i.Address
}

/*
	Some IDLs have the event fields defined in the `types` section of the JSON.
	In this function we check wether we should looks in "events" or "types" to find out the events.
*/
func (i *IDL) MoveEventsIfNecessary() {
	for idx := range i.Events {
		event := &i.Events[idx]

		if len(event.Fields) > 0 {
			continue
		}

		for _, t := range i.Types {
			if event.Name == t.Name {
				MoveTypeToEvent(event, t)
			}
		}
	}
}

func MoveTypeToEvent(event *Event, t Type) {
	for _, f := range t.Type.Struct.Fields {
		event.Fields = append(event.Fields, Field{
			Name: f.Name,
			Type: f.Type,
		})
	}
}

func MoveTypeToAccount(event *Event, t Type) {
	for _, f := range t.Type.Struct.Fields {
		event.Fields = append(event.Fields, Field{
			Name: f.Name,
			Type: f.Type,
		})
	}
}

func (i *IDL) IsTypeEnum(typeName string) bool {
	for _, typeObj := range i.Types {
		if typeObj.Name == typeName {
			return typeObj.Type.IsEnum()
		}
	}

	return false
}

type Metadata struct {
	Address string `json:"address"`
	Name    string `json:"name"` // seen in the 'secret' program
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
	case "PubKey", "pubkey", "publicKey":
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

	return rustType
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
