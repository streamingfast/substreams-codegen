package starknet_events

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"substreams.yaml.gotmpl": "substreams.yaml",
		"README.md.gotmpl":       "README.md",
	})
}
