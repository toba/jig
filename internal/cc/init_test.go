package cc

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestInitAutoDetect(t *testing.T) {
	home := t.TempDir()

	// Source-looking dir (with markers as real files).
	src := filepath.Join(home, ".claude")
	_ = os.MkdirAll(filepath.Join(src, "agents"), 0o755)
	_ = os.MkdirAll(filepath.Join(src, "skills"), 0o755)
	_ = os.WriteFile(filepath.Join(src, ".credentials.json"), []byte("a"), 0o644)
	_ = os.WriteFile(filepath.Join(src, "CLAUDE.md"), []byte("a"), 0o644)

	// Secondary dir with credentials only.
	work := filepath.Join(home, ".claude-work")
	_ = os.MkdirAll(work, 0o755)
	_ = os.WriteFile(filepath.Join(work, ".credentials.json"), []byte("w"), 0o644)
	_ = os.WriteFile(filepath.Join(work, ".claude.json"), []byte("w"), 0o644)
	// Symlinks pointing into source — should NOT be copied.
	_ = os.Symlink(filepath.Join(src, "agents"), filepath.Join(work, "agents"))

	res, err := Init(InitOpts{Home: home})
	if err != nil {
		t.Fatal(err)
	}
	if res.SourceAlias != "main" {
		t.Errorf("source alias: got %q want main", res.SourceAlias)
	}
	if res.SharedSource != src {
		t.Errorf("shared_source: got %q want %q", res.SharedSource, src)
	}
	if !slices.Contains(res.Aliases, "work") {
		t.Errorf("expected work alias, got %v", res.Aliases)
	}
	cfg, err := LoadFrom(res.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Aliases["work"].Path != filepath.Join(home, ".jig", "cc", "work") {
		t.Errorf("work alias path mismatch: %s", cfg.Aliases["work"].Path)
	}
	// Private files should be copied.
	creds := filepath.Join(home, ".jig", "cc", "work", ".credentials.json")
	if _, err := os.Stat(creds); err != nil {
		t.Errorf("private file not copied: %v", err)
	}
	// Symlinks from old dir should NOT be copied as files.
	agents := filepath.Join(home, ".jig", "cc", "work", "agents")
	info, err := os.Lstat(agents)
	if err != nil {
		t.Fatalf("agents missing: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("agents should be a fresh symlink, not a copy")
	}
}

func TestInitRefusesExisting(t *testing.T) {
	home := t.TempDir()
	src := filepath.Join(home, ".claude")
	_ = os.MkdirAll(filepath.Join(src, "agents"), 0o755)
	_ = os.WriteFile(filepath.Join(src, ".credentials.json"), []byte("a"), 0o644)

	if _, err := Init(InitOpts{Home: home}); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(InitOpts{Home: home}); err == nil {
		t.Error("second Init without --force should fail")
	}
	if _, err := Init(InitOpts{Home: home, Force: true}); err != nil {
		t.Errorf("second Init with --force failed: %v", err)
	}
}
