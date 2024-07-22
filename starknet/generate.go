package starknet

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"strings"
	"time"

	pbbuild "github.com/streamingfast/substreams-codegen/pb/sf/codegen/remotebuild/v1"

	"github.com/streamingfast/dgrpc"
	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
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
		"txfilter/lib.rs.gotmpl":     "substreams/src/lib.rs",
		"Cargo.toml.gotmpl":          "substreams/Cargo.toml",
		"rust-toolchain.toml":        "substreams/rust-toolchain.toml",
		".gitignore":                 "substreams/.gitignore",
		"Makefile.gotmpl":            "substreams/Makefile",
		"txfilter/schema.sql.gotmpl": "substreams/schema.sql",
	}

	switch outType {
	case outputTypeSQL:
		switch p.SqlOutputFlavor {
		case sqlTypeSQL:
			templateFiles["txfilter/substreams.sql.yaml.gotmpl"] = "substreams/substreams.yaml"
		case sqlTypeClickhouse:
			templateFiles["txfilter/substreams.clickhouse.yaml.gotmpl"] = "substreams/substreams.yaml"
		}
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
