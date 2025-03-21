package stellartransactionsoperations

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

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
