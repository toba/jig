package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testYAML = `# Some other tool's config
other_tool:
  setting: value

upstream:
  sources:
    - repo: owner/name
      branch: main
      relationship: derived
      notes: "Derived into Sources/Tools/"
      last_checked_sha: abc123
      last_checked_date: "2026-02-18T22:08:27Z"
      paths:
        high:
          - "Sources/**/*.swift"
        medium:
          - "Package.swift"
          - "Tests/**"
        low:
          - ".github/**"
          - "README.md"

# Another section
another:
  key: val
`

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".jig.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	_, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}

	src := cfg.Sources[0]
	if src.Repo != "owner/name" {
		t.Errorf("repo = %q, want owner/name", src.Repo)
	}
	if src.Branch != "main" {
		t.Errorf("branch = %q, want main", src.Branch)
	}
	if src.Relationship != "derived" {
		t.Errorf("relationship = %q, want derived", src.Relationship)
	}
	if src.LastCheckedSHA != "abc123" {
		t.Errorf("last_checked_sha = %q, want abc123", src.LastCheckedSHA)
	}
	if len(src.Paths.High) != 1 || src.Paths.High[0] != "Sources/**/*.swift" {
		t.Errorf("paths.high = %v, want [Sources/**/*.swift]", src.Paths.High)
	}
	if len(src.Paths.Medium) != 2 {
		t.Errorf("paths.medium length = %d, want 2", len(src.Paths.Medium))
	}
	if len(src.Paths.Low) != 2 {
		t.Errorf("paths.low length = %d, want 2", len(src.Paths.Low))
	}
}

func TestLoadDefaultBranch(t *testing.T) {
	yaml := `upstream:
  sources:
    - repo: owner/name
      relationship: watch
      paths:
        high:
          - "*.go"
`
	path := writeTempConfig(t, yaml)
	_, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sources[0].Branch != "main" {
		t.Errorf("default branch = %q, want main", cfg.Sources[0].Branch)
	}
}

func TestSavePreservesOtherSections(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	doc, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the config.
	cfg.Sources[0].LastCheckedSHA = "def456"
	cfg.Sources[0].LastCheckedDate = "2026-02-20T10:00:00Z"

	if err := Save(doc, cfg); err != nil {
		t.Fatal(err)
	}

	// Read back and verify.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Other sections preserved.
	if !strings.Contains(content, "other_tool") {
		t.Error("other_tool section was lost")
	}
	if !strings.Contains(content, "another") {
		t.Error("another section was lost")
	}

	// SHA updated.
	if !strings.Contains(content, "def456") {
		t.Error("SHA was not updated")
	}

	// Re-load to verify structure.
	_, cfg2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.Sources[0].LastCheckedSHA != "def456" {
		t.Errorf("reloaded SHA = %q, want def456", cfg2.Sources[0].LastCheckedSHA)
	}
}

func TestFindSource(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Repo: "owner/name", Branch: "main"},
			{Repo: "other/repo", Branch: "master"},
		},
	}

	// Full match.
	if s := FindSource(cfg, "owner/name"); s == nil || s.Repo != "owner/name" {
		t.Error("full match failed")
	}

	// Short name match.
	if s := FindSource(cfg, "repo"); s == nil || s.Repo != "other/repo" {
		t.Error("short name match failed")
	}

	// No match.
	if s := FindSource(cfg, "nonexistent"); s != nil {
		t.Error("expected nil for nonexistent source")
	}
}
