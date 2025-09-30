package stellartransactionsoperations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidCommaSeparatedString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello,world,test", true},       // Valid
		{"123,456,789", true},            // Valid
		{"abc,123,xyz_edsds", true},      // Valid
		{"hello, world", false},          // Invalid (space present)
		{"hello ,world", false},          // Invalid (space present)
		{"hello,,world", false},          // Invalid (double comma)
		{"hello,world,", false},          // Invalid (trailing comma)
		{",hello,world", false},          // Invalid (leading comma)
		{"singleword", true},             // Valid (single word)
		{"", false},                      // Invalid (empty string)
		{"payment,create_account", true}, // Valid
		{"GADLRTGF4GCU2CNAHYPKAEBGQSBX2M3UYIZZJODZAVC5A5QCAE7AT66C", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			require.Equal(t, test.expected, isFilterCorrect(test.input), "input: %s", test.input)
		})
	}
}
