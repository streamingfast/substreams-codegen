package stellartransactionsoperations

import (
	"regexp"
)

func isFilterCorrect(s string) bool {
	// Regular expression: Allows lowercase letters, numbers, and underscores, separated by commas
	re := regexp.MustCompile(`^[a-z0-9_]+(,[a-z0-9_]+)*$`)
	return re.MatchString(s)
}
