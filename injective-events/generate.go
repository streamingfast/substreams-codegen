package injective_events

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/streamingfast/dgrpc"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
	"go.uber.org/zap"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Render(outType outputType) (substreamsFiles map[string][]byte, projectFiles map[string][]byte, err error) {
	substreamsFiles = map[string][]byte{}
	projectFiles = map[string][]byte{}

	tpls, err := codegen.ParseFS(nil, templatesFS, "**/*.gotmpl")
	if err != nil {
		return nil, nil, fmt.Errorf("parse templates: %w", err)
	}

	templateFiles := map[string]string{
		".gitignore": ".gitignore",
	}

	switch outType {
	case outputTypeSQL:
		templateFiles["sql/README.md"] = "README.md"
		templateFiles["sql/dev-environment/docker-compose.yml.gotmpl"] = "dev-environment/docker-compose.yml"
		templateFiles["sql/dev-environment/start.sh"] = "dev-environment/start.sh"
		templateFiles["lib.rs.gotmpl"] = "substreams/src/lib.rs"
		templateFiles["proto/events.proto"] = "substreams/proto/events.proto"
		templateFiles["sql/buf.gen.yaml"] = "substreams/buf.gen.yaml"
		templateFiles["sql/buf.yaml"] = "substreams/buf.yaml"
		templateFiles["rust-toolchain.toml"] = "substreams/rust-toolchain.toml"
		templateFiles["sql/Makefile.gotmpl"] = "substreams/Makefile"
		templateFiles["sql/run-local.sh.gotmpl"] = "run-local.sh"
		templateFiles["Cargo.toml.gotmpl"] = "substreams/Cargo.toml"

		switch p.SqlOutputFlavor {
		case "clickhouse":
			templateFiles["sql/schema.clickhouse.sql.gotmpl"] = "BOTH/schema.sql"
			templateFiles["sql/substreams.clickhouse.yaml.gotmpl"] = "substreams/substreams.clickhouse.yaml"
		case "sql":
			templateFiles["sql/schema.sql.gotmpl"] = "BOTH/schema.sql"
			templateFiles["sql/substreams.sql.yaml.gotmpl"] = "substreams/substreams.sql.yaml"
		default:
			return nil, nil, fmt.Errorf("unknown sql output flavor %q", p.SqlOutputFlavor)
		}

	case outputTypeSubgraph:
		templateFiles["triggers/README.md"] = "README.md"
		switch p.SubgraphOutputFlavor {
		case "trigger":
			templateFiles["triggers/dev-environment/docker-compose.yml.gotmpl"] = "dev-environment/docker-compose.yml"
			templateFiles["triggers/dev-environment/config.toml.gotmpl"] = "dev-environment/config.toml"
			templateFiles["triggers/dev-environment/start.sh"] = "dev-environment/start.sh"
			templateFiles["triggers/Makefile"] = "substreams/Makefile"
			templateFiles["triggers/substreams.yaml.gotmpl"] = "substreams/substreams.yaml"
			templateFiles["triggers/buf.gen.yaml"] = "buf.gen.yaml"
			templateFiles["triggers/package.json.gotmpl"] = "package.json"
			templateFiles["triggers/tsconfig.json"] = "tsconfig.json"
			templateFiles["triggers/subgraph.yaml.gotmpl"] = "subgraph.yaml"
			templateFiles["triggers/schema.graphql"] = "schema.graphql"
			templateFiles["triggers/src/mappings.ts"] = "src/mappings.ts"
			templateFiles["triggers/run-local.sh"] = "run-local.sh"
		default:
			return nil, nil, fmt.Errorf("unknown subgraph output flavor %q", p.SubgraphOutputFlavor)
		}

	default:
		return nil, nil, fmt.Errorf("invalid output type %q", outType)
	}

	for templateFile, finalFilename := range templateFiles {
		zlog.Debug("reading injective project entry", zap.String("filename", templateFile), zap.String("finalFilename", finalFilename))

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

		if strings.HasPrefix(finalFilename, "substreams/") {
			substreamsFiles[finalFilename] = content
		} else if strings.HasPrefix(finalFilename, "BOTH/") {
			noPrefix, _ := strings.CutPrefix(finalFilename, "BOTH/")
			substreamsFiles["substreams/"+noPrefix] = content
			projectFiles[noPrefix] = content
		} else {
			projectFiles[finalFilename] = content
		}
	}

	return
}

func (p *Project) generate(outType outputType) ([]byte, []byte, error) {
	srcFiles, projectFiles, err := p.Render(outType)
	if err != nil {
		return nil, nil, fmt.Errorf("rendering template: %w", err)
	}

	substreamsZip, err := codegen.ZipFiles(srcFiles)
	if err != nil {
		return nil, nil, fmt.Errorf("zipping: %w", err)
	}

	projectZip, err := codegen.ZipFiles(projectFiles)
	if err != nil {
		return nil, nil, fmt.Errorf("zipping: %w", err)
	}

	return projectZip, substreamsZip, nil
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

	for {
		resp, err := res.Recv()
		if err != nil {
			remoteBuildContentChan <- &codegen.RemoteBuildState{
				Error: err.Error(),
			}
			return
		}

		if resp == nil {
			break
		}

		if resp.Error != "" {
			remoteBuildContentChan <- &codegen.RemoteBuildState{
				Error: resp.Error,
			}
			return
		}

		if len(resp.Artifacts) != 0 {
			remoteBuildContentChan <- &codegen.RemoteBuildState{
				Error:     resp.Error,
				Logs:      []string{resp.Logs},
				Artifacts: resp.Artifacts,
			}
			return
		}

		// send the request as we go
		remoteBuildContentChan <- &codegen.RemoteBuildState{
			Logs: []string{resp.Logs},
		}
	}
}

func cmdGenerate(p *Project, outType outputType) loop.Cmd {
	p.buildStarted = time.Now()

	return func() loop.Msg {
		projectZip, substreamsZip, err := p.generate(outType)
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			ProjectZip:          projectZip,
			SubstreamsSourceZip: substreamsZip,
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

func cmdBuildFailed(err error) loop.Cmd {
	return func() loop.Msg {
		return codegen.ReturnBuild{Err: err}
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
