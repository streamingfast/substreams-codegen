package varaextrinsics

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

// use the output type form the Project to render the templates
func (p *Project) CmdGenerate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"substreams.yaml.gotmpl": "substreams.yaml",
		"README.md.gotmpl":       "README.md",
	})
}
