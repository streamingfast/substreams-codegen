package trontransactions

import "testing"

func TestIsValidCommaSeparatedString(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello,world,test", true},       // Valid
		{"123,456,789", true},            // Valid
		{"abc,123,xyz_edsds", true},            // Valid
		{"hello, world", false},          // Invalid (space present)
		{"hello ,world", false},          // Invalid (space present)
		{"hello,,world", false},          // Invalid (double comma)
		{"hello,world,", false},          // Invalid (trailing comma)
		{",hello,world", false},          // Invalid (leading comma)
		{"singleword", true},             // Valid (single word)
		{"", false},                      // Invalid (empty string)
		{"payment,create_account", true}, // Valid
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := isFilterCorrect(test.input)
			if result != test.expected {
				t.Errorf("For input %q, expected %v but got %v", test.input, test.expected, result)
			}
		})
	}
}
