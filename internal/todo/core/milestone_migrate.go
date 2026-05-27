package core

import (
	"strconv"
	"strings"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/issue"
)

// MilestoneMigration records a single converted milestone-type issue.
type MilestoneMigration struct {
	OldIssueID     string   `json:"old_issue_id"`
	NewMilestoneID string   `json:"new_milestone_id"`
	Short          string   `json:"short"`
	Name           string   `json:"name"`
	ChildIDs       []string `json:"child_ids"`
}

// MigrateMilestoneTypeIssues converts legacy `type: milestone` issues into
// first-class milestone entities. For each such issue it:
//   - creates a milestone entity (name = title, short = derived, due, description = body,
//     carrying over any github milestone_number sync metadata),
//   - sets each direct child's Milestone to the new entity and clears the child's Parent,
//   - deletes the old milestone-type issue.
//
// It is idempotent (no milestone-type issues => no changes). With dryRun=true it
// reports what would change without writing anything.
func (c *Core) MigrateMilestoneTypeIssues(dryRun bool) ([]MilestoneMigration, error) {
	// Snapshot the relevant issues under read lock.
	c.mu.RLock()
	var legacy []*issue.Issue
	childrenByParent := map[string][]string{}
	usedShorts := map[string]bool{}
	for _, b := range c.issues {
		if b.Type == config.TypeMilestone {
			legacy = append(legacy, b)
		}
		if b.Parent != "" {
			childrenByParent[b.Parent] = append(childrenByParent[b.Parent], b.ID)
		}
	}
	for _, m := range c.milestones {
		usedShorts[m.Short] = true
	}
	c.mu.RUnlock()

	var migrations []MilestoneMigration
	for _, old := range legacy {
		short := deriveShort(old.Title, usedShorts)
		usedShorts[short] = true

		mig := MilestoneMigration{
			OldIssueID: old.ID,
			Short:      short,
			Name:       old.Title,
			ChildIDs:   childrenByParent[old.ID],
		}

		if dryRun {
			migrations = append(migrations, mig)
			continue
		}

		m := &issue.Milestone{
			Short:       short,
			Name:        old.Title,
			Due:         old.Due,
			Description: old.Body,
		}
		// Carry over the GitHub milestone number, if any.
		if old.Sync != nil {
			if gh, ok := old.Sync["github"]; ok {
				if num, ok := gh["milestone_number"]; ok {
					m.SetSync("github", map[string]any{"milestone_number": num})
				}
			}
		}
		if err := c.CreateMilestone(m); err != nil {
			return migrations, err
		}
		mig.NewMilestoneID = m.ID

		// Reassign direct children: set milestone, clear parent.
		for _, childID := range mig.ChildIDs {
			child, err := c.Get(childID)
			if err != nil {
				continue
			}
			child.Milestone = m.ID
			child.Parent = ""
			if err := c.Update(child, nil); err != nil {
				return migrations, err
			}
		}

		// Remove the old milestone-type issue.
		if err := c.Delete(old.ID); err != nil {
			return migrations, err
		}

		migrations = append(migrations, mig)
	}

	return migrations, nil
}

// deriveShort builds a 2-3 char short name from a title, avoiding collisions.
func deriveShort(title string, used map[string]bool) string {
	slug := issue.Slugify(title)
	slug = strings.ReplaceAll(slug, "-", "")
	base := slug
	if len(base) > 3 {
		base = base[:3]
	}
	if base == "" {
		base = "ms"
	}
	if !used[base] {
		return base
	}
	// Append a digit to disambiguate, keeping within 3 chars.
	stem := base
	if len(stem) > 2 {
		stem = stem[:2]
	}
	for i := 1; i < 100; i++ {
		cand := stem + strconv.Itoa(i)
		if len(cand) <= 3 && !used[cand] {
			return cand
		}
	}
	return base
}
