package solanchor

import "testing"

func TestProtobufToRustName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SwapRaydiumCPVariant", "SwapRaydiumCpVariant"},
		{"RaydiumCP", "RaydiumCP"},
		{"WrappedI80F48", "WrappedI80f48"},
		{"MyHTTPResponse", "MyHttpResponse"},
		{"URLParser", "UrlParser"},
		{"ABCTestXYZ", "AbcTestXYZ"},
		{"CPUUsageData", "CpuUsageData"},
		{"IOConfig", "IoConfig"},
		{"HTTP2Connection", "Http2Connection"},
		{"DNSInfoPacket", "DnsInfoPacket"},
	}

	for _, tt := range tests {
		result := ToRustPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("ProtobufToRustName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}