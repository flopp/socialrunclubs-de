package utils

import "testing"

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "https URL", value: "https://example.com/path?query=value", want: true},
		{name: "http URL", value: "http://localhost:8080/", want: true},
		{name: "uppercase scheme", value: "HTTPS://example.com", want: true},
		{name: "empty value", value: "", want: false},
		{name: "missing scheme", value: "example.com/path", want: false},
		{name: "missing host", value: "https:///path", want: false},
		{name: "unsupported scheme", value: "ftp://example.com/file", want: false},
		{name: "malformed URL", value: "https://exa mple.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidURL(tt.value); got != tt.want {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}