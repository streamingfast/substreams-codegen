package stellartransactionsoperations

import (
	"embed"
	"fmt"
	"strings"

	codegen "github.com/streamingfast/substreams-codegen"
	"github.com/streamingfast/substreams-codegen/base"
)

//go:embed templates/*
var templatesFS embed.FS

type Project struct {
	base.ConversationState
	Compile    bool   `json:"compile,omitempty"` // optional field to write in state and automatically compile with no confirmation.
	Download   bool   `json:"download,omitempty"`
	FilterType string `json:"filterType,omitempty"`
	Filter     string `json:"filter,omitempty"`
}

// func (p *Project) ModuleName() string { return strings.ReplaceAll(p.Name, "-", "_") }
func (p *Project) KebabName() string { return strings.ReplaceAll(p.Name, "_", "-") }

func (p *Project) ComposeFilter() string {
	keyword := "source_account"
	if p.FilterType == "operations" {
		keyword = "operation"
	}

	var stringBuilder string

	splittedFilter := strings.Split(p.Filter, ",")
	for i, filterPiece := range splittedFilter {
		stringBuilder += fmt.Sprintf("%s:%s", keyword, filterPiece)

		if i < (len(splittedFilter) - 1) {
			stringBuilder += " || "
		}
	}

	return stringBuilder
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
