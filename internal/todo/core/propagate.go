package core

import (
	"time"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/issue"
)

// findChildrenLocked returns all issues whose Parent matches parentID.
// Must be called with c.mu held.
func (c *Core) findChildrenLocked(parentID string) []*issue.Issue {
	var children []*issue.Issue
	for _, b := range c.issues {
		if b.Parent == parentID {
			children = append(children, b)
		}
	}
	return children
}

// computeParentStatus determines what the parent's status should be based on
// its children's statuses. Returns "" if no change is needed.
//
// Rules (evaluated most-specific first):
//  1. All children scrapped → scrapped
//  2. All children completed or scrapped (≥1 completed) → completed
//  3. All children review or completed (≥1 review) → review
//  4. Any child in-progress → in-progress (only if parent is ready or draft)
func computeParentStatus(parent *issue.Issue, children []*issue.Issue) string {
	if len(children) == 0 {
		return ""
	}

	var (
		allScrapped            = true
		allCompletedOrScrapped = true
		allReviewOrCompleted   = true
		hasCompleted           bool
		hasReview              bool
		hasInProgress          bool
	)

	for _, child := range children {
		switch child.Status {
		case config.StatusScrapped:
			// contributes to allScrapped, allCompletedOrScrapped, but breaks allReviewOrCompleted
			allReviewOrCompleted = false
		case config.StatusCompleted:
			allScrapped = false
			hasCompleted = true
		case config.StatusReview:
			allScrapped = false
			allCompletedOrScrapped = false
			hasReview = true
		case config.StatusInProgress:
			allScrapped = false
			allCompletedOrScrapped = false
			allReviewOrCompleted = false
			hasInProgress = true
		default: // ready, draft, or anything else
			allScrapped = false
			allCompletedOrScrapped = false
			allReviewOrCompleted = false
		}
	}

	// Rule 4: all scrapped
	if allScrapped {
		if parent.Status != config.StatusScrapped {
			return config.StatusScrapped
		}
		return ""
	}

	// Rule 3: all completed or scrapped, at least one completed
	if allCompletedOrScrapped && hasCompleted {
		if parent.Status != config.StatusCompleted {
			return config.StatusCompleted
		}
		return ""
	}

	// Rule 2: all review or completed, at least one review
	if allReviewOrCompleted && hasReview {
		if parent.Status != config.StatusReview {
			return config.StatusReview
		}
		return ""
	}

	// Rule 1: any child in-progress, only if parent is ready or draft
	if hasInProgress {
		if parent.Status == config.StatusReady || parent.Status == config.StatusDraft {
			return config.StatusInProgress
		}
	}

	return ""
}

// propagateStatusLocked walks up the parent chain, updating each ancestor's
// status based on its children. Must be called with c.mu held.
// The visited map prevents infinite loops from cyclic parent references.
func (c *Core) propagateStatusLocked(issueID string, visited map[string]bool) {
	b, ok := c.issues[issueID]
	if !ok || b.Parent == "" {
		return
	}

	parent, ok := c.issues[b.Parent]
	if !ok {
		return // broken parent link, skip silently
	}

	if visited[parent.ID] {
		return // cycle detected
	}
	visited[parent.ID] = true

	children := c.findChildrenLocked(parent.ID)
	newStatus := computeParentStatus(parent, children)
	if newStatus == "" {
		return // no change needed
	}

	parent.Status = newStatus
	now := time.Now().UTC().Truncate(time.Second)
	parent.UpdatedAt = &now

	// Persist to disk (best-effort — don't fail the original update)
	if err := c.saveToDisk(parent); err != nil {
		c.logWarn("failed to save propagated status for %s: %v", parent.ID, err)
		return
	}

	// Update search index if active
	if c.searchIndex != nil {
		if err := c.searchIndex.IndexIssue(parent); err != nil {
			c.logWarn("failed to update search index for propagated %s: %v", parent.ID, err)
		}
	}

	// Recurse up the hierarchy
	c.propagateStatusLocked(parent.ID, visited)
}
