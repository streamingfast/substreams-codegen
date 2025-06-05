package solanchor

import (
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
		}
	}

	return fieldList
}

func (i *IDL) PrintProtobufNestedTypes() string {
	// Collect all the complex types
	allFields := i.GetFieldsFromInstructionsEventsTypesAndAccounts()

	// Get all types that we should generate
	rustTypes := make([]string, 0)
	for _, f := range allFields {
		rustTypes = append(rustTypes, f.Type.ResolveRustType())
	}
	// remove duplicates
	rustTypes = uniqueStrings(rustTypes)

	output := "";

	return output
}

func (i *IDL) ProgramID() string {
	if i.Metadata.Address != "" {
		return i.Metadata.Address
	}
	return i.Address
}

func (i *IDL) IsTypeUsed(typeName string) bool {
	for _, instruction := range i.Instructions {
		for _, arg := range instruction.Args {
			if arg.Type.IsTypeUsed(typeName) {
				return true
			}
		}
	}

	for _, tp := range i.Types {
		if tp.Type.IsStruct() {
			for _, arg := range tp.Type.Struct.Fields {
				if arg.Type.IsTypeUsed(typeName) && i.IsTypeUsed(tp.Name) {
					return true
				}
			}
		} else if tp.Type.IsEnum() {
			for _, arg := range tp.Type.Enum.Variants {
				if arg.Name == typeName && i.IsTypeUsed(tp.Name) {
					return true
				}
			}
		}
	}

	for _, tp := range i.Accounts {
		if tp.Type.IsStruct() {
			for _, arg := range tp.Type.Struct.Fields {
				if arg.Type.IsTypeUsed(typeName) && i.IsTypeUsed(tp.Name) {
					return true
				}
			}
		} else if tp.Type.IsEnum() {
			for _, arg := range tp.Type.Enum.Variants {
				if arg.Name == typeName && i.IsTypeUsed(tp.Name) {
					return true
				}
			}
		}
	}

	for _, event := range i.Events {
		for _, field := range event.Fields {
			if field.Type.IsTypeUsed(typeName) {
				return true
			}
		}
	}

	return false
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

func IsPublicKey(idlType string) bool {
	return idlType == "publicKey" || idlType == "pubkey"
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
