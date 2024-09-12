package starknetminimal

import (
	"embed"
	"fmt"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

//go:embed templates/*
var templatesFS embed.FS

// use the output type form the Project to render the templates
func (p *Project) Render() (projectFiles map[string][]byte, err error) {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/mydata.proto.gotmpl": "proto/mydata.proto",
		"src/pb/mod.rs.gotmpl":      "src/pb/mod.rs",
		"src/lib.rs.gotmpl":         "src/lib.rs",
		"Cargo.toml.gotmpl":         "Cargo.toml",
		".gitignore":                ".gitignore",
		"substreams.yaml.gotmpl":    "substreams.yaml",
		"README.md.gotmpl":          "README.md",
	})
}

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
