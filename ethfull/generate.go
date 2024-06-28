package ethfull

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/iancoleman/strcase"
	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/eth-go"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
	"go.uber.org/zap"
)

//go:embed templates/*
var templatesFS embed.FS

func cmdGenerate(p *Project, outType outputType) loop.Cmd {
	return func() loop.Msg {
		substreamsZip, projectZip, err := p.generate(outType)
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			SubstreamsSourceZip: substreamsZip,
			ProjectZip:          projectZip,
		}
	}
}

func cmdBuild(p *Project) loop.Cmd {
	p.buildStarted = time.Now()
	p.compilingBuild = true

	return func() loop.Msg {
		buildResponseChan := make(chan *codegen.RemoteBuildState, 1)
		go func() {
			p.build(buildResponseChan)
			close(buildResponseChan)
		}()

		// go to the state of compiling build
		return codegen.CompilingBuild{
			FirstTime:       true,
			RemoteBuildChan: buildResponseChan,
		}
	}
}
func cmdBuildFailed(logs []string, err error) loop.Cmd {
	return func() loop.Msg {
		return codegen.ReturnBuild{Err: err, Logs: strings.Join(logs, "\n")}
	}
}

func cmdBuildCompleted(content *codegen.RemoteBuildState) loop.Cmd {
	return func() loop.Msg {
		return codegen.ReturnBuild{
			Err:       nil,
			Logs:      strings.Join(content.Logs, "\n"),
			Artifacts: content.Artifacts,
		}
	}
}

func (p *Project) generate(outType outputType) (substreamsZip, projectZip []byte, err error) {
	// TODO: before doing any generation, we'll want to validate
	// all data points that are going into source code.
	// We don't want some weird things getting into `build.rs`
	// and being executed server side, so we'll need pristine validation
	// of all inputs here.
	// TODO: add some checking to make sure `ParentContractName` of DynamicContract
	// do match a Contract that exists here.

	srcFiles, projFiles, err := p.Render(outType)
	if err != nil {
		return nil, nil, fmt.Errorf("rendering template: %w", err)
	}

	substreamsZip, err = codegen.ZipFiles(srcFiles)
	if err != nil {
		return nil, nil, fmt.Errorf("zipping: %w", err)
	}

	projectZip, err = codegen.ZipFiles(projFiles)
	if err != nil {
		return nil, nil, fmt.Errorf("zipping: %w", err)
	}

	return
}

func (p *Project) build(remoteBuildContentChan chan<- *codegen.RemoteBuildState) {
	cloudRunServiceURL := "localhost:9001"
	if url := os.Getenv("BUILD_SERVICE_URL"); url != "" {
		cloudRunServiceURL = url
	}

	plaintext := false
	if strings.HasPrefix(cloudRunServiceURL, "localhost") {
		plaintext = true
	}

	credsOption, err := dgrpc.WithAutoTransportCredentials(false, plaintext, false)
	if err != nil {
		// write the error to the channel and handle it on the other side
		remoteBuildContentChan <- &codegen.RemoteBuildState{
			Error: err.Error(),
		}
		return
	}

	conn, err := dgrpc.NewClientConn(cloudRunServiceURL, credsOption)
	if err != nil {
		// write the error to the channel and handle it on the other side
		remoteBuildContentChan <- &codegen.RemoteBuildState{
			Error: err.Error(),
		}
		return
	}

	defer func() {
		if err := conn.Close(); err != nil {
			zlog.Error("unable to close connection gracefully", zap.Error(err))
		}
	}()

	client := pbbuild.NewBuildServiceClient(conn)
	res, err := client.Build(context.Background(),
		&pbbuild.BuildRequest{
			SourceCode:     p.sourceZip,
			CollectPattern: "*.spkg",
			Subfolder:      "substreams",
		},
	)

	if err != nil {
		remoteBuildContentChan <- &codegen.RemoteBuildState{
			Error: err.Error(),
		}
		return
	}

	var aggregatedLogs []string
	for {
		resp, err := res.Recv()

		if resp != nil && resp.Logs != "" {
			aggregatedLogs = append(aggregatedLogs, resp.Logs)
		}

		if err != nil {
			remoteBuildContentChan <- &codegen.RemoteBuildState{
				Logs:  aggregatedLogs,
				Error: err.Error(),
			}
			return
		}
		if resp == nil {
			break
		}

		if resp.Error != "" {
			remoteBuildContentChan <- &codegen.RemoteBuildState{
				Logs:  aggregatedLogs,
				Error: resp.Error,
			}
			return
		}

		if len(resp.Artifacts) != 0 {
			remoteBuildContentChan <- &codegen.RemoteBuildState{
				Error:     resp.Error,
				Logs:      aggregatedLogs,
				Artifacts: resp.Artifacts,
			}
			return
		}

		// send the request as we go -- not used on the client yet
		remoteBuildContentChan <- &codegen.RemoteBuildState{
			Logs: []string{resp.Logs},
		}
	}
}

// use the output type form the Project to render the templates
func (p *Project) Render(outType outputType) (substreamsFiles map[string][]byte, projectFiles map[string][]byte, err error) {
	substreamsFiles = map[string][]byte{}
	projectFiles = map[string][]byte{}

	tpls, err := codegen.ParseFS(nil, templatesFS, "**/*.gotmpl")
	if err != nil {
		return nil, nil, fmt.Errorf("parse templates: %w", err)
	}

	templateFiles := map[string]string{
		"proto/contract.proto.gotmpl": "substreams/proto/contract.proto",
		"src/abi/mod.rs.gotmpl":       "substreams/src/abi/mod.rs",
		"src/pb/mod.rs.gotmpl":        "substreams/src/pb/mod.rs",
		"src/lib.rs.gotmpl":           "substreams/src/lib.rs",
		"build.rs.gotmpl":             "substreams/build.rs",
		"Cargo.toml.gotmpl":           "substreams/Cargo.toml",
		"rust-toolchain.toml":         "substreams/rust-toolchain.toml",
		".gitignore":                  "substreams/.gitignore",
	}

	switch outType {
	case outputTypeSQL:
		templateFiles["sql/substreams-Makefile.gotmpl"] = "substreams/Makefile"
		templateFiles["sql/Makefile.gotmpl"] = "Makefile"
		templateFiles["sql/substreams.yaml.gotmpl"] = "substreams/substreams.yaml"
		templateFiles["sql/dev-environment/docker-compose.yml.gotmpl"] = "dev-environment/docker-compose.yml"
		templateFiles["sql/dev-environment/start.sh.gotmpl"] = "dev-environment/start.sh"
		templateFiles["sql/README.md.gotmpl"] = "README.md"

		switch p.SqlOutputFlavor {
		case "clickhouse":
			templateFiles["sql/schema.clickhouse.sql.gotmpl"] = "BOTH/schema.sql"
			templateFiles["sql/substreams.clickhouse.yaml.gotmpl"] = "BOTH/substreams.clickhouse.yaml"
		case "sql":
			templateFiles["sql/run-local.sh.gotmpl"] = "run-local.sh"
			templateFiles["sql/schema.sql.gotmpl"] = "BOTH/schema.sql"
			templateFiles["sql/substreams.sql.yaml.gotmpl"] = "substreams/substreams.sql.yaml"
		default:
			return nil, nil, fmt.Errorf("unknown sql output flavor %q", p.SqlOutputFlavor)
		}

	case outputTypeSubgraph:
		switch p.SubgraphOutputFlavor {
		case "entity":
			templateFiles["entities/Makefile.gotmpl"] = "substreams/Makefile"
			templateFiles["entities/substreams.yaml.gotmpl"] = "substreams/substreams.yaml"
			templateFiles["entities/schema.graphql.gotmpl"] = "schema.graphql"
			templateFiles["entities/subgraph.yaml.gotmpl"] = "subgraph.yaml"
			templateFiles["entities/README.md"] = "README.md"
			// TODO: is this really needed in the entity mode? As all the entity changes are coming out of the substreams, this may not be useful
			templateFiles["entities/package.json"] = "package.json"
			templateFiles["entities/dev-environment/config.toml.gotmpl"] = "dev-environment/config.toml"
			templateFiles["entities/dev-environment/docker-compose.yml"] = "dev-environment/docker-compose.yml"
			templateFiles["entities/dev-environment/start.sh"] = "dev-environment/start.sh"

		case "trigger":
			templateFiles["triggers/Makefile.gotmpl"] = "substreams/Makefile"
			templateFiles["triggers/substreams.yaml.gotmpl"] = "substreams/substreams.yaml"
			templateFiles["triggers/subgraph.yaml.gotmpl"] = "subgraph.yaml"
			templateFiles["triggers/schema.graphql.gotmpl"] = "schema.graphql"
			templateFiles["triggers/package.json.gotmpl"] = "package.json"
			templateFiles["triggers/src/mappings.ts.gotmpl"] = "src/mappings.ts"
			templateFiles["triggers/buf.gen.yaml"] = "buf.gen.yaml"
			templateFiles["triggers/run-local.sh.gotmpl"] = "run-local.sh"
			templateFiles["triggers/dev-environment/config.toml.gotmpl"] = "dev-environment/config.toml"
			templateFiles["triggers/dev-environment/docker-compose.yml"] = "dev-environment/docker-compose.yml"
			templateFiles["triggers/dev-environment/start.sh"] = "dev-environment/start.sh"
		default:
			return nil, nil, fmt.Errorf("unknown subgraph output flavor %q", p.SubgraphOutputFlavor)
		}
	default:
		return nil, nil, fmt.Errorf("invalid output type %q", p.outputType)
	}

	for templateFile, finalFileName := range templateFiles {
		zlog.Debug("reading ethereum project entry", zap.String("filename", templateFile))

		var content []byte
		if strings.HasSuffix(templateFile, ".gotmpl") {
			buffer := &bytes.Buffer{}
			if err := tpls.ExecuteTemplate(buffer, templateFile, p); err != nil {
				return nil, nil, fmt.Errorf("embed render entry template %q: %w", templateFile, err)
			}
			content = buffer.Bytes()
		} else {
			content, err = templatesFS.ReadFile("templates/" + templateFile)
			if err != nil {
				return nil, nil, fmt.Errorf("reading %q: %w", templateFile, err)
			}
		}

		if strings.HasPrefix(finalFileName, "substreams/") {
			substreamsFiles[finalFileName] = content
		} else if strings.HasPrefix(finalFileName, "BOTH/") {
			noPrefix, _ := strings.CutPrefix(finalFileName, "BOTH/")
			substreamsFiles["substreams/"+noPrefix] = content
			projectFiles[noPrefix] = content
		} else {
			projectFiles[finalFileName] = content
		}
	}

	for _, contract := range p.Contracts {
		substreamsFiles[fmt.Sprintf("substreams/abi/%s_contract.abi.json", contract.Name)] = []byte(contract.abi.raw)
	}

	for _, dds := range p.DynamicContracts {
		substreamsFiles[fmt.Sprintf("substreams/abi/%s_contract.abi.json", dds.Name)] = []byte(dds.abi.raw)
	}

	return
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
