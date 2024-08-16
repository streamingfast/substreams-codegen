package starknetminimal

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/streamingfast/dgrpc"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"
	"go.uber.org/zap"
)

//go:embed templates/*
var templatesFS embed.FS

func cmdGenerate(p *Project) loop.Cmd {
	return func() loop.Msg {
		projFiles, err := p.generate()
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			ProjectFiles: projFiles,
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

func (p *Project) generate() (projFiles map[string][]byte, err error) {
	// TODO: before doing any generation, we'll want to validate
	// all data points that are going into source code.
	// We don't want some weird things getting into `build.rs`
	// and being executed server side, so we'll need pristine validation
	// of all inputs here.
	// TODO: add some checking to make sure `ParentContractName` of DynamicContract
	// do match a Contract that exists here.

	projFiles, err = p.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering template: %w", err)
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

	projectZip, err := codegen.ZipFiles(p.projectFiles)
	if err != nil {
		remoteBuildContentChan <- &codegen.RemoteBuildState{
			Error: err.Error(),
		}
	}

	client := pbbuild.NewBuildServiceClient(conn)
	res, err := client.Build(context.Background(),
		&pbbuild.BuildRequest{
			SourceCode:     projectZip,
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
func (p *Project) Render() (projectFiles map[string][]byte, err error) {
	projectFiles = map[string][]byte{}

	tpls, err := codegen.ParseFS(nil, templatesFS, "**/*.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	templateFiles := map[string]string{
		"proto/mydata.proto.gotmpl": "proto/mydata.proto",
		"src/pb/mod.rs.gotmpl":      "src/pb/mod.rs",
		"src/lib.rs.gotmpl":         "src/lib.rs",
		"Cargo.toml.gotmpl":         "Cargo.toml",
		".gitignore":                ".gitignore",
		"substreams.yaml.gotmpl":    "substreams.yaml",
		"README.md.gotmpl":          "README.md",
		"CONTRIBUTING.md":           "CONTRIBUTING.md",
	}

	for templateFile, finalFileName := range templateFiles {
		zlog.Debug("reading ethereum project entry", zap.String("filename", templateFile))

		var content []byte
		if strings.HasSuffix(templateFile, ".gotmpl") {
			buffer := &bytes.Buffer{}
			if err := tpls.ExecuteTemplate(buffer, templateFile, p); err != nil {
				return nil, fmt.Errorf("embed render entry template %q: %w", templateFile, err)
			}
			content = buffer.Bytes()
		} else {
			content, err = templatesFS.ReadFile("templates/" + templateFile)
			if err != nil {
				return nil, fmt.Errorf("reading %q: %w", templateFile, err)
			}
		}

		projectFiles[finalFileName] = content
	}

	return
}
