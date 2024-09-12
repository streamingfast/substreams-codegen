package ethminimal

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/loop"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() (projectFiles map[string][]byte, err error) {
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
		projFiles, err := p.Generate()
		if err != nil {
			return codegen.ReturnGenerate{Err: err}
		}
		return codegen.ReturnGenerate{
			ProjectFiles: projFiles,
		}
	}
}
