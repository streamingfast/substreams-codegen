package solanchor

import (
	"fmt"
	"strings"
)

var (
	IDLRustNamespace     = "idl::idl::"
	ProgramRustNamespace = "pb::substreams::v1::program::"
)

// Parent struct
type ResolvedFieldTypeCommon struct {
	Type string
}
type ResolvedFieldType interface {
	ResolveRustType() string
	ResolveProtobufType() string
	PrintNecessaryProtobufMessages() string
	PrintNecessaryRustStructs(fieldName string, types []Type) string
	PrintRustMappings(fieldName string, variableName string, types []Type) string
}

// simple and defiend
type Simple struct {
	ResolvedFieldTypeCommon
}

func (f *Simple) ResolveRustType() string {
	return IDLTypeToRustType(f.Type)
}
func (f *Simple) ResolveProtobufType() string {
	return IDLTypeToProtobufType(f.Type)
}
func (f *Simple) PrintNecessaryProtobufMessages() string {
	return ""
}
func (f *Simple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *Simple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	if variableName != "" {
		if IsPrimitiveTypeToStringNeeded(f.Type) {
			return PrintMapPrimitiveToString(f.Type, fieldName, fmt.Sprintf("%s.%s", variableName, fieldName))
		}

		return fmt.Sprintf("%s: %s.%s as %s,", fieldName, variableName, fieldName, IDLTypeToRustType(f.Type))
	}

	if IsPrimitiveTypeToStringNeeded(f.Type) {
		return PrintMapPrimitiveToString(f.Type, fieldName, variableName)
	}

	return fmt.Sprintf("%s: %s as %s,", fieldName, fieldName, IDLTypeToRustType(f.Type))
}

type Defined struct {
	ResolvedFieldTypeCommon
}

func (f *Defined) ResolveRustType() string {
	return f.Type
}
func (f *Defined) ResolveProtobufType() string {
	return IDLTypeToProtobufType(f.Type)
}
func (f *Defined) PrintNecessaryProtobufMessages() string {
	return ""
}
func (f *Defined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return PrintDefinedTree(f.Type, fieldName, types)
}
func (f *Defined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	for _, t := range types {
		if t.Name == f.Type || IsPublicKey(f.Type) {
			// TODO: enums
			//if t.Type.IsStruct() {
			/*var fieldsInString strings.Builder

			for _, structField := range t.Type.Struct.Fields {
				resolvedFieldType, err := structField.Type.GetResolvedFieldType()
				if err != nil {
					continue
				}

				variableNameInner := variableName
				if _, ok := resolvedFieldType.(*Defined); ok {
					variableNameInner = fmt.Sprintf("%s.%s", variableNameInner, structField.SnakeCaseName())
				}

				rustMappings := resolvedFieldType.PrintRustMappings(structField.SnakeCaseName(), variableNameInner, types)
				fieldsInString.WriteString(fmt.Sprintf("            %s,\n", rustMappings))
			}*/

			if variableName != "" {
				return fmt.Sprintf(`
					%s: Some(map_defined_%s(%s.%s)),
				`, fieldName, toSnakeCase(f.Type, false), variableName, fieldName)
			}

			return fmt.Sprintf(`
				%s: Some(map_defined_%s(%s)),
			`, fieldName, toSnakeCase(f.Type, false), fieldName)

			//}
		}
	}
	return ""
}

// vec
type VecSimple struct {
	ResolvedFieldTypeCommon
}

func (f *VecSimple) ResolveRustType() string {
	return fmt.Sprintf("Vec<%s>", IDLTypeToRustType(f.Type))
}
func (f *VecSimple) ResolveProtobufType() string {
	return fmt.Sprintf("repeated %s", IDLTypeToProtobufType(f.Type))
}
func (f *VecSimple) PrintNecessaryProtobufMessages() string {
	return ""
}
func (f *VecSimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *VecSimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: %s.%s.into_iter().map(|f| f as %s).collect(),", fieldName, variableName, fieldName, IDLTypeToProtobufType(f.Type))
}

type VecDefined struct {
	ResolvedFieldTypeCommon
}

func (f *VecDefined) ResolveRustType() string {
	return fmt.Sprintf("Vec<%s>", f.Type)
}
func (f *VecDefined) ResolveProtobufType() string {
	return fmt.Sprintf("repeated %s", f.Type)
}
func (f *VecDefined) PrintNecessaryProtobufMessages() string {
	return ""
}
func (f *VecDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *VecDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	for _, t := range types {
		if t.Name == f.Type {

			if variableName != "" {
				return fmt.Sprintf("%s: %s.%s.into_iter().map(|f| \t map_defined_%s(f)).collect(),",
					fieldName,
					variableName,
					fieldName,
					toSnakeCase(f.Type, false),
				)
			}

			return fmt.Sprintf("%s: %s.into_iter().map(|f| \t map_defined_%s(f)).collect(),",
				fieldName,
				fieldName,
				toSnakeCase(f.Type, false),
			)

		}
	}
	return ""
}

type VecOptionSimple struct {
	ResolvedFieldTypeCommon
}

func (f *VecOptionSimple) ResolveRustType() string {
	return fmt.Sprintf("Vec<Option<%s>>", IDLTypeToRustType(f.Type))
}
func (f *VecOptionSimple) ResolveProtobufType() string {
	return fmt.Sprintf("repeated %s", IDLTypeToProtobufType(f.Type))
}
func (f *VecOptionSimple) PrintNecessaryProtobufMessages() string {
	return ""
}
func (f *VecOptionSimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *VecOptionSimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: %s.%s.into_iter().flatten().map(|f| f as %s).collect(),", fieldName, variableName, fieldName, IDLTypeToProtobufType(f.Type))
}

type VecOptionDefined struct {
	ResolvedFieldTypeCommon
}

func (f *VecOptionDefined) ResolveRustType() string {
	return fmt.Sprintf("Vec<Option<%s>>", f.Type)
}
func (f *VecOptionDefined) ResolveProtobufType() string {
	return fmt.Sprintf("repeated %s", f.Type)
}
func (f *VecOptionDefined) PrintNecessaryProtobufMessages() string {
	return ""
}
func (f *VecOptionDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *VecOptionDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	for _, t := range types {
		if t.Name == f.Type {
			variableName = "f"
			// TODO: enums
			if t.Type.IsStruct() {
				var fieldsInString strings.Builder

				for _, structField := range t.Type.Struct.Fields {
					resolvedFieldType, err := structField.Type.GetResolvedFieldType()
					if err != nil {
						continue
					}

					variableNameInner := variableName
					if _, ok := resolvedFieldType.(*Defined); ok {
						variableNameInner = fmt.Sprintf("%s.%s", variableNameInner, structField.SnakeCaseName())
					}

					rustMappings := resolvedFieldType.PrintRustMappings(structField.SnakeCaseName(), variableNameInner, types)
					fieldsInString.WriteString(fmt.Sprintf("\t\t\t%s,\n", rustMappings))
				}

				return fmt.Sprintf("%s: %s.%s.into_iter().map(|f| {\n \t return %s {\n%s\n\t}\n}).collect(),",
					fieldName,
					variableName,
					fieldName,
					f.Type,
					fieldsInString.String(),
				)
			}
		}
	}
	return ""
}

// option
type OptionSimple struct {
	ResolvedFieldTypeCommon
}

func (f *OptionSimple) ResolveRustType() string {
	return fmt.Sprintf("Option<%s>", IDLTypeToRustType(f.Type))
}
func (f *OptionSimple) ResolveProtobufType() string {
	return fmt.Sprintf("optional %s", IDLTypeToProtobufType(f.Type))
}
func (f *OptionSimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return fmt.Sprintf(`
			fn map_option_%s(idlType: Option<%s>) -> Option<%s> {
				if (idlType.is_none()) {
					return None
				}
				return Some(idlType.unwrap() as %s)
			}
		`, toSnakeCase(f.Type, false), f.Type, f.Type, IDLTypeToRustType(f.Type))
}
func (f *OptionSimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: map_option_%s(%s.%s),", fieldName, toSnakeCase(f.Type, false), variableName, fieldName)
}
func (f *OptionSimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type OptionDefined struct {
	ResolvedFieldTypeCommon
}

func (f *OptionDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<%s>", f.Type)
}
func (f *OptionDefined) ResolveProtobufType() string {
	return fmt.Sprintf("optional %s", f.Type)
}
func (f *OptionDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	if variableName == "" {
		return fmt.Sprintf("%s: map_option_%s(%s),", fieldName, toSnakeCase(f.Type, false), fieldName)
	}
	return fmt.Sprintf("%s: map_option_%s(%s.%s),", fieldName, toSnakeCase(f.Type, false), variableName, fieldName)
}
func (f *OptionDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return fmt.Sprintf(`
			fn map_option_%s(idlType: Option<%s%s>) -> Option<%s%s> {
				if (idlType.is_none()) {
					return None
				}
				return Some(map_defined_%s(idlType.unwrap()))
			}
		`, toSnakeCase(f.Type, false), IDLRustNamespace, f.Type, ProgramRustNamespace, ToRustPascalCase(f.Type), toSnakeCase(f.Type, false))
}
func (f *OptionDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

type OptionVecSimple struct {
	ResolvedFieldTypeCommon
}

func (f *OptionVecSimple) ResolveRustType() string {
	return fmt.Sprintf("Option<Vec<%s>>", IDLTypeToRustType(f.Type))
}
func (f *OptionVecSimple) ResolveProtobufType() string {
	return fmt.Sprintf("OptionVecSimple%s", IDLTypeToProtobufType(f.Type))
}
func (f *OptionVecSimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: map_option_vec_%s(),", fieldName, toSnakeCase(f.Type, false))
}
func (f *OptionVecSimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("OptionVecSimple%s", f.Type)
	typeNameInner := fmt.Sprintf("OptionVecSimple%sInner", f.Type)
	composedType := ComposeProgramRustNamespaceType(typeName)
	composedInnerType := ComposeProgramRustNamespaceType(typeNameInner)

	return fmt.Sprintf(`
			fn map_option_vec_%s(idlType: %s) -> %s {
				if (idlType.is_none()) {
					return %s {
						inner: None
					}
				}
				
				return %s {
					inner: Some(%s {
						value: idlType.unwrap().into_inter().map(|f| f as %s).collect()
					})
				}
			}
		`, toSnakeCase(f.Type, false), f.ResolveRustType(), composedType, composedType, composedType, composedInnerType, IDLTypeToRustType(f.Type))
}
func (f *OptionVecSimple) PrintNecessaryProtobufMessages() string {
	castType := IDLTypeToProtobufType(f.Type)
	innerMessageName := fmt.Sprintf("OptionVecSimple%sInner", castType)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message OptionVecSimple%s {
			optional %s inner = 1;
		}
	`, innerMessageName, castType, castType, innerMessageName)
}

type OptionVecDefined struct {
	ResolvedFieldTypeCommon
}

func (f *OptionVecDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<Vec<%s>>", f.Type)
}
func (f *OptionVecDefined) ResolveProtobufType() string {
	return fmt.Sprintf("OptionVecDefined%s", f.Type)
}
func (f *OptionVecDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("OptionVecDefined%s", ToRustPascalCase(f.Type))
	typeNameInner := fmt.Sprintf("OptionVecDefined%sInner", ToRustPascalCase(f.Type))
	composedType := ComposeProgramRustNamespaceType(typeName)
	composedInnerType := ComposeProgramRustNamespaceType(typeNameInner)

	return fmt.Sprintf(`
			fn map_option_vec_%s(idlType: %s) -> %s {
				if (idlType.is_none()) {
					return %s {
						inner: None
					}
				}
				
				return %s {
					inner: Some(%s {
						value: idlType.unwrap().into_inter().map(|f| f as map_defined_%s(f)).collect()
					})
				}
			}
		`, toSnakeCase(f.Type, false), f.ResolveRustType(), composedType, composedType, composedType, composedInnerType, toSnakeCase(f.Type, false))
}
func (f *OptionVecDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: map_option_vec_%s(%s.%s),", fieldName, toSnakeCase(f.Type, false), variableName, fieldName)
}
func (f *OptionVecDefined) PrintNecessaryProtobufMessages() string {
	messageName := fmt.Sprintf("OptionVecDefined%s", f.Type)
	innerMessageName := fmt.Sprintf("OptionVecDefined%sInner", f.Type)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message %s {
			optional %s inner = 1;
		}
	`, innerMessageName, f.Type, messageName, innerMessageName)
}

type OptionArraySimple struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *OptionArraySimple) ResolveRustType() string {
	return fmt.Sprintf("Option<[%s;%d]>", IDLTypeToRustType(f.Type), f.Length)
}
func (f *OptionArraySimple) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArraySimple%s", IDLTypeToProtobufType(f.Type))
}
func (f *OptionArraySimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: map_option_array_%s(%s.%s),", fieldName, toSnakeCase(IDLTypeToProtobufType(f.Type), false), variableName, fieldName)
}
func (f *OptionArraySimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("OptionArraySimple%s", IDLTypeToProtobufType(f.Type))
	typeNameInner := fmt.Sprintf("OptionArraySimple%sInner", IDLTypeToProtobufType(f.Type))
	composedType := ComposeProgramRustNamespaceType(typeName)
	composedInnerType := ComposeProgramRustNamespaceType(typeNameInner)

	return fmt.Sprintf(`
			fn map_option_array_%s(idlType: %s) -> %s {
				if (idlType.is_none()) {
					return %s {
						inner: None
					}
				}
				
				return %s {
					inner: Some(%s {
						value: idlType.unwrap().into_inter().map(|f| f as %s).collect()
					})
				}
			}
		`, toSnakeCase(IDLTypeToProtobufType(f.Type), false), f.ResolveRustType(), composedType, composedType, composedType, composedInnerType, IDLTypeToRustType(f.Type))
}
func (f *OptionArraySimple) PrintNecessaryProtobufMessages() string {
	castType := IDLTypeToProtobufType(f.Type)
	innerMessageName := fmt.Sprintf("OptionArraySimple%sInner", castType)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message OptionArraySimple%s {
			optional %s inner = 1;
		}
	`, innerMessageName, castType, castType, innerMessageName)
}

type OptionArrayDefined struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *OptionArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<[%s;%d]>", f.Type, f.Length)
}
func (f *OptionArrayDefined) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArrayDefined%s", f.Type)
}
func (f *OptionArrayDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: map_option_array_%s(%s.%s),", fieldName, toSnakeCase(f.Type, false), variableName, fieldName)
}
func (f *OptionArrayDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("OptionArrayDefined%s", ToRustPascalCase(f.Type))
	typeNameInner := fmt.Sprintf("OptionArrayDefined%sInner", ToRustPascalCase(f.Type))
	composedType := ComposeProgramRustNamespaceType(typeName)
	composedInnerType := ComposeProgramRustNamespaceType(typeNameInner)

	return fmt.Sprintf(`
			fn map_option_array_%s(idlType: %s) -> %s {
				if (idlType.is_none()) {
					return %s {
						inner: None
					}
				}
				
				return %s {
					inner: Some(%s {
						value: idlType.unwrap().into_inter().map(|f| f as map_defined_%s(f)).collect()
					})
				}
			}
		`, toSnakeCase(f.Type, false), f.ResolveRustType(), composedType, composedType, composedType, composedInnerType, toSnakeCase(f.Type, false))
}
func (f *OptionArrayDefined) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionArrayDefined%sInner", f.Type)
	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message OptionArrayDefined%s {
			optional %s inner = 1;
		}
	`, innerMessageName, f.Type, f.Type, innerMessageName)
}

type OptionArrayArraySimple struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *OptionArrayArraySimple) ResolveRustType() string {
	return fmt.Sprintf("Option<[[%s;%d];%d]>", IDLTypeToRustType(f.Type), f.Length, f.OuterLength)
}
func (f *OptionArrayArraySimple) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArrayArraySimple%s", IDLTypeToProtobufType(f.Type))
}
func (f *OptionArrayArraySimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf(
		"%s: map_option_array_array_%s(%s.%s),",
		fieldName,
		toSnakeCase(IDLTypeToProtobufType(f.Type), false),
		variableName,
		fieldName,
	)
}

func (f *OptionArrayArraySimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("OptionArrayArraySimple%s", IDLTypeToProtobufType(f.Type))
	typeNameInner := fmt.Sprintf("OptionArrayArraySimple%sInner", IDLTypeToProtobufType(f.Type))
	typeNameInnerArray := fmt.Sprintf("OptionArrayArraySimple%sInnerArray", IDLTypeToProtobufType(f.Type))

	composedType := ComposeProgramRustNamespaceType(typeName)
	composedInnerType := ComposeProgramRustNamespaceType(typeNameInner)
	composedInnerArrayType := ComposeProgramRustNamespaceType(typeNameInnerArray)

	return fmt.Sprintf(`
fn map_option_array_array_%s(idltype: Option<[[%s; %d]; %d]>) -> %s {
    if idlType.is_none() {
        return %s {
            inner: None,
        };
    }

    let inner = idltype.unwrap().into_iter().map(|inner_array| {
        %s {
            array_inner: inner_array.into_iter().map(|f| *f as %s).collect(),
        }
    }).collect();

    %s {
        inner: Some(%s {
            inner,
        }),
    }
}
`, toSnakeCase(f.Type, false), IDLTypeToRustType(f.Type), f.Length, f.OuterLength,
		composedType,
		composedType,
		composedInnerArrayType,
		IDLTypeToRustType(f.Type),
		composedType,
		composedInnerType)
}
func (f *OptionArrayArraySimple) PrintNecessaryProtobufMessages() string {
	castType := IDLTypeToProtobufType(f.Type)
	innerMessageName := fmt.Sprintf("OptionArrayArraySimple%sInner", castType)
	innerArrayMessageName := fmt.Sprintf("OptionArrayArraySimple%sInnerArray", castType)

	return fmt.Sprintf(`
		message %s {
			repeated %s inner = 1;
		}

		message %s {
			repeated %s arrayInner = 1;
		}

		message OptionArrayArraySimple%s {
			optional %s inner = 1;
		}
	`, innerArrayMessageName, castType, innerMessageName, innerArrayMessageName, castType, innerMessageName)
}

type OptionArrayArrayDefined struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *OptionArrayArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("Option<[[%s;%d];%d]>", f.Type, f.Length, f.OuterLength)
}
func (f *OptionArrayArrayDefined) ResolveProtobufType() string {
	return fmt.Sprintf("OptionArrayArrayDefined%s", f.Type)
}
func (f *OptionArrayArrayDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf(
		"%s: map_option_array_array_%s(%s.%s),",
		fieldName,
		toSnakeCase(f.Type, false),
		variableName,
		fieldName,
	)
}
func (f *OptionArrayArrayDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("OptionArrayArrayDefined%s", ToRustPascalCase(f.Type))
	typeNameInner := fmt.Sprintf("OptionArrayArrayDefined%sInner", ToRustPascalCase(f.Type))
	typeNameInnerArray := fmt.Sprintf("OptionArrayArrayDefined%sInnerArray", ToRustPascalCase(f.Type))

	composedType := ComposeProgramRustNamespaceType(typeName)
	composedInnerType := ComposeProgramRustNamespaceType(typeNameInner)
	composedInnerArrayType := ComposeProgramRustNamespaceType(typeNameInnerArray)

	return fmt.Sprintf(`
fn map_option_array_array_%s(idltype: Option<[[%s; %d]; %d]>) -> %s {
	if idlType.is_none() {
		return %s {
			inner: None,
		};
	}

	let inner = idltype.unwrap().into_iter().map(|inner_array| {
		%s {
			array_inner: inner_array.into_iter().map(|f| map_defined_%s(f)).collect(),
		}
	}).collect();

	%s {
		inner: Some(%s {
			inner,
		}),
	}
}
	`,
		toSnakeCase(f.Type, false), f.Type, f.Length, f.OuterLength,
		composedType,
		composedType,
		composedInnerArrayType, toSnakeCase(f.Type, false),
		composedType, composedInnerType)
}

func (f *OptionArrayArrayDefined) PrintNecessaryProtobufMessages() string {
	innerMessageName := fmt.Sprintf("OptionArrayArrayDefined%sInner", f.Type)
	innerArrayMessageName := fmt.Sprintf("OptionArrayArrayDefined%sInnerArray", f.Type)

	return fmt.Sprintf(`
		message %s {
			repeated %s inner = 1;
		}

		message %s {
			repeated %s arrayInner = 1;
		}

		message OptionArrayArrayDefined%s {
			optional %s inner = 1;
		}
	`, innerArrayMessageName, f.Type, innerMessageName, innerArrayMessageName, f.Type, innerMessageName)
}

// array
type ArraySimple struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *ArraySimple) ResolveRustType() string {
	return fmt.Sprintf("[%s;%d]", IDLTypeToRustType(f.Type), f.Length)
}
func (f *ArraySimple) ResolveProtobufType() string {
	return fmt.Sprintf("repeated %s", IDLTypeToProtobufType(f.Type))
}
func (f *ArraySimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *ArraySimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf("%s: %s.%s.to_vec(),", fieldName, variableName, fieldName)
}
func (f *ArraySimple) PrintNecessaryProtobufMessages() string {
	return ""
}

type ArrayDefined struct {
	ResolvedFieldTypeCommon
	Length int
}

func (f *ArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("[%s;%d]", f.Type, f.Length)
}
func (f *ArrayDefined) ResolveProtobufType() string {
	return fmt.Sprintf("repeated %s", f.Type)
}
func (f *ArrayDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	return ""
}
func (f *ArrayDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	return fmt.Sprintf(
		"%s: %s.%s.into_iter().map(|f| map_defined_%s(f)).collect(),",
		fieldName,
		variableName,
		fieldName,
		toSnakeCase(f.Type, false),
	)
}
func (f *ArrayDefined) PrintNecessaryProtobufMessages() string {
	return ""
}

type ArrayArraySimple struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *ArrayArraySimple) ResolveRustType() string {
	return fmt.Sprintf("[[%s;%d];%d]", IDLTypeToRustType(f.Type), f.Length, f.OuterLength)
}
func (f *ArrayArraySimple) ResolveProtobufType() string {
	return fmt.Sprintf("ArrayArraySimple%s", IDLTypeToProtobufType(f.Type))
}
func (f *ArrayArraySimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	snakeType := toSnakeCase(f.Type, false)
	mapperFunc := fmt.Sprintf("map_array_array_%s", snakeType)
	return fmt.Sprintf("%s: Some(%s(%s.%s)),", fieldName, mapperFunc, variableName, fieldName)
}
func (f *ArrayArraySimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	snakeType := toSnakeCase(f.Type, false)
	funcName := fmt.Sprintf("map_array_array_%s", snakeType)
	protobufType := fmt.Sprintf("ArrayArraySimple%s", IDLTypeToProtobufType(f.Type))
	innerProtobufType := fmt.Sprintf("ArrayArraySimple%sInner", IDLTypeToProtobufType(f.Type))

	return fmt.Sprintf(`
fn %s<const M: usize, const N: usize>(input: [[%s; N]; M]) -> %s {
    %s {
        inner: input.iter()
            .map(|inner_arr| %s {
                value: inner_arr.iter().map(|f| *f as %s).collect()
            })
            .collect()
    }
}
`,
		funcName,
		IDLTypeToRustType(f.Type),
		ComposeProgramRustNamespaceType(protobufType),
		ComposeProgramRustNamespaceType(protobufType),
		ComposeProgramRustNamespaceType(innerProtobufType),
		IDLTypeToRustType(f.Type),
	)
}
func (f *ArrayArraySimple) PrintNecessaryProtobufMessages() string {
	castType := IDLTypeToProtobufType(f.Type)
	messageName := fmt.Sprintf("ArrayArraySimple%s", castType)
	innerMessageName := fmt.Sprintf("ArrayArraySimple%sInner", castType)

	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message %s {
			repeated %s inner = 1;
		}
	`, innerMessageName, castType, messageName, innerMessageName)
}

type ArrayArrayDefined struct {
	ResolvedFieldTypeCommon
	Length      int
	OuterLength int
}

func (f *ArrayArrayDefined) ResolveRustType() string {
	return fmt.Sprintf("[[%s;%d];%d]", f.Type, f.Length, f.OuterLength)
}
func (f *ArrayArrayDefined) ResolveProtobufType() string {
	return fmt.Sprintf("ArrayArrayDefined%s", f.Type)
}
func (f *ArrayArrayDefined) PrintNecessaryRustStructs(fieldName string, types []Type) string {
	typeName := fmt.Sprintf("ArrayArrayDefined%s", ToRustPascalCase(f.Type))
	innerMessage := fmt.Sprintf("ArrayArrayDefined%sInner", ToRustPascalCase(f.Type))
	namespace := ComposeProgramRustNamespaceType(typeName)
	namespaceInner := ComposeProgramRustNamespaceType(innerMessage)
	mapFunc := fmt.Sprintf("map_array_array_%s", toSnakeCase(f.Type, false))

	return fmt.Sprintf(`
fn %s<const M: usize, const N: usize>(input: [[%s; N]; M]) -> %s {
    %s {
        inner: input.iter()
            .map(|inner_arr| %s {
                value: inner_arr.iter()
                    .map(|f| map_defined_%s(f))
                    .collect()
            })
            .collect()
    }
}
`, mapFunc, f.Type, namespace,
		namespace, namespaceInner, toSnakeCase(f.Type, false))
}
func (f *ArrayArrayDefined) PrintRustMappings(fieldName string, variableName string, types []Type) string {
	snakeType := toSnakeCase(f.Type, false)
	mapperFunc := fmt.Sprintf("map_array_array_%s", snakeType)
	return fmt.Sprintf("%s: %s(%s.%s),", fieldName, mapperFunc, variableName, fieldName)
}
func (f *ArrayArrayDefined) PrintNecessaryProtobufMessages() string {
	messageName := fmt.Sprintf("ArrayArrayDefined%s", f.Type)
	innerMessageName := fmt.Sprintf("ArrayArrayDefined%sInner", f.Type)

	return fmt.Sprintf(`
		message %s {
			repeated %s value = 1;
		}

		message %s {
			repeated %s inner = 1;
		}
	`, innerMessageName, f.Type, messageName, innerMessageName)
}

/*
func (f *FieldType) ResolveRustType() string {
	if f.IsSimple() {
		return f.Simple
	}

	if f.IsSimplePubKey() {
		return "PubKey"
	}

	if f.IsDefined() {
		return f.Defined
	}

	// Vec
	if f.IsVec() {
		if f.Vec.Type.IsSimple() {
			return fmt.Sprintf("Vec<%s>", CastInRustIfNeeded(f.Vec.Type.Simple))
		}
		if f.Vec.Type.IsDefined() {
			return fmt.Sprintf("Vec<%s>", f.Vec.Type.Defined)
		}
		if f.Vec.Type.IsOption() {
			if f.Vec.Type.Option.Type.IsSimple() {
				return fmt.Sprintf("Vec<Option<%s>>", CastInRustIfNeeded(f.Vec.Type.Option.Type.Simple))
			}
			if f.Vec.Type.Option.Type.IsDefined() {
				return fmt.Sprintf("Vec<Option<%s>>", f.Vec.Type.Option.Type.Defined)
			}
		}
	}

	// Option
	if f.IsOption() {
		if f.Option.Type.IsSimple() {
			return fmt.Sprintf("Option<%s>", CastInRustIfNeeded(f.Option.Type.Simple))
		}
		if f.Option.Type.IsDefined() {
			return fmt.Sprintf("Option<%s>", f.Option.Type.Defined)
		}
		if f.Option.Type.IsVec() {
			if f.Option.Type.Vec.Type.IsSimple() {
				return fmt.Sprintf("Option<Vec<%s>>", CastInRustIfNeeded(f.Option.Type.Vec.Type.Simple))
			}
			if f.Option.Type.Vec.Type.IsDefined() {
				return fmt.Sprintf("Option<Vec<%s>>", f.Option.Type.Vec.Type.Defined)
			}
		}
		if f.Option.Type.IsArray() {
			if f.Option.Type.Array.Type.IsSimple() {
				return fmt.Sprintf("Option<[%s;%d]>", CastInRustIfNeeded(f.Option.Type.Array.Type.Simple), f.Option.Type.Array.Length)
			}
			if f.Option.Type.Array.Type.IsDefined() {
				return fmt.Sprintf("Option<[%s;%d]>", f.Option.Type.Array.Type.Defined, f.Option.Type.Array.Length)
			}
			if f.Option.Type.Array.Type.IsArray() {
				if f.Option.Type.Array.Type.Array.Type.IsSimple() {
					return fmt.Sprintf("Option<[[%s;%d];%d]>", CastInRustIfNeeded(f.Option.Type.Array.Type.Array.Type.Simple), f.Option.Type.Array.Type.Array.Length, f.Option.Type.Array.Length)
				}
				if f.Option.Type.Array.Type.Array.Type.IsDefined() {
					return fmt.Sprintf("Option<[[%s;%d];%d]>", f.Option.Type.Array.Type.Array.Type.Defined, f.Option.Type.Array.Type.Array.Length, f.Option.Type.Array.Length)
				}
			}
		}
	}

	// Array
	if f.IsArray() {
		if f.Array.Type.IsSimple() {
			return fmt.Sprintf("[%s;%d]", CastInRustIfNeeded(f.Array.Type.Simple), f.Array.Length)
		}
		if f.Array.Type.IsDefined() {
			return fmt.Sprintf("[%s;%d]", f.Array.Type.Defined, f.Array.Length)
		}
		if f.Array.Type.IsArray() {
			if f.Array.Type.Array.Type.IsSimple() {
				return fmt.Sprintf("[[%s;%d];%d]", CastInRustIfNeeded(f.Array.Type.Array.Type.Simple), f.Array.Type.Array.Length, f.Array.Length)
			}
			if f.Array.Type.Array.Type.IsDefined() {
				return fmt.Sprintf("[[%s;%d];%d]", f.Array.Type.Array.Type.Defined, f.Array.Type.Array.Length, f.Array.Length)
			}
		}
	}

	return ""
}*/
