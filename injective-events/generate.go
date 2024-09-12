package injective_events

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() (projectFiles map[string][]byte, err error) {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		".gitignore":             ".gitignore",
		"README.md.gotmpl":       "README.md",
		"substreams.yaml.gotmpl": "substreams.yaml",
	})
}

func cmdGenerate(p *Project) loop.Cmd {
	return func() loop.Msg {
		projectFiles, err := p.Generate(outType)
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			ProjectFiles: projectFiles,
		}
	}
}
