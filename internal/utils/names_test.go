package utils

import (
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "berlin", "berlin"},
		{"uppercase conversion", "BERLIN", "berlin"},
		{"mixed case", "BeRlIn", "berlin"},
		{"spaces replaced with dash", "New York", "new-york"},
		{"german umlauts", "München", "muenchen"},
		{"german characters", "Äpfel Öl Über Straße", "aepfel-oel-ueber-strasse"},
		{"special characters replaced", "St. Peter's (Main)", "st-peter-s-main"},
		{"dots replaced", "Köln.", "koeln"},
		{"quotes removed", `"Quoted" Text'Name`, "quoted-text-name"},
		{"parentheses replaced", "Club (Berlin)", "club-berlin"},
		{"consecutive dashes collapsed", "Two  Spaces", "two-spaces"},
		{"leading dashes removed", "-Started", "started"},
		{"trailing dashes removed", "Ended-", "ended"},
		{"numbers preserved", "Club 123", "club-123"},
		{"norwegian characters", "Tromsø", "tromso"},
		{"empty string", "", ""},
		{"only special characters", "---", ""},
		{"complex mix", "St. Peter's Club (Köln/Cologne)", "st-peter-s-club-koeln-cologne"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{"simple split", "one,two,three", ",", []string{"one", "two", "three"}},
		{"split with spaces", "one , two , three", ",", []string{"one", "two", "three"}},
		{"semicolon separator", "first; second; third", ";", []string{"first", "second", "third"}},
		{"empty parts removed", "one,,three", ",", []string{"one", "three"}},
		{"space parts removed", "one,  ,three", ",", []string{"one", "three"}},
		{"single element", "single", ",", []string{"single"}},
		{"trailing separator", "one,two,", ",", []string{"one", "two"}},
		{"leading separator", ",one,two", ",", []string{"one", "two"}},
		{"empty string", "", ",", []string{}},
		{"only spaces", "   ", ",", []string{}},
		{"multi-char separator", "one OR two OR three", " OR ", []string{"one", "two", "three"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitAndTrim(tt.input, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("SplitAndTrim(%q, %q) returned %d elements, want %d",
					tt.input, tt.sep, len(result), len(tt.expected))
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("SplitAndTrim(%q, %q)[%d] = %q, want %q",
						tt.input, tt.sep, i, v, tt.expected[i])
				}
			}
		})
	}
}
