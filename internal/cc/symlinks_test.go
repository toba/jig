package cc

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func setupSyncTest(t *testing.T) (*Config, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, ".claude")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	// Shared entries.
	for _, n := range []string{"agents", "skills", "CLAUDE.md"} {
		p := filepath.Join(source, n)
		if filepath.Ext(n) == "" {
			_ = os.MkdirAll(p, 0o755)
		} else {
			_ = os.WriteFile(p, []byte("x"), 0o644)
		}
	}
	// Private entries (exist in source, must be ignored).
	_ = os.WriteFile(filepath.Join(source, ".credentials.json"), []byte("secret"), 0o644)

	work := filepath.Join(root, ".jig", "cc", "work")
	c := &Config{
		Version:      1,
		SharedSource: source,
		Private:      DefaultPrivate,
		Aliases: map[string]Alias{
			"main": {CLI: "claude", Path: source, IsSource: true},
			"work": {CLI: "claude", Path: work},
		},
	}
	return c, work
}

func TestSyncCreatesSymlinks(t *testing.T) {
	c, work := setupSyncTest(t)
	rep, err := Sync(c, "work")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"CLAUDE.md", "agents", "skills"}
	got := slices.Clone(rep.Created)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Errorf("created: got %v want %v", got, want)
	}
	for _, n := range want {
		link := filepath.Join(work, n)
		info, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("missing link %s: %v", link, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is not a symlink", link)
		}
	}
	// Private should NOT be linked.
	if _, err := os.Lstat(filepath.Join(work, ".credentials.json")); !os.IsNotExist(err) {
		t.Errorf(".credentials.json should not be linked into alias")
	}
}

func TestSyncIdempotent(t *testing.T) {
	c, _ := setupSyncTest(t)
	if _, err := Sync(c, "work"); err != nil {
		t.Fatal(err)
	}
	rep, err := Sync(c, "work")
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Created) != 0 {
		t.Errorf("second sync should create nothing, got %v", rep.Created)
	}
	if len(rep.Skipped) == 0 {
		t.Error("second sync should skip existing links")
	}
}

func TestSyncRepairsWrongTarget(t *testing.T) {
	c, work := setupSyncTest(t)
	if _, err := Sync(c, "work"); err != nil {
		t.Fatal(err)
	}
	// Repoint a link to a bad target.
	link := filepath.Join(work, "agents")
	_ = os.Remove(link)
	_ = os.Symlink("/nonexistent", link)

	rep, err := Sync(c, "work")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(rep.Repaired, "agents") {
		t.Errorf("expected repaired to include 'agents', got %v", rep.Repaired)
	}
}

func TestSyncReportsConflict(t *testing.T) {
	c, work := setupSyncTest(t)
	_ = os.MkdirAll(work, 0o755)
	// Real file where a symlink should be.
	if err := os.WriteFile(filepath.Join(work, "agents"), []byte("oops"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := Sync(c, "work")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(rep.Conflicts, "agents") {
		t.Errorf("expected conflict on 'agents', got %v", rep.Conflicts)
	}
	// And the real file should still be intact.
	data, _ := os.ReadFile(filepath.Join(work, "agents"))
	if string(data) != "oops" {
		t.Error("conflicted file should not be overwritten")
	}
}

func TestCheckHealth(t *testing.T) {
	c, work := setupSyncTest(t)
	if _, err := Sync(c, "work"); err != nil {
		t.Fatal(err)
	}
	h, err := CheckHealth(c, "work")
	if err != nil {
		t.Fatal(err)
	}
	if h.HasIssues() {
		t.Errorf("clean state should report no issues: %+v", h)
	}

	// Break a symlink.
	link := filepath.Join(work, "agents")
	_ = os.Remove(link)
	_ = os.Symlink("/nonexistent", link)
	h, _ = CheckHealth(c, "work")
	if !slices.Contains(h.Broken, "agents") {
		t.Errorf("expected broken to include 'agents', got %+v", h)
	}
}
