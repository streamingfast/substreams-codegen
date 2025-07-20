package solanchor

import (
	"fmt"
	//"regexp"
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
			idl::idl::program::types::%s::%s {} => pb::substreams::v1::program::%s {
				kind: Some(%s(%s {}))
			},`,
					typeName, variant.Name,
					ToRustPascalCase(typeName),
					oneofWrapper,
					ComposeProgramRustNamespaceType(ToRustPascalCase(protobufVariantName)),
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
			idl::idl::program::types::%s::%s { %s } => pb::substreams::v1::program::%s {
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
			fn map_defined_%s(idlType: idl::idl::program::types::%s) -> pb::substreams::v1::program::%s {
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

func toLowerCaseCapitalized(input string) string {
	if input == "" {
		return ""
	}
	lower := strings.ToLower(input)
	runes := []rune(lower)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)

}

var knownTrailingAcronyms = map[string]struct{}{
	"FX":   {},
	"V2":   {},
	"FXV2": {},
	"XYZ":  {}, // if you want to preserve XYZ too
}

// NormalizePascalCase converts Protobuf-style PascalCase to Rust-style PascalCase,
// preserving trailing acronyms like FXV2.
func ToRustPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Step 1: Split into segments
	segments := splitPascal(s)

	// Step 2: Check trailing suffix against known acronyms
	for i := range segments {
		tail := strings.Join(segments[i:], "")
		if _, ok := knownTrailingAcronyms[tail]; ok {
			// Preserve tail as-is, normalize head
			head := segments[:i]
			for j, seg := range head {
				head[j] = normalizeSegment(seg)
			}
			return strings.Join(append(head, tail), "")
		}
	}

	// No match — normalize all parts
	for i, seg := range segments {
		segments[i] = normalizeSegment(seg)
	}
	return strings.Join(segments, "")
}

// splitPascal splits PascalCase string into segments using character rules
func splitPascal(s string) []string {
	var segments []string
	var current strings.Builder
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])
		if i+1 < len(runes) && isBoundary(runes, i) {
			segments = append(segments, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}
	return segments
}

// normalizeSegment capitalizes first letter, lowercases others (for letters)
func normalizeSegment(part string) string {
	if len(part) == 0 {
		return part
	}

	var sb strings.Builder
	runes := []rune(part)
	sb.WriteRune(unicode.ToUpper(runes[0]))
	for _, r := range runes[1:] {
		if unicode.IsLetter(r) {
			sb.WriteRune(unicode.ToLower(r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// isBoundary determines if there's a word boundary between runes[i] and runes[i+1]
func isBoundary(runes []rune, i int) bool {
	curr := runes[i]
	next := runes[i+1]

	if unicode.IsLower(curr) && unicode.IsUpper(next) {
		return true
	}
	if (unicode.IsLetter(curr) && unicode.IsDigit(next)) || (unicode.IsDigit(curr) && unicode.IsLetter(next)) {
		return true
	}
	if unicode.IsUpper(curr) && unicode.IsUpper(next) {
		if i+2 < len(runes) && unicode.IsLower(runes[i+2]) {
			return true
		}
	}
	return false
}
