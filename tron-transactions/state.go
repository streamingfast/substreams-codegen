package trontransactions

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/base"
)

//go:embed templates/*
var templatesFS embed.FS

type Project struct {
	base.ConversationState
	FilterType string `json:"filterType,omitempty"`
	Filter     string `json:"filter,omitempty"`
}

func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"src/lib.rs.gotmpl":             "src/lib.rs",
		"Cargo.toml.gotmpl":             "Cargo.toml",
		".gitignore.gotmpl":             ".gitignore",
		"substreams.yaml.gotmpl":        "substreams.yaml",
		"README.md.gotmpl":              "README.md",
		"common-templates/buf.gen.yaml": "buf.gen.yaml",
	})
}
