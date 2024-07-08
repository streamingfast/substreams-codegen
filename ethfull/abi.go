package ethfull

import (
	"encoding/hex"
	"fmt"
	"github.com/golang-cz/textcase"
	"github.com/huandu/xstrings"
	"sort"
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/streamingfast/eth-go"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type ABI struct {
	abi *eth.ABI
	raw string
}

func cmdDecodeABI(contract *Contract) loop.Cmd {
	return func() loop.Msg {
		abi, err := eth.ParseABIFromBytes([]byte(contract.RawABI))
		return ReturnRunDecodeContractABI{abi: &ABI{abi, string(contract.RawABI)}, err: err}
	}
}

func cmdDecodeDynamicABI(contract *DynamicContract) loop.Cmd {
	return func() loop.Msg {
		abi, err := eth.ParseABIFromBytes([]byte(contract.RawABI))
		return ReturnRunDecodeDynamicContractABI{abi: &ABI{abi, string(contract.RawABI)}, err: err}
	}
}

var pluralizerSingleton = pluralize.NewClient()

func (a *ABI) EventIDsToSig() (out map[string]string) {
	abi := a.abi
	names := maps.Keys(abi.LogEventsByNameMap)
	sort.StringSlice(names).Sort()

	out = make(map[string]string)
	for _, name := range names {
		for _, event := range abi.FindLogsByName(name) {
			id := hex.EncodeToString(event.LogID())

			out[id] = event.Signature()
		}
	}

	return
}

func (a *ABI) BuildEventModels() (out []codegenEvent, err error) {
	abi := a.abi

	names := maps.Keys(abi.LogEventsByNameMap)
	sort.StringSlice(names).Sort()

	// We allocate as many names + 16 to potentially account for duplicates
	out = make([]codegenEvent, 0, len(names)+16)
	for _, name := range names {
		events := abi.FindLogsByName(name)

		for i, event := range events {
			rustABIStructName := name
			if len(events) > 1 { // will result in OriginalName, OriginalName1, OriginalName2
				rustABIStructName = name + strconv.FormatUint(uint64(i+1), 10)
			}
			for i, param := range event.Parameters {
				if param.Name == "" {
					if event.Parameters[i].Indexed {
						param.Name = fmt.Sprintf("topic%d", i)
					} else {
						param.Name = fmt.Sprintf("param%d", i)
					}
				}
			}
			event.Signature()

			// Sanitize abi struct name base on rust proto-gen sanitizer
			rustABIStructName = sanitizeABIStructName(rustABIStructName)

			protoFieldName := xstrings.ToSnakeCase(pluralizerSingleton.Plural(rustABIStructName))
			// prost will do a to_lower_camel_case() on any struct name
			rustGeneratedStructName := textcase.PascalCase(xstrings.ToSnakeCase(rustABIStructName))

			eventID := hex.EncodeToString(event.LogID())

			codegenEvent := codegenEvent{
				Rust: &rustEventModel{
					ABIStructName:                             rustGeneratedStructName,
					ProtoMessageName:                          rustGeneratedStructName,
					ProtoOutputModuleFieldName:                protoFieldName,
					ProtoOutputModuleFieldSubgraphTriggerName: pluralizerSingleton.Plural(rustGeneratedStructName),
					TableChangeEntityName:                     xstrings.ToSnakeCase(rustABIStructName),
				},

				Proto: &protoEventModel{
					MessageName:           rustGeneratedStructName,
					MessageHash:           eventID,
					OutputModuleFieldName: protoFieldName,
				},
			}

			if err := codegenEvent.Rust.populateFields(event); err != nil {
				return nil, fmt.Errorf("populating codegen Rust fields: %w", err)
			}

			if err := codegenEvent.Proto.populateFields(event); err != nil {
				return nil, fmt.Errorf("populating codegen Proto fields: %w", err)
			}

			out = append(out, codegenEvent)
		}
	}

	return
}

func (a *ABI) BuildCallModels() (out []codegenCall, err error) {
	abi := a.abi
	names := maps.Keys(abi.FunctionsByNameMap)
	sort.StringSlice(names).Sort()

	// We allocate as many names + 16 to potentially account for duplicates
	out = make([]codegenCall, 0, len(names)+16)
	for _, name := range names {
		calls := abi.FindFunctionsByName(name)

		for i, call := range calls {
			// We skip "pure" and "view" functions because they don't affect the state of the chain
			if call.StateMutability == eth.StateMutabilityPure || call.StateMutability == eth.StateMutabilityView {
				continue
			}
			rustABIStructName := name
			if len(calls) > 1 { // will result in OriginalName, OriginalName1, OriginalName2
				rustABIStructName = name + strconv.FormatUint(uint64(i+1), 10)
			}
			for i, param := range call.Parameters {
				if param.Name == "" {
					param.Name = fmt.Sprintf("param%d", i)
				}
			}
			for i, param := range call.ReturnParameters {
				if param.Name == "" {
					param.Name = fmt.Sprintf("param%d", i)
				}
			}

			// Sanitize abi struct name base on rust proto-gen sanitizer
			rustABIStructName = sanitizeABIStructName(rustABIStructName)

			protoFieldName := "call_" + xstrings.ToSnakeCase(pluralizerSingleton.Plural(rustABIStructName))

			// prost will do a to_lower_camel_case() on any struct name
			rustGeneratedStructName := textcase.PascalCase(xstrings.ToSnakeCase(rustABIStructName))
			protoMessageName := textcase.PascalCase(xstrings.ToSnakeCase(rustABIStructName) + "Call")

			codegenCall := codegenCall{
				Rust: &rustCallModel{
					ABIStructName:                             rustGeneratedStructName,
					ProtoMessageName:                          protoMessageName,
					ProtoOutputModuleFieldName:                protoFieldName,
					ProtoOutputModuleFieldSubgraphTriggerName: fmt.Sprintf("Call%s", pluralizerSingleton.Plural(rustGeneratedStructName)),
					TableChangeEntityName:                     "call_" + xstrings.ToSnakeCase(rustABIStructName),
				},

				Proto: &protoCallModel{
					MessageName:           protoMessageName,
					OutputModuleFieldName: protoFieldName,
				},
			}

			if err := codegenCall.Rust.populateFields(call); err != nil {
				return nil, fmt.Errorf("populating codegen Rust fields: %w", err)
			}

			if err := codegenCall.Proto.populateFields(call); err != nil {
				return nil, fmt.Errorf("populating codegen Proto fields: %w", err)
			}

			out = append(out, codegenCall)
		}
	}

	return
}

type codegenEvent struct {
	Rust  *rustEventModel
	Proto *protoEventModel
}

type codegenCall struct {
	Rust  *rustCallModel
	Proto *protoCallModel
}

type rustEventModel struct {
	ABIStructName                             string
	ProtoMessageName                          string
	ProtoOutputModuleFieldName                string
	ProtoOutputModuleFieldSubgraphTriggerName string
	TableChangeEntityName                     string
	ProtoFieldABIConversionMap                map[string]string
	ProtoFieldTableChangesMap                 map[string]tableChangeSetField
	ProtoFieldSqlmap                          map[string]string
	ProtoFieldClickhouseMap                   map[string]string
	ProtoFieldGraphQLMap                      map[string]string
	ProtoFieldSubgraphMappings                map[string]string
}

type rustCallModel struct {
	ABIStructName                             string
	ProtoMessageName                          string
	ProtoOutputModuleFieldName                string
	ProtoOutputModuleFieldSubgraphTriggerName string
	TableChangeEntityName                     string
	OutputFieldsString                        string
	ProtoFieldABIConversionMap                map[string]string
	ProtoFieldTableChangesMap                 map[string]tableChangeSetField
	ProtoFieldSqlmap                          map[string]string
	ProtoFieldClickhouseMap                   map[string]string
	ProtoFieldGraphQLMap                      map[string]string
	ProtoFieldSubgraphMappings                map[string]string
}

type tableChangeSetField struct {
	Setter          string
	ValueAccessCode string
}

func (e *rustEventModel) populateFields(log *eth.LogEventDef) error {
	if len(log.Parameters) == 0 {
		return nil
	}

	e.ProtoFieldABIConversionMap = map[string]string{}
	e.ProtoFieldTableChangesMap = map[string]tableChangeSetField{}
	e.ProtoFieldSqlmap = map[string]string{}
	e.ProtoFieldClickhouseMap = map[string]string{}
	e.ProtoFieldGraphQLMap = map[string]string{}
	e.ProtoFieldSubgraphMappings = map[string]string{}
	paramNames := make([]string, len(log.Parameters))
	for i := range log.Parameters {
		paramNames[i] = log.Parameters[i].Name
	}

	zlog.Info("Generating ABI Events", zap.String("name", log.Name), zap.String("param_names", strings.Join(paramNames, ",")))

	for _, parameter := range log.Parameters {
		name := codegen.SanitizeProtoFieldName(parameter.Name)
		name = xstrings.ToSnakeCase(name)

		toProtoCode := generateFieldTransformCode(parameter.Type, "event."+name, false)
		if toProtoCode == SKIP_FIELD {
			continue
		}
		if toProtoCode == "" {
			return fmt.Errorf("transform - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		toDatabaseChangeSetter, toDatabaseChangeCode := generateFieldTableChangeCode(parameter.Type, "evt."+name, true)
		if toDatabaseChangeCode == SKIP_FIELD {
			continue
		}
		if toDatabaseChangeSetter == "" {
			return fmt.Errorf("table change - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		toSqlCode := generateFieldSqlTypes(parameter.Type)
		if toSqlCode == SKIP_FIELD {
			continue
		}
		if toSqlCode == "" {
			return fmt.Errorf("sql - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		toClickhouseCode := generateFieldClickhouseTypes(parameter.Type)
		if toClickhouseCode == SKIP_FIELD {
			continue
		}
		if toClickhouseCode == "" {
			return fmt.Errorf("clickhouse - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		toGraphQLCode := generateFieldGraphQLTypes(parameter.Type)
		if toGraphQLCode == "" {
			return fmt.Errorf("graphql - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		toSubgraphMappingCode := generateFieldSubgraphMappingCode(name, true)
		if toSubgraphMappingCode == "" {
			return fmt.Errorf("subgraph trigger mappings - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		columnName := sanitizeTableChangesColumnNames(name)

		e.ProtoFieldABIConversionMap[name] = toProtoCode
		e.ProtoFieldTableChangesMap[name] = tableChangeSetField{Setter: toDatabaseChangeSetter, ValueAccessCode: toDatabaseChangeCode}
		e.ProtoFieldSqlmap[columnName] = toSqlCode
		e.ProtoFieldClickhouseMap[columnName] = toClickhouseCode
		e.ProtoFieldGraphQLMap[name] = toGraphQLCode
		e.ProtoFieldSubgraphMappings[strcase.ToLowerCamel(name)] = toSubgraphMappingCode
	}

	return nil
}

func convertMethodParameters(parameters []*eth.MethodParameter, optionalPrefix string, isEvent bool) (
	tableChangesMap map[string]tableChangeSetField,
	sqlMap map[string]string,
	clickhouseMap map[string]string,
	graphqlMap map[string]string,
	subgraphMappingMap map[string]string,
	err error,
) {
	tableChangesMap = map[string]tableChangeSetField{}
	sqlMap = map[string]string{}
	clickhouseMap = map[string]string{}
	graphqlMap = map[string]string{}
	subgraphMappingMap = map[string]string{}

	for _, parameter := range parameters {
		name := optionalPrefix + xstrings.ToSnakeCase(parameter.Name)
		name = codegen.SanitizeProtoFieldName(name)
		columnName := sanitizeTableChangesColumnNames(name)

		toDatabaseChangeSetter, toDatabaseChangeCode := generateFieldTableChangeCode(parameter.Type, "call."+name, true)
		if toDatabaseChangeCode != SKIP_FIELD {
			if toDatabaseChangeSetter == "" {
				err = fmt.Errorf("table change - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
				return
			}
			tableChangesMap[name] = tableChangeSetField{Setter: toDatabaseChangeSetter, ValueAccessCode: toDatabaseChangeCode}
		}

		toSqlCode := generateFieldSqlTypes(parameter.Type)
		if toSqlCode != SKIP_FIELD {
			if toSqlCode == "" {
				err = fmt.Errorf("sql - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
				return
			}
			sqlMap[columnName] = toSqlCode
		}

		toClickhouseCode := generateFieldClickhouseTypes(parameter.Type)
		if toClickhouseCode != SKIP_FIELD {
			if toClickhouseCode == "" {
				err = fmt.Errorf("clickhouse - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
				return
			}
			clickhouseMap[columnName] = toClickhouseCode
		}

		toGraphQLCode := generateFieldGraphQLTypes(parameter.Type)
		if toGraphQLCode != SKIP_FIELD {
			if toGraphQLCode == "" {
				err = fmt.Errorf("graphql - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
				return
			}
			graphqlMap[name] = toGraphQLCode
		}

		toSubgraphMappingCode := generateFieldSubgraphMappingCode(name, isEvent)
		if toSubgraphMappingCode != "" {
			subgraphMappingMap[strcase.ToLowerCamel(name)] = toSubgraphMappingCode
		}
	}
	return
}

func methodToABIConversionMaps(
	parameters []*eth.MethodParameter,
	outputParameters []*eth.MethodParameter,
) (
	abiConversionMap map[string]string,
	outputString string,
	err error,
) {

	if len(parameters) != 0 || len(outputParameters) != 0 {
		abiConversionMap = make(map[string]string)
	}
	for _, parameter := range parameters {
		name := codegen.SanitizeProtoFieldName(parameter.Name)
		name = xstrings.ToSnakeCase(name)

		toProtoCode := generateFieldTransformCode(parameter.Type, "decoded_call."+name, false)
		if toProtoCode != SKIP_FIELD {
			if toProtoCode == "" {
				err = fmt.Errorf("transform - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
				return
			}
			abiConversionMap[name] = toProtoCode
		}
	}

	if len(outputParameters) == 0 {
		return
	}

	outputNames := make([]string, len(outputParameters))
	for i, parameter := range outputParameters {
		name := "output_" + xstrings.ToSnakeCase(parameter.Name)
		name = codegen.SanitizeProtoFieldName(name)
		outputNames[i] = name

		toProtoCode := generateFieldTransformCode(parameter.Type, name, false)
		if toProtoCode != SKIP_FIELD {
			if toProtoCode == "" {
				err = fmt.Errorf("transform - field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
				return
			}
			abiConversionMap[name] = toProtoCode
		}
	}
	if len(outputNames) == 1 {
		outputString = strings.Join(outputNames, ", ")
	} else {
		outputString = "(" + strings.Join(outputNames, ", ") + ")"
	}

	return
}

func (e *rustCallModel) populateFields(call *eth.MethodDef) error {
	if len(call.Parameters) == 0 && call.ReturnParameters == nil {
		return nil
	}

	paramNames := make([]string, len(call.Parameters))
	for i := range call.Parameters {
		paramNames[i] = call.Parameters[i].Name
	}
	outputParamNames := make([]string, len(call.ReturnParameters))
	for i := range call.ReturnParameters {
		outputParamNames[i] = call.ReturnParameters[i].Name
	}

	zlog.Info("Generating ABI Calls", zap.String("name", call.Name), zap.String("param_names", strings.Join(paramNames, ",")), zap.String("output_param_names", strings.Join(outputParamNames, ",")))

	var err error
	e.ProtoFieldTableChangesMap, e.ProtoFieldSqlmap, e.ProtoFieldClickhouseMap, e.ProtoFieldGraphQLMap, e.ProtoFieldSubgraphMappings, err = convertMethodParameters(call.Parameters, "", false)
	if err != nil {
		return err
	}

	outputTableChanges, outputSql, outputClickhouse, outputGraphQL, subgraphMappingMap, err := convertMethodParameters(call.ReturnParameters, "output_", false)
	if err != nil {
		return err
	}
	for k, v := range outputTableChanges {
		e.ProtoFieldTableChangesMap[k] = v
	}
	for k, v := range outputSql {
		e.ProtoFieldSqlmap[k] = v
	}
	for k, v := range outputClickhouse {
		e.ProtoFieldClickhouseMap[k] = v
	}
	for k, v := range outputGraphQL {
		e.ProtoFieldGraphQLMap[k] = v
	}
	for k, v := range subgraphMappingMap {
		e.ProtoFieldSubgraphMappings[k] = v
	}

	e.ProtoFieldABIConversionMap, e.OutputFieldsString, err = methodToABIConversionMaps(call.Parameters, call.ReturnParameters)

	return err
}

type protoEventModel struct {
	// MessageName is the name of the message representing this specific event
	MessageName string
	MessageHash string

	OutputModuleFieldName string
	Fields                []protoField
}

type protoCallModel struct {
	// MessageName is the name of the message representing this specific call
	MessageName string

	OutputModuleFieldName string
	Fields                []protoField
}

func (e *protoEventModel) populateFields(log *eth.LogEventDef) error {
	if len(log.Parameters) == 0 {
		return nil
	}

	e.Fields = make([]protoField, 0, len(log.Parameters))
	for _, parameter := range log.Parameters {
		fieldName := codegen.SanitizeProtoFieldName(parameter.Name)
		fieldName = xstrings.ToSnakeCase(fieldName)

		fieldType := getProtoFieldType(parameter.Type)
		if fieldType == SKIP_FIELD {
			continue
		}

		if fieldType == "" {
			return fmt.Errorf("field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		e.Fields = append(e.Fields, protoField{Name: fieldName, Type: fieldType})
	}

	return nil
}

func (e *protoCallModel) populateFields(call *eth.MethodDef) error {
	if len(call.Parameters) == 0 && len(call.ReturnParameters) == 0 {
		return nil
	}

	e.Fields = make([]protoField, 0, len(call.Parameters)+len(call.ReturnParameters))

	for _, parameter := range call.Parameters {
		fieldName := codegen.SanitizeProtoFieldName(parameter.Name)
		fieldName = xstrings.ToSnakeCase(fieldName)
		fieldType := getProtoFieldType(parameter.Type)
		if fieldType == SKIP_FIELD {
			continue
		}

		if fieldType == "" {
			return fmt.Errorf("field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		e.Fields = append(e.Fields, protoField{Name: fieldName, Type: fieldType})
	}

	for _, parameter := range call.ReturnParameters {
		fieldName := xstrings.ToSnakeCase("output_" + parameter.Name)
		fieldName = codegen.SanitizeProtoFieldName(fieldName)
		fieldType := getProtoFieldType(parameter.Type)
		if fieldType == SKIP_FIELD {
			continue
		}

		if fieldType == "" {
			return fmt.Errorf("field type %q on parameter with name %q is not supported right now", parameter.TypeName, parameter.Name)
		}

		e.Fields = append(e.Fields, protoField{Name: fieldName, Type: fieldType})
	}

	return nil
}

func getProtoFieldType(solidityType eth.SolidityType) string {
	switch v := solidityType.(type) {
	case eth.AddressType, eth.BytesType, eth.FixedSizeBytesType:
		return "bytes"

	case eth.BooleanType:
		return "bool"

	case eth.StringType:
		return "string"

	case eth.SignedIntegerType:
		if v.ByteSize <= 8 {
			return "int64"
		}

		return "string"

	case eth.UnsignedIntegerType:
		if v.ByteSize <= 8 {
			return "uint64"
		}

		return "string"

	case eth.SignedFixedPointType, eth.UnsignedFixedPointType:
		return "string"

	case eth.FixedSizeArrayType:
		// Flaky, I think we should support a single level of "array"
		fieldType := getProtoFieldType(v.ElementType)
		if fieldType == SKIP_FIELD {
			return SKIP_FIELD
		}
		return "repeated " + fieldType

	case eth.ArrayType:
		// Flaky, I think we should support a single level of "array"
		fieldType := getProtoFieldType(v.ElementType)
		if fieldType == SKIP_FIELD {
			return SKIP_FIELD
		}
		return "repeated " + fieldType

	case eth.StructType:
		return SKIP_FIELD

	default:
		return ""
	}
}

type protoField struct {
	Name string
	Type string
}
