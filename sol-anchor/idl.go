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
