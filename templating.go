package codegen

import (
	"bytes"
	"embed"
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

type GenerateConfig struct {
	FS    embed.FS
	Files map[string]string
}

func GenerateTemplateTree(projectData any, templatesFS embed.FS, templateFiles map[string]string) ReturnGenerate {
	projFiles, err := generateTemplateTree(projectData, templatesFS, templateFiles)
	if err != nil {
		return ReturnGenerate{Err: err}
	}
	return ReturnGenerate{ProjectFiles: projFiles}
}

func generateTemplateTree(projectData any, templatesFS embed.FS, templateFiles map[string]string) (map[string][]byte, error) {
	projectFiles := map[string][]byte{}

	tpls, err := ParseFS(templatesFS, "**/*.gotmpl") // TODO: close when refactored
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	for templateFile, finalFileName := range templateFiles {
		//zlog.Debug("reading template file", zap.String("filename", templateFile))

		var content []byte
		if strings.HasSuffix(templateFile, ".gotmpl") {
			buffer := &bytes.Buffer{}
			if err := tpls.ExecuteTemplate(buffer, templateFile, projectData); err != nil {
				return nil, fmt.Errorf("embed render entry template %q: %w", templateFile, err)
			}
			content = buffer.Bytes()
		} else {
			content, err = templatesFS.ReadFile("templates/" + templateFile)
			if err != nil {
				return nil, fmt.Errorf("reading %q: %w", templateFile, err)
			}
		}

		projectFiles[finalFileName] = content
	}

	return projectFiles, nil
}

// ParseFS reads the files from the embedded FS and parses them into named templates.
func ParseFS(fsys fs.FS, pattern string) (*template.Template, error) {
	t, err := commonTemplates.Clone()
	if err != nil {
		return nil, err
	}

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
