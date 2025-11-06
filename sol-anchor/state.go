package solanchor

import (
	"embed"

	codegen "github.com/streamingfast/substreams-codegen"
)

//go:embed templates/*
var templatesFS embed.FS

type Project struct {
	codegen.BaseConversationState
	InitialBlock    uint64 `json:"initialBlock,omitempty"`
	InitialBlockSet bool   `json:"initialBlockSet,omitempty"`
	IdlFormat       string `json:"idlFormat,omitempty"`
	IdlString       string `json:"idlString,omitempty"`

	idl *IDL
}

// use the output type form the Project to render the templates
func (p *Project) Generate() codegen.ReturnGenerate {
	return codegen.GenerateTemplateTree(p, templatesFS, map[string]string{
		"proto/program.proto.gotmpl": "proto/program.proto",
		"idls/program.json.gotmpl":   "idls/program.json",
		"src/lib.rs.gotmpl":          "src/lib.rs",
		"src/idl/mod.rs.gotmpl":      "src/idl/mod.rs",
		".gitignore.gotmpl":          ".gitignore",
		"buf.gen.yaml.gotmpl":        "buf.gen.yaml",
		"Cargo.lock.gotmpl":          "Cargo.lock",
		"Cargo.toml.gotmpl":          "Cargo.toml",
		"substreams.yaml.gotmpl":     "substreams.yaml",
	})
}

func (p *Project) Idl() *IDL { return p.idl }
