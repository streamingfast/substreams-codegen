package starknet_events

import (
	"embed"
	"fmt"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

func (p *Project) Generate() codegen.ReturnGenerate {
	res := codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/events.proto.gotmpl": "proto/events.proto",
		"src/abi/mod.rs.gotmpl":     "src/abi/mod.rs",
		"src/lib.rs.gotmpl":         "src/lib.rs",
		"build.rs.gotmpl":           "build.rs",
		"Cargo.toml.gotmpl":         "Cargo.toml",
		"rust-toolchain.toml":       "rust-toolchain.toml",
		".gitignore":                ".gitignore",
		"substreams.yaml.gotmpl":    "substreams.yaml",
		"README.md.gotmpl":          "README.md",
	})
	if res.Err != nil {
		return res
	}

	for _, contract := range p.Contracts {
		res.ProjectFiles[fmt.Sprintf("abi/%s_contract.abi.json", contract.Name)] = []byte(contract.Abi.raw)
	}

	return res
}
