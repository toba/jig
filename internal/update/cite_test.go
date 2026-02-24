package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSkill = `---
name: upstream
description: Check upstream repos
---

# Upstream Change Tracker

## Upstream Repos

| Repo | Default Branch | Relationship | Derived Into / Used By |
|------|---------------|-------------|----------------------|
| ` + "`acme/widgets`" + ` | ` + "`main`" + ` | Derived code | ` + "`src/widgets/`" + ` |
| ` + "`acme/gadgets`" + ` | ` + "`develop`" + ` | Dependency (pinned 1.2.3) | Gadget library |
| ` + "`other/tools`" + ` | ` + "`main`" + ` | Feature watch | Monitor for ideas |

## Workflow

### Step 3: Classify Changed Files by Relevance

#### acme/widgets

| Relevance | Path Patterns |
|-----------|--------------|
| **HIGH** | ` + "`src/**/*.go`" + ` |
| **MEDIUM** | ` + "`go.mod`" + `, ` + "`go.sum`" + ` |
| **LOW** | ` + "`.github/**`" + `, ` + "`README.md`" + ` |

#### acme/gadgets (dependency watch)

| Relevance | Path Patterns |
|-----------|--------------|
| **HIGH** | ` + "`Sources/**`" + ` |
| **MEDIUM** | ` + "`CHANGELOG.md`" + `, ` + "`Package.swift`" + ` |
| **LOW** | ` + "`.github/**`" + ` |

#### other/tools (feature watch)

| Relevance | Path Patterns |
|-----------|--------------|
| **HIGH** | ` + "`src/**/*Tool*.go`" + ` |
| **LOW** | ` + "`README.md`" + ` |
`

func TestParseSkill(t *testing.T) {
	sources := parseSkill(sampleSkill)
	if len(sources) != 3 {
		t.Fatalf("expected 3 sources, got %d", len(sources))
	}

	// First source.
	s := sources[0]
	if s.Repo != "acme/widgets" {
		t.Errorf("repo = %q, want acme/widgets", s.Repo)
	}
	if s.Branch != "main" {
		t.Errorf("branch = %q, want main", s.Branch)
	}
	if s.Notes != "src/widgets/" {
		t.Errorf("notes = %q, want src/widgets/", s.Notes)
	}
	if len(s.Paths.High) != 1 || s.Paths.High[0] != "src/**/*.go" {
		t.Errorf("high = %v, want [src/**/*.go]", s.Paths.High)
	}
	if len(s.Paths.Medium) != 2 {
		t.Errorf("medium = %v, want [go.mod go.sum]", s.Paths.Medium)
	}
	if len(s.Paths.Low) != 2 {
		t.Errorf("low = %v, want [.github/** README.md]", s.Paths.Low)
	}

	// Second source.
	s = sources[1]
	if s.Branch != "develop" {
		t.Errorf("branch = %q, want develop", s.Branch)
	}

	// Third source.
	if sources[2].Repo != "other/tools" {
		t.Errorf("repo = %q, want other/tools", sources[2].Repo)
	}
}

func TestParseRepoTable(t *testing.T) {
	entries := parseRepoTable(sampleSkill)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].repo != "acme/widgets" {
		t.Errorf("repo[0] = %q", entries[0].repo)
	}
	if entries[1].branch != "develop" {
		t.Errorf("branch[1] = %q", entries[1].branch)
	}
}

func TestParsePathTables(t *testing.T) {
	paths := parsePathTables(sampleSkill)
	if len(paths) != 3 {
		t.Fatalf("expected 3 repos in path tables, got %d", len(paths))
	}

	w := paths["acme/widgets"]
	if len(w.High) != 1 || w.High[0] != "src/**/*.go" {
		t.Errorf("widgets high = %v", w.High)
	}
	if len(w.Medium) != 2 {
		t.Errorf("widgets medium = %v", w.Medium)
	}
	if len(w.Low) != 2 {
		t.Errorf("widgets low = %v", w.Low)
	}

	g := paths["acme/gadgets"]
	if len(g.High) != 1 || g.High[0] != "Sources/**" {
		t.Errorf("gadgets high = %v", g.High)
	}

	o := paths["other/tools"]
	if len(o.High) != 1 || o.High[0] != "src/**/*Tool*.go" {
		t.Errorf("tools high = %v", o.High)
	}
	if len(o.Medium) != 0 {
		t.Errorf("tools medium = %v, want empty", o.Medium)
	}
}

func TestMigrateCiteSkill(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
		check func(t *testing.T, dir string)
	}{
		{
			name: "parses SKILL.md and creates citations section",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				mkdir(t, filepath.Join(dir, ".claude/skills/upstream"))
				writeFile(t, filepath.Join(dir, ".claude/skills/upstream/SKILL.md"), sampleSkill)
			},
			check: func(t *testing.T, dir string) {
				t.Helper()
				data := readFile(t, filepath.Join(dir, ".jig.yaml"))
				if !strings.Contains(data, "citations:") {
					t.Error("missing citations: section")
				}
				if !strings.Contains(data, "acme/widgets") {
					t.Error("missing acme/widgets repo")
				}
				if !strings.Contains(data, "acme/gadgets") {
					t.Error("missing acme/gadgets repo")
				}
				if !strings.Contains(data, "src/**/*.go") {
					t.Error("missing path pattern")
				}
				// Skill directory should be cleaned up.
				if _, err := os.Stat(filepath.Join(dir, ".claude/skills/upstream")); !os.IsNotExist(err) {
					t.Error(".claude/skills/upstream should have been removed")
				}
			},
		},
		{
			name: "migrates marker data from last-checked.json",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				mkdir(t, filepath.Join(dir, ".claude/skills/upstream/references"))
				writeFile(t, filepath.Join(dir, ".claude/skills/upstream/SKILL.md"), sampleSkill)
				writeFile(t, filepath.Join(dir, ".claude/skills/upstream/references/last-checked.json"), `{
  "acme/widgets": {
    "last_checked_sha": "abc123",
    "last_checked_date": "2026-01-20T10:00:00Z"
  }
}`)
			},
			check: func(t *testing.T, dir string) {
				t.Helper()
				data := readFile(t, filepath.Join(dir, ".jig.yaml"))
				if !strings.Contains(data, "last_checked_sha: abc123") {
					t.Errorf("missing last_checked_sha in output:\n%s", data)
				}
				if !strings.Contains(data, "last_checked_date: \"2026-01-20T10:00:00Z\"") {
					t.Errorf("missing last_checked_date in output:\n%s", data)
				}
				// No marker fields for acme/gadgets (not in JSON).
				// Skill directory should be cleaned up.
				if _, err := os.Stat(filepath.Join(dir, ".claude/skills/upstream")); !os.IsNotExist(err) {
					t.Error(".claude/skills/upstream should have been removed")
				}
			},
		},
		{
			name: "removes empty .claude/skills parent",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				mkdir(t, filepath.Join(dir, ".claude/skills/upstream"))
				writeFile(t, filepath.Join(dir, ".claude/skills/upstream/SKILL.md"), sampleSkill)
			},
			check: func(t *testing.T, dir string) {
				t.Helper()
				if _, err := os.Stat(filepath.Join(dir, ".claude/skills")); !os.IsNotExist(err) {
					t.Error(".claude/skills should have been removed when empty")
				}
				// .claude itself should remain (might have other files).
			},
		},
		{
			name: "skips when citations section already exists",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				writeFile(t, filepath.Join(dir, ".jig.yaml"), "citations: []\n")
				mkdir(t, filepath.Join(dir, ".claude/skills/upstream"))
				writeFile(t, filepath.Join(dir, ".claude/skills/upstream/SKILL.md"), sampleSkill)
			},
			check: func(t *testing.T, dir string) {
				t.Helper()
				data := readFile(t, filepath.Join(dir, ".jig.yaml"))
				if data != "citations: []\n" {
					t.Error(".jig.yaml should not have been modified")
				}
				// Skill directory should be cleaned up even when migration is skipped.
				if _, err := os.Stat(filepath.Join(dir, ".claude/skills/upstream")); !os.IsNotExist(err) {
					t.Error(".claude/skills/upstream should have been removed")
				}
			},
		},
		{
			name: "no skill file — no error, no citations section",
			setup: func(t *testing.T, dir string) {
				t.Helper()
			},
			check: func(t *testing.T, dir string) {
				t.Helper()
				if _, err := os.Stat(filepath.Join(dir, ".jig.yaml")); err == nil {
					t.Error(".jig.yaml should not have been created")
				}
			},
		},
		{
			name: "appends citations to existing .jig.yaml content",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				writeFile(t, filepath.Join(dir, ".jig.yaml"), "nope:\n  rules: []\n")
				mkdir(t, filepath.Join(dir, ".claude/skills/upstream"))
				writeFile(t, filepath.Join(dir, ".claude/skills/upstream/SKILL.md"), sampleSkill)
			},
			check: func(t *testing.T, dir string) {
				t.Helper()
				data := readFile(t, filepath.Join(dir, ".jig.yaml"))
				if !strings.Contains(data, "nope:") {
					t.Error("existing nope section was lost")
				}
				if !strings.Contains(data, "citations:") {
					t.Error("citations section not appended")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("warning: could not restore dir: %v", err)
				}
			}()

			tc.setup(t, dir)

			jigPath := filepath.Join(dir, ".jig.yaml")
			migrated, _, mErr := migrateCiteSkill(jigPath)
			if mErr != nil {
				t.Fatalf("migrateCiteSkill: %v", mErr)
			}
			_ = migrated
			tc.check(t, dir)
		})
	}
}
