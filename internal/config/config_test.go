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

citations:
  - repo: owner/name
    branch: main
    scope: "Tools subsystem — file inspection, path matching"
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

	if len(*cfg) != 1 {
		t.Fatalf("expected 1 source, got %d", len(*cfg))
	}

	src := (*cfg)[0]
	if src.Repo != "owner/name" {
		t.Errorf("repo = %q, want owner/name", src.Repo)
	}
	if src.Branch != "main" {
		t.Errorf("branch = %q, want main", src.Branch)
	}
	if src.Scope != "Tools subsystem — file inspection, path matching" {
		t.Errorf("scope = %q, want Tools subsystem — file inspection, path matching", src.Scope)
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
	yaml := `citations:
  - repo: owner/name
    paths:
      high:
        - "*.go"
`
	path := writeTempConfig(t, yaml)
	_, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if (*cfg)[0].Branch != "main" {
		t.Errorf("default branch = %q, want main", (*cfg)[0].Branch)
	}
}

func TestSavePreservesOtherSections(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	doc, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Modify the config.
	(*cfg)[0].LastCheckedSHA = "def456"
	(*cfg)[0].LastCheckedDate = "2026-02-20T10:00:00Z"

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
	if (*cfg2)[0].LastCheckedSHA != "def456" {
		t.Errorf("reloaded SHA = %q, want def456", (*cfg2)[0].LastCheckedSHA)
	}
}

func TestAppendSource(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	newSrc := Source{
		Repo:   "other/project",
		Branch: "main",
		Notes:  "A new project",
		Paths: PathDefs{
			High: []string{"**/*.go"},
			Low:  []string{"README.md"},
		},
	}

	if err := AppendSource(doc, newSrc); err != nil {
		t.Fatal(err)
	}

	// Re-load and verify.
	_, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(*cfg) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(*cfg))
	}
	if (*cfg)[1].Repo != "other/project" {
		t.Errorf("repo = %q, want other/project", (*cfg)[1].Repo)
	}
	if (*cfg)[1].Notes != "A new project" {
		t.Errorf("notes = %q, want A new project", (*cfg)[1].Notes)
	}

	// Verify other sections preserved.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "other_tool") {
		t.Error("other_tool section was lost")
	}
	if !strings.Contains(content, "another") {
		t.Error("another section was lost")
	}
}

func TestAppendSourceNoExistingCitations(t *testing.T) {
	yaml := `other_tool:
  setting: value
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	newSrc := Source{
		Repo:   "new/repo",
		Branch: "main",
		Paths: PathDefs{
			High: []string{"**/*.rs"},
		},
	}

	if err := AppendSource(doc, newSrc); err != nil {
		t.Fatal(err)
	}

	// Re-load and verify.
	_, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(*cfg) != 1 {
		t.Fatalf("expected 1 source, got %d", len(*cfg))
	}
	if (*cfg)[0].Repo != "new/repo" {
		t.Errorf("repo = %q, want new/repo", (*cfg)[0].Repo)
	}
}

func TestHasSource(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if !HasSource(doc, "owner/name") {
		t.Error("expected HasSource to return true for existing repo")
	}
	if HasSource(doc, "other/missing") {
		t.Error("expected HasSource to return false for missing repo")
	}
}

func TestTracksReleases(t *testing.T) {
	s := Source{Track: "releases"}
	if !s.TracksReleases() {
		t.Error("expected TracksReleases() = true for track=releases")
	}
	s2 := Source{}
	if s2.TracksReleases() {
		t.Error("expected TracksReleases() = false for empty track")
	}
}

func TestMarkSourceRelease(t *testing.T) {
	src := &Source{Repo: "owner/repo"}
	MarkSourceRelease(src, "v1.2.0", "abc123")
	if src.LastCheckedTag != "v1.2.0" {
		t.Errorf("tag = %q, want v1.2.0", src.LastCheckedTag)
	}
	if src.LastCheckedSHA != "abc123" {
		t.Errorf("sha = %q, want abc123", src.LastCheckedSHA)
	}
	if src.LastCheckedDate == "" {
		t.Error("expected date to be set")
	}
}

func TestLoadReleaseTrackedSource(t *testing.T) {
	yaml := `citations:
  - repo: owner/lib
    track: releases
    last_checked_tag: v1.0.0
    last_checked_sha: abc123
    paths:
      high:
        - "**/*.go"
`
	path := writeTempConfig(t, yaml)
	_, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	src := (*cfg)[0]
	if !src.TracksReleases() {
		t.Error("expected TracksReleases() = true")
	}
	if src.LastCheckedTag != "v1.0.0" {
		t.Errorf("tag = %q, want v1.0.0", src.LastCheckedTag)
	}
	// Branch should remain empty for release-tracked sources (not defaulted).
	if src.Branch != "" {
		t.Errorf("branch = %q, want empty for release-tracked source", src.Branch)
	}
}

func TestSaveReleaseTrackedRoundTrip(t *testing.T) {
	yaml := `citations:
  - repo: owner/lib
    track: releases
    last_checked_tag: v1.0.0
    last_checked_sha: abc123
    paths:
      high:
        - "**/*.go"
`
	path := writeTempConfig(t, yaml)
	doc, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	MarkSourceRelease(&(*cfg)[0], "v2.0.0", "def456")
	if err := Save(doc, cfg); err != nil {
		t.Fatal(err)
	}

	_, cfg2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	src := (*cfg2)[0]
	if src.LastCheckedTag != "v2.0.0" {
		t.Errorf("tag = %q, want v2.0.0", src.LastCheckedTag)
	}
	if src.LastCheckedSHA != "def456" {
		t.Errorf("sha = %q, want def456", src.LastCheckedSHA)
	}
	if src.Track != "releases" {
		t.Errorf("track = %q, want releases", src.Track)
	}
}

func TestSaveUsesFlowStylePaths(t *testing.T) {
	yaml := `citations:
  - repo: owner/name
    branch: main
    last_checked_sha: abc123
    paths:
      high: ["**/*.go"]
      medium: [go.mod, go.sum]
      low: [".github/**", README.md, LICENSE]
`
	path := writeTempConfig(t, yaml)
	doc, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Modify and save.
	(*cfg)[0].LastCheckedSHA = "def456"
	if err := Save(doc, cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	// Paths must be in flow style [a, b], not bullet-list style.
	if !strings.Contains(content, `['**/*.go']`) {
		t.Errorf("expected flow-style high paths, got:\n%s", content)
	}
	if !strings.Contains(content, `[go.mod, go.sum]`) {
		t.Errorf("expected flow-style medium paths, got:\n%s", content)
	}
	if !strings.Contains(content, `[.github/**, README.md, LICENSE]`) {
		t.Errorf("expected flow-style low paths, got:\n%s", content)
	}
}

func TestUpdateSource(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	doc, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	// Update several fields.
	src := FindSource(cfg, "owner/name")
	if src == nil {
		t.Fatal("source not found")
	}

	src.Branch = "develop"
	src.Scope = "Updated scope"
	src.Notes = "Updated notes"
	src.Track = "releases"
	src.Paths.High = []string{"**/*.rs"}
	src.Paths.Medium = []string{"Cargo.toml"}

	if err := Save(doc, cfg); err != nil {
		t.Fatal(err)
	}

	// Re-load and verify all fields persisted.
	_, cfg2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := FindSource(cfg2, "owner/name")
	if updated == nil {
		t.Fatal("source not found after save")
	}
	if updated.Branch != "develop" {
		t.Errorf("branch = %q, want develop", updated.Branch)
	}
	if updated.Scope != "Updated scope" {
		t.Errorf("scope = %q, want Updated scope", updated.Scope)
	}
	if updated.Notes != "Updated notes" {
		t.Errorf("notes = %q, want Updated notes", updated.Notes)
	}
	if updated.Track != "releases" {
		t.Errorf("track = %q, want releases", updated.Track)
	}
	if len(updated.Paths.High) != 1 || updated.Paths.High[0] != "**/*.rs" {
		t.Errorf("paths.high = %v, want [**/*.rs]", updated.Paths.High)
	}
	if len(updated.Paths.Medium) != 1 || updated.Paths.Medium[0] != "Cargo.toml" {
		t.Errorf("paths.medium = %v, want [Cargo.toml]", updated.Paths.Medium)
	}

	// Verify other sections preserved.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "other_tool") {
		t.Error("other_tool section was lost")
	}
	if !strings.Contains(content, "another") {
		t.Error("another section was lost")
	}
}

func TestUpdateSourceRenameRepo(t *testing.T) {
	path := writeTempConfig(t, testYAML)
	doc, cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	src := FindSource(cfg, "owner/name")
	if src == nil {
		t.Fatal("source not found")
	}
	src.Repo = "newowner/newname"

	if err := Save(doc, cfg); err != nil {
		t.Fatal(err)
	}

	_, cfg2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if FindSource(cfg2, "newowner/newname") == nil {
		t.Error("renamed source not found")
	}
	if FindSource(cfg2, "owner/name") != nil {
		t.Error("old source name should not exist")
	}
}

func TestFindSource(t *testing.T) {
	cfg := &Config{
		{Repo: "owner/name", Branch: "main"},
		{Repo: "other/repo", Branch: "master"},
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
