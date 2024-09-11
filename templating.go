package codegen

import (
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
	"github.com/iancoleman/strcase"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/golang-cz/textcase"
)

var templateFuncs = template.FuncMap{
	"add": func(left int, right int) int {
		return left + right
	},
	"toUpper":                strings.ToUpper,
	"toKebabCase":            textcase.KebabCase,
	"toSnakeCase":            xstrings.ToSnakeCase,
	"toLowerCamelCase":       strcase.ToLowerCamel,
	"toPascalCase":           textcase.PascalCase,
	"sanitizeProtoFieldName": SanitizeProtoFieldName,
}

// ParseFS reads the files from the embedded FS and parses them into named templates.
func ParseFS(myFuncs template.FuncMap, fsys fs.FS, pattern string) (*template.Template, error) {
	t, err := commonTemplates.Clone()
	if err != nil {
		return nil, err
	}
	t.Funcs(myFuncs)

	filenames, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return nil, err
	}
	if len(filenames) == 0 {
		return nil, fmt.Errorf("template: pattern matches no files: %#q", pattern)
	}

	for _, filename := range filenames {
		b, err := fs.ReadFile(fsys, filename)
		if err != nil {
			return nil, err
		}

		name, _ := strings.CutPrefix(filename, "templates/")

		_, err = t.New(name).Parse(string(b))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func SanitizeProtoFieldName(name string) string {
	regex := regexp.MustCompile("(\\d+)(_*)")
	name = regex.ReplaceAllStringFunc(name, func(match string) string {
		if strings.HasSuffix(match, "_") {
			return match
		}
		return match + "_"
	})

	if strings.HasPrefix(name, "_") {
		name = "u" + name
	}

	if strings.HasSuffix(name, "_") {
		name = strings.TrimSuffix(name, "_")
	}
	return name
}
