package zed

import (
	"testing"
)

func TestParseLanguages(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"CSS", []string{"CSS"}},
		{"Go Text Template,Go HTML Template", []string{"Go Text Template", "Go HTML Template"}},
		{" CSS , HTML ", []string{"CSS", "HTML"}},
		{"", nil},
		{",,,", nil},
	}
	for _, tt := range tests {
		got := parseLanguages(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("parseLanguages(%q) = %v, want %v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseLanguages(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
}
