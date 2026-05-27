package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/toba/jig/internal/todo/issue"
)

// ErrMilestoneNotFound is returned when a milestone ID does not resolve.
var ErrMilestoneNotFound = errors.New("milestone not found")

// loadMilestonesLocked loads all milestone files from the milestones subdirectory.
// Must be called with the lock held. A missing directory is not an error.
func (c *Core) loadMilestonesLocked() error {
	dir := filepath.Join(c.root, issue.MilestonesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		m, loadErr := loadMilestoneFile(path, c.root)
		if loadErr != nil {
			return fmt.Errorf("loading milestone %s: %w", path, loadErr)
		}
		c.milestones[m.ID] = m
	}
	return nil
}

// loadMilestoneFile reads and parses a single milestone file.
func loadMilestoneFile(path, root string) (*issue.Milestone, error) {
	f, err := os.Open(path) //nolint:gosec // path from known directory
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only file

	m, err := issue.ParseMilestone(f)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}
	m.Path = relPath
	m.ID, m.Slug = issue.ParseFilename(filepath.Base(path))

	if m.CreatedAt == nil {
		if info, statErr := os.Stat(path); statErr == nil {
			t := info.ModTime().UTC().Truncate(time.Second)
			m.CreatedAt = &t
		}
	}
	if m.UpdatedAt == nil {
		m.UpdatedAt = m.CreatedAt
	}
	return m, nil
}

// AllMilestones returns a slice of all milestones.
func (c *Core) AllMilestones() []*issue.Milestone {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*issue.Milestone, 0, len(c.milestones))
	for _, m := range c.milestones {
		result = append(result, m)
	}
	return result
}

// GetMilestone finds a milestone by exact ID match.
func (c *Core) GetMilestone(id string) (*issue.Milestone, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if m, ok := c.milestones[id]; ok {
		return m, nil
	}
	return nil, ErrMilestoneNotFound
}

// MilestoneExists reports whether a milestone with the given ID is loaded.
func (c *Core) MilestoneExists(id string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.milestones[id]
	return ok
}

// CreateMilestone adds a new milestone, generating an ID if needed, and writes it to disk.
func (c *Core) CreateMilestone(m *issue.Milestone) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m.ID == "" {
		m.ID = issue.NewID()
	}
	// Ensure a slug so the filename uses the "--" separator; without it,
	// a hyphenated milestone ID (e.g. "cs3-pmi.md") would be mis-parsed.
	if m.Slug == "" {
		m.Slug = issue.Slugify(m.Short)
		if m.Slug == "" {
			m.Slug = issue.Slugify(m.Name)
		}
	}
	now := time.Now().UTC().Truncate(time.Second)
	m.CreatedAt = &now
	m.UpdatedAt = &now

	if err := c.saveMilestoneToDisk(m); err != nil {
		return err
	}
	c.milestones[m.ID] = m
	return nil
}

// UpdateMilestone modifies an existing milestone and writes it to disk.
func (c *Core) UpdateMilestone(m *issue.Milestone) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.milestones[m.ID]; !ok {
		return ErrMilestoneNotFound
	}
	now := time.Now().UTC().Truncate(time.Second)
	m.UpdatedAt = &now

	if err := c.saveMilestoneToDisk(m); err != nil {
		return err
	}
	c.milestones[m.ID] = m
	return nil
}

// SaveMilestoneSyncOnly persists a milestone whose only changes are sync metadata,
// without bumping updated_at.
func (c *Core) SaveMilestoneSyncOnly(m *issue.Milestone) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.milestones[m.ID]; !ok {
		return ErrMilestoneNotFound
	}
	if err := c.saveMilestoneToDisk(m); err != nil {
		return err
	}
	c.milestones[m.ID] = m
	return nil
}

// DeleteMilestone removes a milestone by ID. It does NOT unassign issues that
// reference it; callers should handle reference cleanup if desired.
func (c *Core) DeleteMilestone(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	m, ok := c.milestones[id]
	if !ok {
		return ErrMilestoneNotFound
	}
	if err := os.Remove(filepath.Join(c.root, m.Path)); err != nil {
		return err
	}
	delete(c.milestones, id)
	return nil
}

// MilestonesSorted returns all milestones ordered by due date (soonest first),
// then by name. Milestones without a due date sort after those with one.
func (c *Core) MilestonesSorted() []*issue.Milestone {
	all := c.AllMilestones()
	slices.SortFunc(all, func(a, b *issue.Milestone) int {
		switch {
		case a.Due == nil && b.Due == nil:
			return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case a.Due == nil:
			return 1
		case b.Due == nil:
			return -1
		}
		if c := a.Due.Compare(b.Due.Time); c != 0 {
			return c
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return all
}

// MilestoneOrder returns a map of milestone ID to sort rank, derived from
// MilestonesSorted (due date then name). Used to sort issues by milestone.
func (c *Core) MilestoneOrder() map[string]int {
	sorted := c.MilestonesSorted()
	order := make(map[string]int, len(sorted))
	for i, m := range sorted {
		order[m.ID] = i
	}
	return order
}

// saveMilestoneToDisk writes a milestone to the filesystem.
func (c *Core) saveMilestoneToDisk(m *issue.Milestone) error {
	var path string
	if m.Path != "" {
		path = filepath.Join(c.root, m.Path)
	} else {
		relPath := issue.BuildMilestonePath(m.ID, m.Slug)
		path = filepath.Join(c.root, relPath)
		m.Path = relPath
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	content, err := m.Render()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}
