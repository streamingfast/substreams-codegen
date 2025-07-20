package solanchor

import "testing"

func TestNormalizePascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SwapRaydiumCPVariant", "SwapRaydiumCpVariant"},
		{"SwapGooseFXV2Variant", "SwapGooseFxv2Variant"},
		{"SwapGooseFXVariant", "SwapGooseFxVariant"},
		{"XMLHttpRequest", "XmlHttpRequest"},
		{"HTTPRequestV2", "HttpRequestV2"},
		{"Test", "Test"},
		{"A", "A"},
		{"CPU", "Cpu"},
		{"SwapEvent", "SwapEvent"},
		{"GooseFX", "GooseFX"},
		{"GooseFXV2", "GooseFXV2"},
		{"GooseFXV2", "GooseFxv2"},

	}

	for _, tt := range tests {
		got := ToRustPascalCase(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizePascalCase(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
