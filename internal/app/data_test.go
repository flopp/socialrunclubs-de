package app

import "testing"

func TestClubLastMod(t *testing.T) {
	tests := []struct {
		name           string
		updated, added string
		want           string
	}{
		{"updated", "2026-02-03", "2025-01-01", "2026-02-03"},
		{"only added", "", "2025-01-01", "2025-01-01"},
		{"invalid updated falls back", "2026-10--28", "2025-01-01", "2025-01-01"},
		{"year too small", "0206-05-14", "", ""},
		{"future", "2999-01-01", "", ""},
		{"garbage", "yesterday", "", ""},
	}
	for _, tt := range tests {
		c := &Club{UpdatedRaw: tt.updated, AddedRaw: tt.added}
		if got := c.LastMod(); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
