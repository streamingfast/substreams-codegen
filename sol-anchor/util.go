package solanchor

import (
	"fmt"
	"strings"
	"unicode"
)

func IDLTypeToRustType(idlType string) string {
	if IsPublicKey(idlType) {
		return "[u8;32]"
	}

	switch idlType {
	case "u8", "i8", "u16", "i16", "u64", "i64", "u32", "i32":
		return "u64"
	case "string":
		return "String"
	case "bytes":
		return "Vec<u8>"
	default:
		return idlType
	}
}

func IDLTypeToProtobufType(idlType string) string {
	if IsPublicKey(idlType) {
		return "PubKey"
	}

	switch idlType {
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "u8", "i8", "u16", "i16", "u64", "i64", "u32", "i32":
		return "uint64"
	case "u128", "i128":
		return "string" // Protobuf doesn't support 128-bit ints directly
	case "f64":
		return "double"
	case "bytes":
		return "bytes"
	default:
		// Assume it's a user-defined message
		return idlType
	}
}

func IsPublicKey(typeName string) bool {
	return typeName == "publicKey" || typeName == "pubKey" || typeName == "Pubkey" || typeName == "PubKey" || typeName == "pubkey"
}

func GetTypeByName(types []Type, name string) *Type {
	for _, t := range types {
		if t.Name == name {
			return &t
		}
	}

	return nil
}

func ComposeIDLRustNamespaceType(typeName string) string {
	return fmt.Sprintf("%s%s", IDLRustNamespace, typeName)
}

func ComposeProgramRustNamespaceType(typeName string) string {
	return fmt.Sprintf("%s%s", ProgramRustNamespace, typeName)
}

func PrintDefinedTree(typeName string, fieldName string, types []Type) string {
	t := GetTypeByName(types, typeName)
	if t == nil {
		return ""
	}

	if t.Type.IsStruct() {
		var fieldsInString strings.Builder

		for _, structField := range t.Type.Struct.Fields {
			resolvedFieldType, err := structField.Type.GetResolvedFieldType()
			if err != nil {
				continue
			}

			variableNameInner := "idlType"
			if _, ok := resolvedFieldType.(*Defined); ok {
				variableNameInner = fmt.Sprintf("%s", variableNameInner)
			}

			rustMappings := resolvedFieldType.PrintRustMappings(structField.SnakeCaseName(), variableNameInner, types)
			fieldsInString.WriteString(fmt.Sprintf("\t\t\t\t\t%s\n", rustMappings))
		}

		return fmt.Sprintf(`
			fn map_defined_%s(idlType: %s%s) -> %s%s {
				return %s%s {
%s
				}
			}
		`, toSnakeCase(typeName, false), IDLRustNamespace, typeName, ProgramRustNamespace, ToRustPascalCase(typeName), ProgramRustNamespace, ToRustPascalCase(typeName), fieldsInString.String())
	}

	if t.Type.IsEnum() {
		var matchArms strings.Builder

		for _, variant := range t.Type.Enum.Variants {
			protobufVariantName := fmt.Sprintf("%s%sVariant", typeName, variant.Name)
			oneofWrapper := fmt.Sprintf("pb::substreams::v1::program::%s::Kind::%s", toSnakeCase(typeName, false), variant.Name)

			if len(variant.Fields) == 0 {
				// Unit variant
				matchArms.WriteString(fmt.Sprintf(`
			idl::idl::%s::%s {} => pb::substreams::v1::program::%s {
				kind: Some(%s(%s {}))
			},`,
					typeName, variant.Name,
					ToRustPascalCase(typeName),
					oneofWrapper,
					ComposeProgramRustNamespaceType(protobufVariantName),
				))
			} else {
				var bindingFields []string
				var mappedFields []string

				for _, field := range variant.Fields {
					bindingName := fmt.Sprintf("%s", field.SnakeCaseName())
					bindingFields = append(bindingFields, fmt.Sprintf("%s", bindingName))

					resolvedFieldType, err := field.Type.GetResolvedFieldType()
					if err != nil {
						continue
					}
					mapping := resolvedFieldType.PrintRustMappings(field.SnakeCaseName(), "", types)
					mappedFields = append(mappedFields, fmt.Sprintf("%s", mapping))
				}

				matchArms.WriteString(fmt.Sprintf(`
			idl::idl::%s::%s { %s } => pb::substreams::v1::program::%s {
				kind: Some(%s(%s {
%s
				}))
			},`,
					typeName,
					variant.Name,
					strings.Join(bindingFields, ", "),
					ToRustPascalCase(typeName),
					oneofWrapper,
					ComposeProgramRustNamespaceType(protobufVariantName),
					indentLines(mappedFields, 4),
				))
			}
		}

		return fmt.Sprintf(`
			fn map_defined_%s(idlType: idl::idl::%s) -> pb::substreams::v1::program::%s {
				match idlType {
%s
				}
			}
		`, toSnakeCase(typeName, false), typeName, ToRustPascalCase(typeName), matchArms.String())
	}

	return ""
}

func indentLines(lines []string, indentLevel int) string {
	indent := strings.Repeat("\t", indentLevel)
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(fmt.Sprintf("%s%s\n", indent, line))
	}
	return b.String()
}

func ProtobufEnumVariantToRust(name string) string {
	var result strings.Builder
	runes := []rune(name)

	for i := 0; i < len(runes); i++ {
		// If we find a sequence of 2+ uppercase letters (acronym), normalize it
		if unicode.IsUpper(runes[i]) {
			start := i
			for i+1 < len(runes) && unicode.IsUpper(runes[i+1]) {
				i++
			}

			// If it was a single uppercase letter, just append as-is
			if i == start {
				result.WriteRune(runes[start])
			} else {
				// Normalize acronym (e.g. FX → Fx)
				result.WriteRune(runes[start])
				for j := start + 1; j <= i; j++ {
					result.WriteRune(unicode.ToLower(runes[j]))
				}
			}
		} else {
			result.WriteRune(runes[i])
		}
	}

	return result.String()
}

func IsPrimitiveTypeToStringNeeded(typeName string) bool {
	if typeName == "u128" || typeName == "f128" {
		return true
	}

	return false
}

func PrintMapPrimitiveToString(primitiveType string, fieldName string, variableName string) string {
	return fmt.Sprintf("%s: map_primitive_to_string(%s),", fieldName, variableName)
}

func ToRustPascalCase(t string) string {
	return t
}

func toLowerCaseCapitalized(input string) string {
	if input == "" {
		return ""
	}
	lower := strings.ToLower(input)
	runes := []rune(lower)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)

}

func ToRustFriendlyPascalCase(input string) string {
	var result strings.Builder
	runes := []rune(input)
	capNext := true
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			capNext = true
			continue
		}
		if capNext {
			result.WriteRune(unicode.ToUpper(ch))
			capNext = false
		} else {
			result.WriteRune(unicode.ToLower(ch))
		}
	}
	return result.String()
}
