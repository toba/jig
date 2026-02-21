package github

import (
	"testing"
	"time"
)

func TestCommitNormalize(t *testing.T) {
	c := Commit{
		SHA: "abc123",
		RawCommit: rawCommit{
			Message: "Fix bug\n\nDetailed description",
			RawAuthor: rawDate{
				Name: "Test User",
				Date: time.Date(2026, 2, 18, 22, 8, 27, 0, time.UTC),
			},
		},
		RawAuthor: &rawAuthor{Login: "testuser"},
	}
	c.Normalize()

	if c.Message != "Fix bug" {
		t.Errorf("message = %q, want %q", c.Message, "Fix bug")
	}
	if c.Author != "testuser" {
		t.Errorf("author = %q, want %q", c.Author, "testuser")
	}
	if !c.Date.Equal(time.Date(2026, 2, 18, 22, 8, 27, 0, time.UTC)) {
		t.Errorf("date = %v, want 2026-02-18T22:08:27Z", c.Date)
	}
}

func TestCommitNormalizeFallbackAuthor(t *testing.T) {
	c := Commit{
		RawCommit: rawCommit{
			Message: "Single line",
			RawAuthor: rawDate{Name: "Fallback Name"},
		},
	}
	c.Normalize()

	if c.Author != "Fallback Name" {
		t.Errorf("author = %q, want %q", c.Author, "Fallback Name")
	}
	if c.Message != "Single line" {
		t.Errorf("message = %q, want %q", c.Message, "Single line")
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"hello\nworld", "hello"},
		{"", ""},
		{"first\nsecond\nthird", "first"},
	}
	for _, tt := range tests {
		if got := firstLine(tt.in); got != tt.want {
			t.Errorf("firstLine(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
