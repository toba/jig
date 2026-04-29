package cc

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cc.yaml")
	c := &Config{
		Version:      1,
		SharedSource: "/x/.claude",
		Private:      DefaultPrivate,
		Aliases: map[string]Alias{
			"main": {CLI: "claude", Path: "/x/.claude", IsSource: true},
			"work": {CLI: "claude", Path: "/x/.jig/cc/work"},
		},
	}
	if err := c.SaveTo(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.SharedSource != c.SharedSource {
		t.Errorf("shared_source: got %q want %q", got.SharedSource, c.SharedSource)
	}
	if !got.Aliases["main"].IsSource {
		t.Error("main alias lost is_source")
	}
	if got.Aliases["work"].IsSource {
		t.Error("work alias gained is_source")
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		c    Config
		err  string
	}{
		{
			"version mismatch",
			Config{Version: 2, SharedSource: "/x", Aliases: map[string]Alias{"m": {CLI: "claude", Path: "/x", IsSource: true}}},
			"version",
		},
		{
			"no source",
			Config{Version: 1, SharedSource: "/x", Aliases: map[string]Alias{"m": {CLI: "claude", Path: "/x"}}},
			"is_source",
		},
		{
			"two sources",
			Config{Version: 1, SharedSource: "/x", Aliases: map[string]Alias{
				"a": {CLI: "claude", Path: "/x", IsSource: true},
				"b": {CLI: "claude", Path: "/y", IsSource: true},
			}},
			"is_source",
		},
		{
			"missing cli",
			Config{Version: 1, SharedSource: "/x", Aliases: map[string]Alias{"m": {Path: "/x", IsSource: true}}},
			"cli",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.c.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Errorf("Validate() = %v, want substring %q", err, tc.err)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	c := &Config{
		Aliases: map[string]Alias{
			"main": {CLI: "claude", Path: "/m"},
			"work": {CLI: "claude", Path: "/w"},
			"wow":  {CLI: "claude", Path: "/wow"},
		},
	}

	// Exact match.
	n, _, err := c.Resolve("main")
	if err != nil || n != "main" {
		t.Errorf("exact: got (%q, %v)", n, err)
	}

	// Unique prefix.
	n, _, err = c.Resolve("ma")
	if err != nil || n != "main" {
		t.Errorf("unique-prefix: got (%q, %v)", n, err)
	}

	// Ambiguous.
	if _, _, err := c.Resolve("w"); err == nil {
		t.Error("ambiguous prefix should error")
	}

	// No match.
	if _, _, err := c.Resolve("zz"); err == nil {
		t.Error("nonexistent should error")
	}
}
