package stellartransactionsoperations

import (
	"regexp"
)

// Regular expression: Allows letters, numbers, and underscores, separated by commas
var filterRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+(,[a-zA-Z0-9_]+)*$`)

func isFilterCorrect(s string) bool {
	return filterRegexp.MatchString(s)
}
