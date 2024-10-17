package solminimal

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/mydata.proto.gotmpl":     "proto/mydata.proto",
		"src/pb/mod.rs.gotmpl":          "src/pb/mod.rs",
		"src/lib.rs.gotmpl":             "src/lib.rs",
		"Cargo.toml.gotmpl":             "Cargo.toml",
		".gitignore.gotmpl":             ".gitignore",
		"substreams.yaml.gotmpl":        "substreams.yaml",
		"README.md.gotmpl":              "README.md",
		"common-templates/buf.gen.yaml": "buf.gen.yaml",
	})
}
