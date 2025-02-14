package injective_events

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		".gitignore.gotmpl":             ".gitignore",
		"README.md.gotmpl":              "README.md",
		"substreams.yaml.gotmpl":        "substreams.yaml",
		"common-templates/buf.gen.yaml": "buf.gen.yaml",
		"Cargo.toml.gotmpl":             "Cargo.toml",
		"src/lib.rs.gotmpl":             "src/lib.rs",
		"proto/mydata.proto.gotmpl":     "proto/mydata.proto",
	})
}
