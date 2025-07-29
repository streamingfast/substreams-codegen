package codegen

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/streamingfast/firehose-networks"
)

//go:embed common-templates/*
var commonTemplatesFS embed.FS

var commonTemplates *template.Template

func init() {
	var err error
	commonTemplates, err = parseCommonTemplates()
	if err != nil {
		panic(err)
	}
}

func MarkdownEscape(s string) string {
	return "```\n" + s + "\n```\n"
}

// NetworkToEndpoint maps network names to their corresponding StreamingFast endpoints
// using the firehose-networks library
func NetworkToEndpoint(network string) string {
	endpoint := networks.GetSubstreamsEndpoint(network)
	if endpoint == "" {
		// Fallback to a generic pattern if not found in registry
		return network + ".streamingfast.io:443"
	}
	return endpoint
}

func parseCommonTemplates() (*template.Template, error) {
	t := template.New("").Funcs(templateFuncs)
	filenames, err := doublestar.Glob(commonTemplatesFS, "**/*.gotmpl")
	if err != nil {
		return nil, err
	}

	for _, filename := range filenames {
		b, err := fs.ReadFile(commonTemplatesFS, filename)
		if err != nil {
			return nil, err
		}
		_, err = t.New(filename).Parse(string(b))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
