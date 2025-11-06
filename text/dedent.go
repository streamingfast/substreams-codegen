package text

import (
	"fmt"
	"strings"

	"github.com/lithammer/dedent"
)

func Dedent(input string, args ...any) string {
	return strings.TrimSpace(dedent.Dedent(fmt.Sprintf(input, args...)))

}
