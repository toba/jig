package update

import (
	"fmt"
	"os"
	"strings"
)

type migration struct {
	sectionKey string
	candidates []string
}

var migrations = []migration{
	{sectionKey: "nope", candidates: []string{".claude/nope.yml", ".claude/nope.yaml"}},
	{sectionKey: "todo", candidates: []string{".todo.yml", ".todo.yaml"}},
}

// Run discovers legacy config files and merges them into jigPath.
func Run(jigPath string) error {
	type found struct {
		migration
		path    string
		content []byte
	}

	// Discover legacy files.
	var hits []found
	for _, m := range migrations {
		for _, c := range m.candidates {
			data, err := os.ReadFile(c) //nolint:gosec // candidate paths are hardcoded
			if err != nil {
				continue
			}
			hits = append(hits, found{migration: m, path: c, content: data})
			break // first hit wins
		}
	}

	// Migrate upstream skill (parsed from SKILL.md, not a verbatim copy).
	upMigrated, upPath, err := migrateUpstreamSkill(jigPath)
	if err != nil {
		return fmt.Errorf("upstream skill migration: %w", err)
	}
	if upMigrated {
		fmt.Fprintf(os.Stderr, "update: migrated %s → %s (upstream section)\n", upPath, jigPath)
	}

	// Migrate commit command (scripts/commit.sh → jig commit).
	commitMigrated, err := migrateCommitCommand(jigPath)
	if err != nil {
		return fmt.Errorf("commit command migration: %w", err)
	}

	// Migrate legacy todo config (issues: + sync: → todo:).
	todoMigrated, err := migrateTodoConfig(jigPath)
	if err != nil {
		return fmt.Errorf("todo config migration: %w", err)
	}

	if len(hits) == 0 && !upMigrated && !commitMigrated && !todoMigrated {
		fmt.Fprintln(os.Stderr, "update: no legacy config files found")
		return nil
	}

	if len(hits) == 0 {
		return nil
	}

	// Read existing .jig.yaml (may have been updated by upstream migration).
	existing, err := os.ReadFile(jigPath) //nolint:gosec // path from caller
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", jigPath, err)
	}
	tobaContent := string(existing)

	var migrated []found
	lines := splitLines(tobaContent)

	for _, h := range hits {
		if sectionExists(lines, h.sectionKey) {
			fmt.Fprintf(os.Stderr, "update: skipped %s — section already exists in %s\n", h.sectionKey, jigPath)
			continue
		}
		migrated = append(migrated, h)
	}

	if len(migrated) == 0 {
		return nil
	}

	// Append migrated content.
	for _, h := range migrated {
		// Ensure trailing newline + blank separator.
		if tobaContent != "" && !strings.HasSuffix(tobaContent, "\n\n") {
			if !strings.HasSuffix(tobaContent, "\n") {
				tobaContent += "\n"
			}
			tobaContent += "\n"
		}
		tobaContent += string(wrapInSection(h.sectionKey, h.content))
		// Ensure content ends with newline.
		if !strings.HasSuffix(tobaContent, "\n") {
			tobaContent += "\n"
		}
	}

	// Write .jig.yaml.
	if err := os.WriteFile(jigPath, []byte(tobaContent), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", jigPath, err)
	}

	// Cleanup legacy files and report.
	for _, h := range migrated {
		if err := os.Remove(h.path); err != nil {
			fmt.Fprintf(os.Stderr, "update: warning: could not remove %s: %v\n", h.path, err)
		}
		fmt.Fprintf(os.Stderr, "update: migrated %s → %s (%s section)\n", h.path, jigPath, h.sectionKey)
	}

	return nil
}

// wrapInSection wraps content under sectionKey if it doesn't already start with it.
func wrapInSection(key string, content []byte) []byte {
	prefix := key + ":"
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && (lines[0] == prefix || strings.HasPrefix(lines[0], prefix+" ")) {
		return content // already wrapped
	}
	var b strings.Builder
	b.WriteString(prefix + "\n")
	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}
	return []byte(strings.TrimRight(b.String(), "\n") + "\n")
}

// splitLines splits s into lines, preserving empty lines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// sectionExists checks if sectionKey appears as a top-level YAML key.
func sectionExists(lines []string, key string) bool {
	prefix := key + ":"
	for _, l := range lines {
		if l == prefix || strings.HasPrefix(l, prefix+" ") {
			return true
		}
	}
	return false
}
