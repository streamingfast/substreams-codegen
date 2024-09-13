package evm_events_calls

import (
	"embed"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/streamingfast/eth-go"
	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() codegen.ReturnGenerate {
	res := codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/contract.proto.gotmpl": "proto/contract.proto",
		"src/abi/mod.rs.gotmpl":       "src/abi/mod.rs",
		"src/pb/mod.rs.gotmpl":        "src/pb/mod.rs",
		"src/lib.rs.gotmpl":           "src/lib.rs",
		"build.rs.gotmpl":             "build.rs",
		"Cargo.toml.gotmpl":           "Cargo.toml",
		"rust-toolchain.toml":         "rust-toolchain.toml",
		".gitignore.gotmpl":           ".gitignore",
		"substreams.yaml.gotmpl":      "substreams.yaml",
		"README.md.gotmpl":            "README.md",
	})
	if res.Err != nil {
		return res
	}

	for _, contract := range p.Contracts {
		res.ProjectFiles[fmt.Sprintf("abi/%s_contract.abi.json", contract.Name)] = []byte(contract.Abi.raw)
	}

	for _, dds := range p.DynamicContracts {
		res.ProjectFiles[fmt.Sprintf("abi/%s_contract.abi.json", dds.Name)] = []byte(dds.Abi.raw)
	}

	return res
}

func sanitizeTableChangesColumnNames(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

const SKIP_FIELD = "skip"

func generateFieldClickhouseTypes(fieldType eth.SolidityType) string {
	switch v := fieldType.(type) {
	case eth.AddressType:
		return "VARCHAR(40)"

	case eth.BooleanType:
		return "BOOL"

	case eth.BytesType, eth.FixedSizeBytesType, eth.StringType:
		return "TEXT"

	case eth.SignedIntegerType:
		switch {
		case v.BitsSize <= 8:
			return "Int8"
		case v.BitsSize <= 16:
			return "Int16"
		case v.BitsSize <= 32:
			return "Int32"
		case v.BitsSize <= 64:
			return "Int64"
		case v.BitsSize <= 128:
			return "Int128"
		}
		return "Int256"

	case eth.UnsignedIntegerType:
		switch {
		case v.BitsSize <= 8:
			return "UInt8"
		case v.BitsSize <= 16:
			return "UInt16"
		case v.BitsSize <= 32:
			return "UInt32"
		case v.BitsSize <= 64:
			return "UInt64"
		case v.BitsSize <= 128:
			return "UInt128"
		}
		return "UInt256"

	case eth.SignedFixedPointType:
		precision := v.Decimals
		if precision > 76 {
			precision = 76
		}
		switch {
		case v.BitsSize <= 32:
			return fmt.Sprintf("Decimal128(%d)", precision)
		case v.BitsSize <= 64:
			return fmt.Sprintf("Decimal128(%d)", precision)
		case v.BitsSize <= 128:
			return fmt.Sprintf("Decimal128(%d)", precision)
		}
		return fmt.Sprintf("Decimal256(%d)", precision)

	case eth.UnsignedFixedPointType:
		precision := v.Decimals
		if precision > 76 {
			precision = 76
		}
		switch {
		case v.BitsSize <= 31:
			return fmt.Sprintf("Decimal32(%d)", precision)
		case v.BitsSize <= 63:
			return fmt.Sprintf("Decimal64(%d)", precision)
		case v.BitsSize <= 127:
			return fmt.Sprintf("Decimal128(%d)", precision)
		}
		return fmt.Sprintf("Decimal256(%d)", precision)

	case eth.StructType, eth.FixedSizeArrayType:
		return SKIP_FIELD

	case eth.ArrayType:
		elemType := generateFieldClickhouseTypes(v.ElementType)
		if elemType == "" || elemType == SKIP_FIELD {
			return SKIP_FIELD
		}

		return fmt.Sprintf("Array(%s)", elemType)

	default:
		return ""
	}
}

func generateFieldSqlTypes(fieldType eth.SolidityType) string {
	switch v := fieldType.(type) {
	case eth.AddressType:
		return "VARCHAR(40)"

	case eth.BooleanType:
		return "BOOL"

	case eth.BytesType, eth.FixedSizeBytesType, eth.StringType:
		return "TEXT"

	case eth.SignedIntegerType:
		if v.ByteSize <= 8 {
			return "INT"
		}
		return "DECIMAL"

	case eth.UnsignedIntegerType:
		if v.ByteSize <= 8 {
			return "INT"
		}
		return "DECIMAL"

	case eth.SignedFixedPointType, eth.UnsignedFixedPointType:
		return "DECIMAL"

	case eth.StructType:
		return SKIP_FIELD

	case eth.FixedSizeArrayType:
		elemType := generateFieldSqlTypes(v.ElementType)
		if elemType == "" || elemType == SKIP_FIELD {
			return SKIP_FIELD
		}

		return elemType + "[]"
	case eth.ArrayType:
		elemType := generateFieldSqlTypes(v.ElementType)
		if elemType == "" || elemType == SKIP_FIELD {
			return SKIP_FIELD
		}

		return elemType + "[]"

	default:
		return ""
	}
}

func generateFieldTableChangeCode(fieldType eth.SolidityType, fieldAccess string, byRef bool) (setter string, valueAccessCode string) {
	switch v := fieldType.(type) {
	case eth.AddressType, eth.BytesType, eth.FixedSizeBytesType:
		return "set", fmt.Sprintf("Hex(&%s).to_string()", fieldAccess)

	case eth.BooleanType:
		return "set", fieldAccess

	case eth.StringType:
		return "set", fmt.Sprintf("&%s", fieldAccess)

	case eth.SignedIntegerType:
		if v.ByteSize <= 8 {
			return "set", fieldAccess
		}
		return "set", fmt.Sprintf("BigDecimal::from_str(&%s).unwrap()", fieldAccess)

	case eth.UnsignedIntegerType:
		if v.ByteSize <= 8 {
			return "set", fieldAccess
		}
		return "set", fmt.Sprintf("BigDecimal::from_str(&%s).unwrap()", fieldAccess)

	case eth.SignedFixedPointType, eth.UnsignedFixedPointType:
		return "set", fmt.Sprintf("BigDecimal::from_str(&%s).unwrap()", fieldAccess)

	case eth.FixedSizeArrayType:
		// FIXME: Implement multiple contract support, check what is the actual semantics there
		_, inner := generateFieldTableChangeCode(v.ElementType, "x", byRef)
		if inner == SKIP_FIELD {
			return SKIP_FIELD, SKIP_FIELD
		}

		iter := "into_iter()"
		if byRef {
			iter = "iter()"
		}

		return "set_psql_array", fmt.Sprintf("%s.%s.map(|x| %s).collect::<Vec<_>>()", fieldAccess, iter, inner)
	case eth.ArrayType:
		// FIXME: Implement multiple contract support, check what is the actual semantics there
		_, inner := generateFieldTableChangeCode(v.ElementType, "x", byRef)
		if inner == SKIP_FIELD {
			return SKIP_FIELD, SKIP_FIELD
		}

		iter := "into_iter()"
		if byRef {
			iter = "iter()"
		}

		return "set_psql_array", fmt.Sprintf("%s.%s.map(|x| %s).collect::<Vec<_>>()", fieldAccess, iter, inner)

	case eth.StructType:
		return SKIP_FIELD, SKIP_FIELD

	default:
		return "", ""
	}
}

func generateFieldTransformCode(fieldType eth.SolidityType, fieldAccess string, byRef bool) string {
	switch v := fieldType.(type) {
	case eth.AddressType:
		return fieldAccess

	case eth.BooleanType, eth.StringType:
		return fieldAccess

	case eth.BytesType:
		return fieldAccess

	case eth.FixedSizeBytesType:
		return fmt.Sprintf("Vec::from(%s)", fieldAccess)

	case eth.SignedIntegerType:
		if v.ByteSize <= 8 {
			return fmt.Sprintf("Into::<num_bigint::BigInt>::into(%s).to_i64().unwrap()", fieldAccess)
		}
		return fmt.Sprintf("%s.to_string()", fieldAccess)

	case eth.UnsignedIntegerType:
		if v.ByteSize <= 8 {
			return fmt.Sprintf("%s.to_u64()", fieldAccess)
		}
		return fmt.Sprintf("%s.to_string()", fieldAccess)

	case eth.SignedFixedPointType, eth.UnsignedFixedPointType:
		return fmt.Sprintf("%s.to_string()", fieldAccess)

	case eth.FixedSizeArrayType:
		inner := generateFieldTransformCode(v.ElementType, "x", byRef)
		if inner == SKIP_FIELD {
			fmt.Println("skip case eth.FixedSizeArrayType:")
			return SKIP_FIELD
		}

		iter := "into_iter()"
		if byRef {
			iter = "iter()"
		}

		return fmt.Sprintf("%s.%s.map(|x| %s).collect::<Vec<_>>()", fieldAccess, iter, inner)

	case eth.ArrayType:
		inner := generateFieldTransformCode(v.ElementType, "x", byRef)
		if inner == SKIP_FIELD {
			return SKIP_FIELD
		}

		iter := "into_iter()"
		if byRef {
			iter = "iter()"
		}

		return fmt.Sprintf("%s.%s.map(|x| %s).collect::<Vec<_>>()", fieldAccess, iter, inner)

	case eth.StructType:
		return SKIP_FIELD

	default:
		return ""
	}
}

func generateFieldGraphQLTypes(fieldType eth.SolidityType) string {
	switch v := fieldType.(type) {
	case eth.AddressType:
		return "String!"

	case eth.BooleanType:
		return "Boolean!"

	case eth.BytesType, eth.FixedSizeBytesType, eth.StringType:
		return "String!"

	case eth.SignedIntegerType:
		if v.ByteSize <= 8 {
			return "BigInt!"
		}
		return "BigDecimal!"

	case eth.UnsignedIntegerType:
		if v.ByteSize <= 8 {
			return "BigInt!"
		}
		return "BigDecimal!"

	case eth.SignedFixedPointType, eth.UnsignedFixedPointType:
		return "BigDecimal!"

	case eth.ArrayType:
		return "[" + generateFieldGraphQLTypes(v.ElementType) + "]!"

	case eth.FixedSizeArrayType:
		return "[" + generateFieldGraphQLTypes(v.ElementType) + "]!"

	case eth.StructType:
		return SKIP_FIELD

	default:
		return ""
	}
}

func generateFieldSubgraphMappingCode(attributeName string, isEvent bool) string {
	if isEvent {
		return fmt.Sprintf("e.%s", strcase.ToLowerCamel(attributeName))
	}

	return fmt.Sprintf("c.%s", strcase.ToLowerCamel(attributeName))
}
