package nope

import (
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	if code := RunHelp(); code != 0 {
		t.Fatalf("RunHelp() = %d, want 0", code)
	}
}

func TestHelpTextContainsKeySections(t *testing.T) {
	sections := []string{
		"USAGE",
		"CONFIGURATION",
		"RULE FIELDS",
		"BUILTINS",
		"TOOL SCOPING",
		"EXIT CODES",
		"EXAMPLES",
	}
	for _, s := range sections {
		if !strings.Contains(HelpText, s) {
			t.Errorf("HelpText missing section %q", s)
		}
	}
}

func TestHelpTextDocumentsAllBuiltins(t *testing.T) {
	builtins := []string{
		"multiline",
		"pipe",
		"chained",
		"redirect",
		"subshell",
		"credential-read",
		"network",
	}
	for _, b := range builtins {
		if !strings.Contains(HelpText, b) {
			t.Errorf("HelpText missing builtin %q", b)
		}
	}
}
