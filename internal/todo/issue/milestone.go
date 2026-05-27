package issue

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

// MilestonesDir is the subdirectory under the issues root where milestone files live.
// The issue loader skips this directory so milestones are never treated as issues.
const MilestonesDir = "milestones"

// Milestone is a lightweight, optional grouping an issue may be assigned to.
// It is stored as a markdown file with YAML front matter, like an issue, but
// is NOT an issue and NOT an issue type. Issues reference a milestone by its ID.
type Milestone struct {
	// ID is the unique NanoID identifier (from filename).
	ID string `yaml:"-" json:"id"`
	// Slug is the optional human-readable part of the filename.
	Slug string `yaml:"-" json:"slug,omitempty"`
	// Path is the relative path from the issues root (e.g., "milestones/abc-def--v1.md").
	Path string `yaml:"-" json:"path"`

	// Front matter fields
	Short     string     `yaml:"short" json:"short"` // 2-3 char grid token
	Name      string     `yaml:"name" json:"name"`
	Due       *DueDate   `yaml:"due,omitempty" json:"due,omitempty"`
	CreatedAt *time.Time `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt *time.Time `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`

	// Description is the markdown content after the front matter.
	Description string `yaml:"-" json:"description,omitempty"`

	// Sync holds sync integration metadata keyed by integration name
	// (e.g. github -> milestone_number).
	Sync map[string]map[string]any `yaml:"sync,omitempty" json:"sync,omitempty"`
}

// milestoneFrontMatter is the subset of Milestone serialized to YAML front matter.
type milestoneFrontMatter struct {
	Short     string                    `yaml:"short"`
	Name      string                    `yaml:"name"`
	Due       *DueDate                  `yaml:"due,omitempty"`
	CreatedAt *time.Time                `yaml:"created_at,omitempty"`
	UpdatedAt *time.Time                `yaml:"updated_at,omitempty"`
	Sync      map[string]map[string]any `yaml:"sync,omitempty"`
}

// ParseMilestone reads a milestone from a reader (markdown with YAML front matter).
func ParseMilestone(r io.Reader) (*Milestone, error) {
	var fm milestoneFrontMatter
	body, err := frontmatter.Parse(r, &fm)
	if err != nil {
		return nil, fmt.Errorf("parsing front matter: %w", err)
	}

	desc := strings.TrimSuffix(strings.TrimPrefix(string(body), "\n"), "\n")

	return &Milestone{
		Short:       fm.Short,
		Name:        fm.Name,
		Due:         fm.Due,
		CreatedAt:   fm.CreatedAt,
		UpdatedAt:   fm.UpdatedAt,
		Description: desc,
		Sync:        fm.Sync,
	}, nil
}

// Render serializes the milestone back to markdown with YAML front matter.
func (m *Milestone) Render() ([]byte, error) {
	fm := milestoneFrontMatter{
		Short:     m.Short,
		Name:      m.Name,
		Due:       m.Due,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		Sync:      m.Sync,
	}

	fmBytes, err := yaml.Marshal(&fm)
	if err != nil {
		return nil, fmt.Errorf("marshaling front matter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	if m.ID != "" {
		buf.WriteString("# ")
		buf.WriteString(m.ID)
		buf.WriteString("\n")
	}
	buf.Write(fmBytes)
	buf.WriteString("---\n")
	if m.Description != "" {
		if !strings.HasPrefix(m.Description, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString(m.Description)
		if !strings.HasSuffix(m.Description, "\n") {
			buf.WriteString("\n")
		}
	} else {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// BuildMilestonePath returns the relative path for a milestone file
// (under the milestones subdirectory, not hash-sharded since there are few).
func BuildMilestonePath(id, slug string) string {
	return filepath.Join(MilestonesDir, BuildFilename(id, slug))
}

// HasSync returns true if the milestone has sync data for the given name.
func (m *Milestone) HasSync(name string) bool {
	if m.Sync == nil {
		return false
	}
	_, ok := m.Sync[name]
	return ok
}

// SetSync sets the sync data for a name (full replacement).
func (m *Milestone) SetSync(name string, data map[string]any) {
	if m.Sync == nil {
		m.Sync = make(map[string]map[string]any)
	}
	m.Sync[name] = data
}

// ValidateShort checks that a short name is a non-empty token of at most 3 characters
// with no whitespace, suitable for the TUI grid.
func ValidateShort(short string) error {
	s := strings.TrimSpace(short)
	if s == "" {
		return errors.New("short name cannot be empty")
	}
	if len([]rune(s)) > 3 {
		return fmt.Errorf("short name %q must be at most 3 characters", short)
	}
	if strings.ContainsAny(s, " \t\n") {
		return fmt.Errorf("short name %q must not contain whitespace", short)
	}
	return nil
}
