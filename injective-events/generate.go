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
		".gitignore":             ".gitignore",
		"README.md":              "README.md",
		"substreams.yaml.gotmpl": "substreams.yaml",
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
		projectFiles[finalFilename] = content
	}

	return
}

func (p *Project) generate(outType outputType) (map[string][]byte, map[string][]byte, error) {
	srcFiles, projectFiles, err := p.Render(outType)
	if err != nil {
		return nil, nil, fmt.Errorf("rendering template: %w", err)
	}

	return projectFiles, srcFiles, nil
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
		projectFiles, sourceFiles, err := p.generate(outType)
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			ProjectFiles: projectFiles,
			SourceFiles:  sourceFiles,
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
