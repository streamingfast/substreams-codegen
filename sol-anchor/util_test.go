package solanchor

import "testing"

func TestToRustFriendlyPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"WrappedI80F48", "WrappedI80f48"},
		{"MyXMLParser", "Myxmlparser"},
		{"ETHPriceFeed", "EthPriceFeed"},
		{"I80F48Wrapper", "I80f48Wrapper"},
		{"SimpleName", "SimpleName"},
		{"bigUIDHandler", "BigUidHandler"},
		{"multiPARTnameTEST", "Multipartnametest"},
		{"", ""},
		{"some_number_123", "SomeNumber123"},
	}

	for _, tt := range tests {
		result := ToRustFriendlyPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToRustFriendlyPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}