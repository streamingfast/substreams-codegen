package ethfull

import (
	"regexp"
	"strings"
)

func sanitizeABIStructName(rustABIStructName string) string {
	if strings.HasPrefix(rustABIStructName, "_") {
		rustABIStructName = "u_" + rustABIStructName[1:]
	}

	reg := regexp.MustCompile("_+")

	result := reg.ReplaceAllStringFunc(rustABIStructName, func(s string) string {
		count := len(s)

		replacement := strings.Repeat("_u", count-1) + "_"

		return replacement
	})

	return result
}
