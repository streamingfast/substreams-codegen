package codegen

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"
)

//go:embed common-templates/*.gotmpl
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
