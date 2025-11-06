package soltransactions

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

type Project struct {
	codegen.BaseConversationState
	InitialBlock          uint64 `json:"initialBlock,omitempty"`
	InitialBlockSet       bool   `json:"initialBlockSet,omitempty"`
	Filter                string `json:"filter,omitempty"`
	FilterContainsAccount bool   `json:"filterContainsAccount,omitempty"`
}

// use the output type form the Project to render the templates
func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"substreams.yaml.gotmpl":        "substreams.yaml",
		"README.md.gotmpl":              "README.md",
		".gitignore.gotmpl":             ".gitignore",
		"common-templates/buf.gen.yaml": "buf.gen.yaml",
		"Cargo.toml.gotmpl":             "Cargo.toml",
		"src/lib.rs.gotmpl":             "src/lib.rs",
		"proto/mydata.proto.gotmpl":     "proto/mydata.proto",
	})
}
