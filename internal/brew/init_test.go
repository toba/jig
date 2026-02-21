package brew

import (
	"strings"
	"testing"
)

func TestGenerateReadme(t *testing.T) {
	readme := generateReadme("todo", "toba", "an issue tracker")

	checks := []string{
		"# Homebrew Tap for todo",
		"[todo](https://github.com/toba/todo)",
		"brew tap toba/todo",
		"brew install todo",
		"brew upgrade todo",
		"brew untap toba/todo",
		"toba/todo/issues",
	}
	for _, want := range checks {
		if !strings.Contains(readme, want) {
			t.Errorf("readme missing %q", want)
		}
	}
}
