This codegen takes a JSON IDL string and generate the necessarys structs to decode the data.

## Missing

### Figure out the correct `ToRustPascalCase` function

- Problem: the conversion Protobuf -> Rust modifies the name of the structures. In the IDL, types are PascalCase, but the `prost` plugin does a weird conversion. For example:

Original Name              Prost Converted Name            Notes
SwapGooseFXV2Variant       SwapGooseFxv2Variant            FX -> Fx
SwapRaydiumCPVariant       SwapRaydiumCpVariant            CP -> Cp


## IDL Versions

The `discriminator` is an identification code for every instruction and event. When you get the Solana raw data, you use the discriminator to identify the specific instruction or event.

- Instructions: the discriminator is in the first 8 bytes of the instruction data (`[0..8]`)
- Events: the discriminator is the first 8 bytes of the matched log data.
    - First, use the `sologger_log_context` library to match only events from a specific program ID
    - Second, decode the base-64 encoded string of the log
    - Third, match against the first 8 bytes of the discriminator

Where is the discriminator? It depends:
- Older IDLs: do not contain the discriminator in the JSON IDL. Therefore, the only way to get it is by generate the Rust macro. This macro contains the actual discriminator.
- New IDLs: the discriminator is present in the JSON IDL. There's no need to use the Anchor library to generate the Rust macro. Instead, you build the Rust structures dynamically.

In order to know if it's an old or new IDL, this function in `idl.go` is used:

```go
func (i *IDL) IsOldIDLFormat() bool {
	if len(i.Instructions) > 0 {
		return len(i.Instructions[0].Discriminator) == 0
	}

	if len(i.Events) > 0 {
		return len(i.Events[0].Discriminator) == 0
	}

	return false
}
```

## Files

### `idl.go`

Contains the entrypoint of the JSON decoding

### `idl_events.go`

Contains the structs to decode an Event. Note that in older versions of Anchor, the `discriminator` field is NOT present.
If the discriminator field is NOT present, you must use the Anchor library to decode the data.

### `idl_fields.go`

Instructions and events have `fields`, where the actual data is contained. A field could be a complex nested structure (e.g. `Vec<Option<MyData>>`). In this file, you use the `FieldType` struct to identify what kind of structure is it. The following function unmarshalls the `FieldType`, and allows you to know the kind of structure.

```go
func (t *FieldType) UnmarshalJSON(data []byte) error {
	fieldType, err := unmarshallFieldType(data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Type: %s", string(data))
	}

	*t = *fieldType

	return nil
}
```

The `FieldType` contains three recursive objects (the three nested possibilties: array, vec or option). By following the recursive tree, you can figure out what kind of structure it is.

```go
type FieldType struct {
	ArrayRecursive  *FieldTypeArray
	VecRecursive    *FieldTypeVec
	OptionRecursive *FieldTypeOption

    ...
}
```

Any recursive iterative requires a "base case" (i.e. when you stop the iteration). The base cases are: `Simple` or `Defined`.

- A `Simple` type includes `u8`, `string`... In the IDL, they are represented:

```json
{
    "name": "id",
    "type": "u8"
}
```

- A `Defined` type is a custom object, which is usually defined in the `types` array of the IDL:

```json
{
    "name": "swap",
    "type": {
        "defined": {
            "name": "Swap"
        }
    }
},
```

This could lead to complex types like (`Vec<RoutePlanStep>`):

```json
{
    "name": "route_plan",
    "type": {
        "vec": {
            "defined": {
                "name": "RoutePlanStep"
            }
        }
    }
}
```

By iterating over recursively over the object, we can figure out the type. For example, in the following snippet you check if the type is a `VecDefined` struct (a vector of objects e.g., `Vec<MyData>`)

```go
if fieldType.VecRecursive.Type.IsDefined() {

}
```

For all fields, you recursively iterate until you find a match with a supported type (NOT all nested types are supported -- take a look at the definition of `FieldType` to find the supported ones). It is pretty easy to add a new supported type.

### `idl_fields_types.go`

Once you have all the fields assigned to a specific supported type (`Simple`, `Defined`, `OptionDefined`...), you need to generate the Rust and Protobuf structures, as well as the mapping functions.

Every `FieldType` implements an interface, `ResolvedFieldType`, which generate the actual template code.

```go
type ResolvedFieldType interface {
	ResolveRustType() string
	ResolveProtobufType() string
	PrintNecessaryProtobufMessages() string
	PrintNecessaryRustStructs(fieldName string, types []Type) string
	PrintRustMappings(fieldName string, variableName string, types []Type) string
}
```

For example, let's take a look at the `VecSimple` type, which is a vector of simple types (e.g. `Vec<String>`, `Vec<u8>`, etc)

```go
type VecSimple struct {
	ResolvedFieldTypeCommon
}

func (f *VecSimple) ResolveRustType() string {
    /*
        The Rust type signature
    */
	return fmt.Sprintf("Vec<%s>", IDLTypeToRustType(f.Type))
}
func (f *VecSimple) ResolveProtobufType() string {
    /*
        The Protobuf type equivalent. In this case "Vec<T>" will be converted to "repeated T"
    */
	return fmt.Sprintf("repeated %s", IDLTypeToProtobufType(f.Type))
}
func (f *VecSimple) PrintNecessaryProtobufMessages() string {
    // The Protobuf messages necessary. In this case, it's not necessary because it's simple
	return ""
}
func (f *VecSimple) PrintNecessaryRustStructs(fieldName string, types []Type) string {
    // The Rust functions necessary. In this case, it's not necessary
	return ""
}
func (f *VecSimple) PrintRustMappings(fieldName string, variableName string, types []Type) string {
    // The mapping function in Rust (Rust -> Protobuf)
	return fmt.Sprintf("%s: %s.%s.into_iter().map(|f| f as %s).collect(),", fieldName, variableName, fieldName, IDLTypeToProtobufType(f.Type))
}
```


