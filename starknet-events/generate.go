package starknet_events

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

//go:embed templates/*
var templatesFS embed.FS

func cmdGenerate(p *Project) loop.Cmd {
	return func() loop.Msg {
		projFiles, err := p.Generate()
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			ProjectFiles: projFiles,
		}
	}
}

// use the output type form the Project to render the templates
func (p *Project) Generate() (projectFiles map[string][]byte, err error) {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"substreams.yaml.gotmpl": "substreams.yaml",
		"README.md.gotmpl":       "README.md",
	})
}
