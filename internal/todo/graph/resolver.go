package graph

import (
	"errors"
	"fmt"
	"strings"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/issue"
)

//go:generate go tool gqlgen generate

// Resolver is the root resolver for the GraphQL schema.
// It holds a reference to core.Core for data access.
type Resolver struct {
	Core *core.Core
}

// validateETag checks if the provided ifMatch etag matches the issue's current etag.
// Returns an error if validation fails or if require_if_match is enabled and no etag provided.
func (r *Resolver) validateETag(b *issue.Issue, ifMatch *string) error {
	cfg := r.Core.Config()
	requireIfMatch := cfg != nil && cfg.RequireIfMatch

	// If require_if_match is enabled and no etag provided, reject
	if requireIfMatch && (ifMatch == nil || *ifMatch == "") {
		return &core.ETagRequiredError{}
	}

	// If ifMatch provided, validate it
	if ifMatch != nil && *ifMatch != "" {
		currentETag := b.ETag()
		if currentETag != *ifMatch {
			return &core.ETagMismatchError{Provided: *ifMatch, Current: currentETag}
		}
	}

	return nil
}

// validateParentCompletion guards the transition of an issue into a complete
// status (completed, scrapped, deferred). A parent may only enter such a status
// once all of its children are themselves in a complete status. It is a no-op
// when newStatus is not a complete status or the status is not actually changing
// (so non-status edits to an already-complete issue are never blocked). Callers
// must pass b before its status field is mutated so b.Status reflects the
// current value.
func (r *Resolver) validateParentCompletion(b *issue.Issue, newStatus string) error {
	if newStatus == b.Status || !config.IsCompleteStatus(newStatus) {
		return nil
	}

	var incomplete []string
	for _, child := range r.Core.Children(b.ID) {
		if !config.IsCompleteStatus(child.Status) {
			incomplete = append(incomplete, fmt.Sprintf("%s (%s)", child.ID, child.Status))
		}
	}
	if len(incomplete) > 0 {
		return fmt.Errorf("cannot set %s to %q: %d child issue(s) not complete: %s",
			b.ID, newStatus, len(incomplete), strings.Join(incomplete, ", "))
	}
	return nil
}

// validateAndSetParent validates and sets the parent relationship.
func (r *Resolver) validateAndSetParent(b *issue.Issue, parentID string) error {
	if parentID == "" {
		b.Parent = ""
		return nil
	}

	// Normalise short ID to full ID
	normalizedParent, _ := r.Core.NormalizeID(parentID)

	// Validate parent type hierarchy
	if err := r.Core.ValidateParent(b, normalizedParent); err != nil {
		return err
	}

	// Check for cycles
	if cycle := r.Core.DetectCycle(b.ID, issue.LinkTypeParent, normalizedParent); cycle != nil {
		return fmt.Errorf("setting parent would create cycle: %v", cycle)
	}

	b.Parent = normalizedParent
	return nil
}

// inheritMilestoneFromParent copies the parent's milestone onto the child when
// the child has no milestone of its own. It is a no-op if the issue has no
// parent, already has a milestone, or the parent has none. This lets newly
// parented issues fall into their parent's milestone automatically without
// clobbering an explicit milestone choice on the child.
func (r *Resolver) inheritMilestoneFromParent(b *issue.Issue) {
	if b.Parent == "" || b.Milestone != "" {
		return
	}
	if parent, err := r.Core.Get(b.Parent); err == nil && parent.Milestone != "" {
		b.Milestone = parent.Milestone
	}
}

// validateAndAddBlocking validates and adds blocking relationships.
func (r *Resolver) validateAndAddBlocking(b *issue.Issue, targetIDs []string) error {
	for _, targetID := range targetIDs {
		// Normalise short ID to full ID
		normalizedTargetID, _ := r.Core.NormalizeID(targetID)

		// Validate: cannot block itself
		if normalizedTargetID == b.ID {
			return errors.New("issue cannot block itself")
		}

		// Validate: target must exist
		if _, err := r.Core.Get(normalizedTargetID); err != nil {
			return fmt.Errorf("blocking target issue not found: %s", targetID)
		}

		// Check for cycles in both directions
		if cycle := r.Core.DetectCycle(b.ID, issue.LinkTypeBlocking, normalizedTargetID); cycle != nil {
			return fmt.Errorf("adding blocking relationship would create cycle: %v", cycle)
		}
		if cycle := r.Core.DetectCycle(normalizedTargetID, issue.LinkTypeBlockedBy, b.ID); cycle != nil {
			return fmt.Errorf("adding blocking relationship would create cycle: %v", cycle)
		}

		b.AddBlocking(normalizedTargetID)
	}
	return nil
}

// removeBlockingRelationships removes blocking relationships.
func (r *Resolver) removeBlockingRelationships(b *issue.Issue, targetIDs []string) {
	for _, targetID := range targetIDs {
		normalizedTargetID, _ := r.Core.NormalizeID(targetID)
		b.RemoveBlocking(normalizedTargetID)
	}
}

// validateAndAddBlockedBy validates and adds blocked-by relationships.
func (r *Resolver) validateAndAddBlockedBy(b *issue.Issue, targetIDs []string) error {
	for _, targetID := range targetIDs {
		// Normalise short ID to full ID
		normalizedTargetID, _ := r.Core.NormalizeID(targetID)

		// Validate: cannot be blocked by itself
		if normalizedTargetID == b.ID {
			return errors.New("issue cannot be blocked by itself")
		}

		// Validate: blocker must exist
		if _, err := r.Core.Get(normalizedTargetID); err != nil {
			return fmt.Errorf("blocker issue not found: %s", targetID)
		}

		// Check for cycles in both directions
		if cycle := r.Core.DetectCycle(normalizedTargetID, issue.LinkTypeBlocking, b.ID); cycle != nil {
			return fmt.Errorf("adding blocked-by relationship would create cycle: %v", cycle)
		}
		if cycle := r.Core.DetectCycle(b.ID, issue.LinkTypeBlockedBy, normalizedTargetID); cycle != nil {
			return fmt.Errorf("adding blocked-by relationship would create cycle: %v", cycle)
		}

		b.AddBlockedBy(normalizedTargetID)
	}
	return nil
}

// removeBlockedByRelationships removes blocked-by relationships.
func (r *Resolver) removeBlockedByRelationships(b *issue.Issue, targetIDs []string) {
	for _, targetID := range targetIDs {
		normalizedTargetID, _ := r.Core.NormalizeID(targetID)
		b.RemoveBlockedBy(normalizedTargetID)
	}
}
