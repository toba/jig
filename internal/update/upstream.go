package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// skillCandidates are paths to look for the PROJECT upstream skill.
var skillCandidates = []string{
	".claude/skills/upstream/SKILL.md",
}

// upstreamConfig mirrors config.Config for YAML generation.
type upstreamConfig struct {
	Sources []upstreamSource `yaml:"sources"`
}

type upstreamSource struct {
	Repo            string       `yaml:"repo"`
	Branch          string       `yaml:"branch"`
	Relationship    string       `yaml:"relationship"`
	Notes           string       `yaml:"notes,omitempty"`
	LastCheckedSHA  string       `yaml:"last_checked_sha,omitempty"`
	LastCheckedDate string       `yaml:"last_checked_date,omitempty"`
	Paths           upstreamPath `yaml:"paths"`
}

// markerEntry represents one repo's entry in last-checked.json.
type markerEntry struct {
	LastCheckedSHA  string `json:"last_checked_sha"`
	LastCheckedDate string `json:"last_checked_date"`
}

type upstreamPath struct {
	High   []string `yaml:"high,omitempty"`
	Medium []string `yaml:"medium,omitempty"`
	Low    []string `yaml:"low,omitempty"`
}

// migrateUpstreamSkill looks for a PROJECT upstream skill file, parses it,
// and generates the upstream: section in .toba.yaml.
// Returns (migrated bool, source path, error).
func migrateUpstreamSkill(tobaPath string) (bool, string, error) {
	// Find the skill file.
	var skillPath string
	var skillData []byte
	for _, c := range skillCandidates {
		data, err := os.ReadFile(c) //nolint:gosec // hardcoded paths
		if err != nil {
			continue
		}
		skillPath = c
		skillData = data
		break
	}
	if skillPath == "" {
		return false, "", nil
	}

	// Check if upstream: already exists in .toba.yaml.
	existing, err := os.ReadFile(tobaPath) //nolint:gosec // path from caller
	if err != nil && !os.IsNotExist(err) {
		return false, "", fmt.Errorf("reading %s: %w", tobaPath, err)
	}
	if sectionExists(splitLines(string(existing)), "upstream") {
		fmt.Fprintf(os.Stderr, "update: skipped upstream — section already exists in %s\n", tobaPath)
		return false, skillPath, nil
	}

	// Parse the skill file.
	content := string(skillData)
	sources, err := parseSkill(content)
	if err != nil {
		return false, skillPath, fmt.Errorf("parsing %s: %w", skillPath, err)
	}
	if len(sources) == 0 {
		return false, skillPath, nil
	}

	// Merge marker data from references/last-checked.json if present.
	markerPath := filepath.Join(filepath.Dir(skillPath), "references", "last-checked.json")
	if markerData, err := os.ReadFile(markerPath); err == nil { //nolint:gosec // path derived from hardcoded skill path
		var markers map[string]markerEntry
		if err := json.Unmarshal(markerData, &markers); err == nil {
			for i := range sources {
				if m, ok := markers[sources[i].Repo]; ok {
					sources[i].LastCheckedSHA = m.LastCheckedSHA
					sources[i].LastCheckedDate = m.LastCheckedDate
				}
			}
		}
	}

	// Generate YAML.
	cfg := upstreamConfig{Sources: sources}
	yamlBytes, err := yaml.Marshal(map[string]upstreamConfig{"upstream": cfg})
	if err != nil {
		return false, skillPath, fmt.Errorf("generating upstream YAML: %w", err)
	}

	// Append to .toba.yaml.
	tobaContent := string(existing)
	if tobaContent != "" && !strings.HasSuffix(tobaContent, "\n\n") {
		if !strings.HasSuffix(tobaContent, "\n") {
			tobaContent += "\n"
		}
		tobaContent += "\n"
	}
	tobaContent += string(yamlBytes)

	if err := os.WriteFile(tobaPath, []byte(tobaContent), 0o644); err != nil {
		return false, skillPath, fmt.Errorf("writing %s: %w", tobaPath, err)
	}

	// Clean up the legacy skill directory.
	skillDir := filepath.Dir(skillPath)
	if err := os.RemoveAll(skillDir); err != nil {
		fmt.Fprintf(os.Stderr, "update: warning: could not remove %s: %v\n", skillDir, err)
	}
	// Remove .claude/skills/ if now empty.
	skillsParent := filepath.Dir(skillDir)
	if entries, err := os.ReadDir(skillsParent); err == nil && len(entries) == 0 {
		_ = os.Remove(skillsParent)
	}

	return true, skillPath, nil
}

// parseSkill extracts upstream source definitions from a SKILL.md file.
func parseSkill(content string) ([]upstreamSource, error) {
	repos := parseRepoTable(content)
	if len(repos) == 0 {
		return nil, nil
	}

	pathsByRepo := parsePathTables(content)

	var sources []upstreamSource
	for _, r := range repos {
		s := upstreamSource{
			Repo:         r.repo,
			Branch:       r.branch,
			Relationship: normalizeRelationship(r.relationship),
			Notes:        r.notes,
		}
		if p, ok := pathsByRepo[r.repo]; ok {
			s.Paths = p
		}
		sources = append(sources, s)
	}
	return sources, nil
}

type repoEntry struct {
	repo         string
	branch       string
	relationship string
	notes        string
}

var backtickRe = regexp.MustCompile("`([^`]+)`")

// parseRepoTable parses the "## Upstream Repos" markdown table.
func parseRepoTable(content string) []repoEntry {
	lines := strings.Split(content, "\n")
	var inTable bool
	var entries []repoEntry

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if inTable {
				break // left the table
			}
			continue
		}

		cells := splitTableRow(line)
		if len(cells) < 3 {
			continue
		}

		// Detect header row.
		lower0 := strings.ToLower(cells[0])
		if strings.Contains(lower0, "repo") {
			inTable = true
			continue
		}
		// Skip separator row.
		if strings.TrimLeft(cells[0], "- ") == "" {
			continue
		}
		if !inTable {
			continue
		}

		repo := extractBacktick(cells[0])
		if repo == "" {
			repo = strings.TrimSpace(cells[0])
		}

		branch := strings.TrimSpace(cells[1])
		if b := extractBacktick(cells[1]); b != "" {
			branch = b
		}

		relationship := strings.TrimSpace(cells[2])

		var notes string
		if len(cells) >= 4 {
			notes = strings.TrimSpace(cells[3])
			// Strip backticks from notes for cleaner YAML.
			notes = backtickRe.ReplaceAllString(notes, "$1")
		}

		entries = append(entries, repoEntry{
			repo:         repo,
			branch:       branch,
			relationship: relationship,
			notes:        notes,
		})
	}
	return entries
}

// parsePathTables parses per-repo path classification tables.
// They appear under headings like "#### owner/repo" or "#### owner/repo (description)".
func parsePathTables(content string) map[string]upstreamPath {
	result := make(map[string]upstreamPath)
	lines := strings.Split(content, "\n")

	for i := range len(lines) {
		line := strings.TrimSpace(lines[i])

		// Look for #### headings containing a repo-like pattern.
		if !strings.HasPrefix(line, "####") {
			continue
		}
		heading := strings.TrimPrefix(line, "####")
		heading = strings.TrimSpace(heading)

		// Extract repo name (owner/name) from heading.
		repo := extractRepoFromHeading(heading)
		if repo == "" {
			continue
		}

		// Find the path table following this heading.
		var paths upstreamPath
		foundTable := false
		for j := i + 1; j < len(lines); j++ {
			tl := strings.TrimSpace(lines[j])
			if strings.HasPrefix(tl, "####") || strings.HasPrefix(tl, "### ") || strings.HasPrefix(tl, "## ") {
				break // next section
			}
			if !strings.HasPrefix(tl, "|") {
				if foundTable {
					break // left the table
				}
				continue
			}

			cells := splitTableRow(tl)
			if len(cells) < 2 {
				continue
			}

			// Skip header/separator rows.
			lower := strings.ToLower(cells[0])
			if strings.Contains(lower, "relevance") || strings.TrimLeft(cells[0], "- ") == "" {
				foundTable = true
				continue
			}
			foundTable = true

			level := strings.ToLower(strings.TrimSpace(cells[0]))
			level = strings.Trim(level, "*") // remove bold markers
			patterns := extractPatterns(cells[1])

			switch {
			case strings.Contains(level, "high"):
				paths.High = append(paths.High, patterns...)
			case strings.Contains(level, "medium"):
				paths.Medium = append(paths.Medium, patterns...)
			case strings.Contains(level, "low"):
				paths.Low = append(paths.Low, patterns...)
			}
		}
		if len(paths.High) > 0 || len(paths.Medium) > 0 || len(paths.Low) > 0 {
			result[repo] = paths
		}
	}
	return result
}

// splitTableRow splits a markdown table row into cells.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// extractBacktick returns the first backtick-quoted string, or "".
func extractBacktick(s string) string {
	m := backtickRe.FindStringSubmatch(s)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

// extractPatterns extracts backtick-quoted patterns from a table cell.
// Patterns are comma-separated, each in backticks: `foo/**`, `bar.txt`
func extractPatterns(cell string) []string {
	matches := backtickRe.FindAllStringSubmatch(cell, -1)
	var patterns []string
	for _, m := range matches {
		if len(m) >= 2 {
			patterns = append(patterns, m[1])
		}
	}
	// Fallback: if no backticks, split by comma and trim.
	if len(patterns) == 0 {
		for p := range strings.SplitSeq(cell, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				patterns = append(patterns, p)
			}
		}
	}
	return patterns
}

var repoPattern = regexp.MustCompile(`[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+`)

// extractRepoFromHeading pulls an "owner/repo" string from a heading.
func extractRepoFromHeading(heading string) string {
	m := repoPattern.FindString(heading)
	return m
}

// normalizeRelationship maps human-readable relationship labels to short keys.
func normalizeRelationship(rel string) string {
	lower := strings.ToLower(rel)
	switch {
	case strings.Contains(lower, "derived"):
		return "derived"
	case strings.Contains(lower, "dependency"):
		return "dependency"
	case strings.Contains(lower, "watch"):
		return "watch"
	default:
		return strings.ToLower(strings.TrimSpace(rel))
	}
}
