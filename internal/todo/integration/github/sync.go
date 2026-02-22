package github

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/integration/syncutil"
	"github.com/toba/jig/internal/todo/issue"
	"golang.org/x/sync/errgroup"
)

// Syncer handles syncing issues to GitHub issues.
type Syncer struct {
	client    *Client
	config    *Config
	opts      SyncOptions
	core      *core.Core
	syncStore SyncStateProvider

	// Tracking for relationship pass
	mu              sync.RWMutex
	issueToGHNumber map[string]int      // local issue ID -> GitHub issue number
	childrenOf      map[string][]string // parent ID -> child IDs (built during sync)
}

// NewSyncer creates a new syncer with the given client and options.
func NewSyncer(client *Client, cfg *Config, opts SyncOptions, c *core.Core, syncStore SyncStateProvider) *Syncer {
	return &Syncer{
		client:          client,
		config:          cfg,
		opts:            opts,
		core:            c,
		syncStore:       syncStore,
		issueToGHNumber: make(map[string]int),
		childrenOf:      make(map[string][]string),
	}
}

// SyncIssues syncs a list of issues to GitHub issues.
// Uses a multi-pass approach:
// 1. Create/update parent issues (issues without parents, or parents not in this sync)
// 2. Create/update child issues with sub-issue relationships
// 3. Update relationship references in issue bodies (Parent, Children, Blocks, Blocked by)
func (s *Syncer) SyncIssues(ctx context.Context, issues []*issue.Issue) ([]SyncResult, error) {
	// Pre-fetch authenticated user to avoid per-issue API calls
	if _, err := s.client.GetAuthenticatedUser(ctx); err != nil {
		_ = err // Non-fatal
	}

	// Pre-populate label cache
	if err := s.client.PopulateLabelCache(ctx); err != nil {
		_ = err // Non-fatal
	}

	// Ensure all labels that will be needed exist
	s.ensureAllLabels(ctx, issues)

	// Pre-populate mapping with already-synced issues from sync store
	for _, b := range issues {
		issueNumber := s.syncStore.GetIssueNumber(b.ID)
		if issueNumber != nil && *issueNumber != 0 {
			s.issueToGHNumber[b.ID] = *issueNumber
		}
	}

	// Build a set of issue IDs being synced and a children index
	syncingIDs := make(map[string]bool)
	for _, b := range issues {
		syncingIDs[b.ID] = true
	}
	for _, b := range issues {
		if b.Parent != "" && syncingIDs[b.Parent] {
			s.childrenOf[b.Parent] = append(s.childrenOf[b.Parent], b.ID)
		}
	}

	// Separate issues into layers: parents first, then children
	var parents, children []*issue.Issue
	for _, b := range issues {
		if b.Parent == "" || !syncingIDs[b.Parent] {
			parents = append(parents, b)
		} else {
			children = append(children, b)
		}
	}

	// Create index mapping for results
	issueIndex := make(map[string]int)
	for i, b := range issues {
		issueIndex[b.ID] = i
	}
	results := make([]SyncResult, len(issues))
	total := len(issues)

	var wg sync.WaitGroup
	var completed int

	reportProgress := func(result SyncResult) {
		if s.opts.OnProgress != nil {
			s.mu.Lock()
			completed++
			current := completed
			s.mu.Unlock()
			s.opts.OnProgress(result, current, total)
		}
	}

	// Pass 1: Create/update parent issues in parallel
	for _, b := range parents {
		wg.Go(func() {
			result := s.syncIssue(ctx, b)
			idx := issueIndex[b.ID]
			results[idx] = result

			if result.Error == nil && result.Action != syncutil.ActionSkipped && result.ExternalID != "" {
				s.mu.Lock()
				var n int
				if _, err := fmt.Sscanf(result.ExternalID, "%d", &n); err == nil {
					s.issueToGHNumber[b.ID] = n
				}
				s.mu.Unlock()
			}
			reportProgress(result)
		})
	}
	wg.Wait()

	// Pass 2: Create/update child issues in parallel (parents now exist)
	for _, b := range children {
		wg.Go(func() {
			result := s.syncIssue(ctx, b)
			idx := issueIndex[b.ID]
			results[idx] = result

			if result.Error == nil && result.Action != syncutil.ActionSkipped && result.ExternalID != "" {
				s.mu.Lock()
				var n int
				if _, err := fmt.Sscanf(result.ExternalID, "%d", &n); err == nil {
					s.issueToGHNumber[b.ID] = n
				}
				s.mu.Unlock()
			}
			reportProgress(result)
		})
	}
	wg.Wait()

	// Pass 3: Update relationship references in issue bodies (if not disabled)
	if !s.opts.NoRelationships && !s.opts.DryRun {
		for _, b := range issues {
			wg.Go(func() {
				if err := s.syncRelationships(ctx, b); err != nil {
					_ = err // Best-effort
				}
			})
		}
		wg.Wait()
	}

	return results, nil
}

// syncIssue syncs a single issue to a GitHub issue.
func (s *Syncer) syncIssue(ctx context.Context, b *issue.Issue) SyncResult {
	result := SyncResult{
		IssueID:    b.ID,
		IssueTitle: b.Title,
	}

	// Upload local images and replace paths with remote URLs
	if !s.opts.DryRun {
		if urlMap, err := UploadImages(ctx, s.client, b.Body); err == nil && len(urlMap) > 0 {
			refs := syncutil.FindLocalImages(b.Body)
			b.Body = syncutil.ReplaceImages(b.Body, refs, urlMap)
			_ = s.core.Update(b, nil)
		}
	}

	// Compute labels, state, and type
	labels := s.computeLabels(b)
	state := s.getGitHubState(b.Status)
	ghType := s.getGitHubType(b.Type)
	body := s.buildIssueBody(b)

	// Check if already linked (from sync store)
	issueNumber := s.syncStore.GetIssueNumber(b.ID)
	if issueNumber != nil && *issueNumber != 0 {
		result.ExternalID = fmt.Sprintf("%d", *issueNumber)

		// Check if issue has changed since last sync
		if !s.opts.Force && !s.needsSync(b) {
			result.Action = syncutil.ActionSkipped
			return result
		}

		// Verify issue still exists
		ghIssue, err := s.client.GetIssue(ctx, *issueNumber)
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
				s.syncStore.Clear(b.ID)
				// Fall through to create new issue
			} else {
				result.Action = syncutil.ActionError
				result.Error = fmt.Errorf("fetching issue #%d: %w", *issueNumber, err)
				return result
			}
		} else {
			// Issue exists - update it
			result.ExternalURL = ghIssue.HTMLURL

			if s.opts.DryRun {
				result.Action = syncutil.ActionWouldUpdate
				return result
			}

			update := s.buildUpdateRequest(ghIssue, b, body, state, ghType, labels)

			if update.hasChanges() {
				updatedIssue, err := s.client.UpdateIssue(ctx, *issueNumber, update)
				if err != nil {
					result.Action = syncutil.ActionError
					result.Error = fmt.Errorf("updating issue: %w", err)
					return result
				}
				result.ExternalURL = updatedIssue.HTMLURL
				result.Action = syncutil.ActionUpdated
			} else {
				result.Action = syncutil.ActionUnchanged
			}

			// Sync sub-issue link (handles add, remove, and re-parent)
			s.syncSubIssueLink(ctx, b, *issueNumber)

			// Update synced_at timestamp
			s.syncStore.SetSyncedAt(b.ID, time.Now().UTC())
			return result
		}
	}

	// Create new issue
	if s.opts.DryRun {
		result.Action = syncutil.ActionWouldCreate
		return result
	}

	createReq := &CreateIssueRequest{
		Title:     b.Title,
		Body:      body,
		Labels:    labels,
		Assignees: s.getAssignees(ctx),
		Type:      ghType,
	}

	ghIssue, err := s.client.CreateIssue(ctx, createReq)
	if err != nil {
		result.Action = syncutil.ActionError
		result.Error = fmt.Errorf("creating issue: %w", err)
		return result
	}

	result.ExternalID = fmt.Sprintf("%d", ghIssue.Number)
	result.ExternalURL = ghIssue.HTMLURL
	s.mu.Lock()
	s.issueToGHNumber[b.ID] = ghIssue.Number
	s.mu.Unlock()

	// Close issue if state should be closed (can't create closed issues directly)
	if state == StateClosed {
		closedState := StateClosed
		_, err := s.client.UpdateIssue(ctx, ghIssue.Number, &UpdateIssueRequest{State: &closedState})
		if err != nil {
			_ = err // Best-effort
		}
	}

	// Link as sub-issue if parent is synced
	s.syncSubIssueLink(ctx, b, ghIssue.Number)

	// Store issue number and sync timestamp
	s.syncStore.SetIssueNumber(b.ID, ghIssue.Number)
	s.syncStore.SetSyncedAt(b.ID, time.Now().UTC())

	result.Action = syncutil.ActionCreated
	return result
}

// needsSync checks if an issue needs to be synced based on timestamps.
func (s *Syncer) needsSync(b *issue.Issue) bool {
	syncedAt := s.syncStore.GetSyncedAt(b.ID)
	if syncedAt == nil {
		return true // Never synced
	}
	if b.UpdatedAt == nil {
		return false // No update time, assume in sync
	}
	return b.UpdatedAt.After(*syncedAt)
}

// buildIssueBody builds the GitHub issue body from a local issue.
// Includes the issue body and a hidden HTML comment with the issue ID.
func (s *Syncer) buildIssueBody(b *issue.Issue) string {
	var parts []string
	if b.Body != "" {
		parts = append(parts, b.Body)
	}
	parts = append(parts, syncutil.SyncFooter)
	parts = append(parts, fmt.Sprintf("<!-- todo:%s -->", b.ID))
	return strings.Join(parts, "\n\n")
}

// getGitHubState maps an issue status to a GitHub issue state.
func (s *Syncer) getGitHubState(issueStatus string) string {
	if state, ok := DefaultStatusMapping[issueStatus]; ok {
		return state
	}
	return StateOpen
}

// getGitHubType maps a local issue type to a GitHub native issue type name.
func (s *Syncer) getGitHubType(issueType string) string {
	if ghType, ok := DefaultTypeMapping[issueType]; ok {
		return ghType
	}
	return ""
}

// computeLabels returns only tag-based labels for an issue.
func (s *Syncer) computeLabels(b *issue.Issue) []string {
	return b.Tags
}

// ensureAllLabels pre-creates all labels that will be needed.
func (s *Syncer) ensureAllLabels(ctx context.Context, issues []*issue.Issue) {
	needed := make(map[string]bool)
	for _, b := range issues {
		for _, label := range s.computeLabels(b) {
			needed[label] = true
		}
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5)

	for label := range needed {
		g.Go(func() error {
			_ = s.client.EnsureLabel(gctx, label, "ededed") // Best-effort
			return nil
		})
	}
	_ = g.Wait()
}

// getAssignees returns the assignee list for issue creation.
func (s *Syncer) getAssignees(ctx context.Context) []string {
	// Assign to token owner
	user, err := s.client.GetAuthenticatedUser(ctx)
	if err != nil {
		return nil
	}
	return []string{user.Login}
}

// buildUpdateRequest builds an UpdateIssueRequest containing only fields that differ from current.
func (s *Syncer) buildUpdateRequest(current *Issue, b *issue.Issue, body, state, ghType string, labels []string) *UpdateIssueRequest {
	update := &UpdateIssueRequest{}

	// Only include title if changed
	if current.Title != b.Title {
		update.Title = &b.Title
	}

	// Only include body if changed
	if current.Body != body {
		update.Body = &body
	}

	// Only include state if changed
	if current.State != state {
		update.State = &state
	}

	// Only include labels if changed
	currentLabels := make([]string, len(current.Labels))
	for i, l := range current.Labels {
		currentLabels[i] = l.Name
	}
	sort.Strings(currentLabels)
	sortedNew := make([]string, len(labels))
	copy(sortedNew, labels)
	sort.Strings(sortedNew)
	if !slices.Equal(currentLabels, sortedNew) {
		update.Labels = labels
	}

	// Only include type if changed
	currentType := ""
	if current.Type != nil {
		currentType = current.Type.Name
	}
	if ghType != "" && currentType != ghType {
		update.Type = &ghType
	}

	return update
}

// syncSubIssueLink ensures the GitHub sub-issue relationship matches the local parent field.
// It handles adding, removing, and re-parenting sub-issues.
func (s *Syncer) syncSubIssueLink(ctx context.Context, b *issue.Issue, ghNumber int) {
	if s.opts.NoRelationships {
		return
	}

	wantParent := ""
	if b.Parent != "" {
		s.mu.RLock()
		if n, ok := s.issueToGHNumber[b.Parent]; ok {
			wantParent = fmt.Sprintf("%d", n)
		}
		s.mu.RUnlock()
	}

	// Check current parent on GitHub
	currentParent, err := s.client.GetParentIssue(ctx, ghNumber)
	if err != nil {
		return // Best-effort
	}

	currentParentNumber := 0
	if currentParent != nil {
		currentParentNumber = currentParent.Number
	}

	wantParentNumber := 0
	if wantParent != "" {
		fmt.Sscanf(wantParent, "%d", &wantParentNumber)
	}

	if currentParentNumber == wantParentNumber {
		return // Already correct
	}

	if wantParentNumber != 0 {
		// Add or re-parent (replace_parent handles both cases)
		_ = s.client.AddSubIssue(ctx, wantParentNumber, ghNumber, true)
	} else if currentParentNumber != 0 {
		// Parent removed locally — unlink
		_ = s.client.RemoveSubIssue(ctx, currentParentNumber, ghNumber)
	}
}

// relationshipPrefixes are the bold-label prefixes used for relationship lines in issue bodies.
var relationshipPrefixes = []string{
	"**Parent:**",
	"**Children:**",
	"**Blocks:**",
	"**Blocked by:**",
}

// syncRelationships updates relationship reference lines in the GitHub issue body.
// Writes Parent, Children, Blocks, and Blocked-by lines as clickable #N links.
// Removes stale lines when relationships are removed locally.
func (s *Syncer) syncRelationships(ctx context.Context, b *issue.Issue) error {
	ghNumber, ok := s.issueToGHNumber[b.ID]
	if !ok {
		return nil
	}

	// Build all relationship lines
	var relLines []string

	// Parent
	if b.Parent != "" {
		if parentNumber, ok := s.issueToGHNumber[b.Parent]; ok {
			relLines = append(relLines, fmt.Sprintf("**Parent:** #%d", parentNumber))
		}
	}

	// Children
	s.mu.RLock()
	childIDs := s.childrenOf[b.ID]
	s.mu.RUnlock()
	if len(childIDs) > 0 {
		var childRefs []string
		for _, childID := range childIDs {
			if childNumber, ok := s.issueToGHNumber[childID]; ok {
				childRefs = append(childRefs, fmt.Sprintf("#%d", childNumber))
			}
		}
		if len(childRefs) > 0 {
			relLines = append(relLines, fmt.Sprintf("**Children:** %s", strings.Join(childRefs, ", ")))
		}
	}

	// Blocks
	if len(b.Blocking) > 0 {
		var blockRefs []string
		for _, blockedID := range b.Blocking {
			if blockedNumber, ok := s.issueToGHNumber[blockedID]; ok {
				blockRefs = append(blockRefs, fmt.Sprintf("#%d", blockedNumber))
			}
		}
		if len(blockRefs) > 0 {
			relLines = append(relLines, fmt.Sprintf("**Blocks:** %s", strings.Join(blockRefs, ", ")))
		}
	}

	// Blocked by
	if len(b.BlockedBy) > 0 {
		var blockedByRefs []string
		for _, blockerID := range b.BlockedBy {
			if blockerNumber, ok := s.issueToGHNumber[blockerID]; ok {
				blockedByRefs = append(blockedByRefs, fmt.Sprintf("#%d", blockerNumber))
			}
		}
		if len(blockedByRefs) > 0 {
			relLines = append(relLines, fmt.Sprintf("**Blocked by:** %s", strings.Join(blockedByRefs, ", ")))
		}
	}

	// Get current issue body
	ghIssue, err := s.client.GetIssue(ctx, ghNumber)
	if err != nil {
		return err
	}

	// Strip all existing relationship lines from the body
	newBody := stripRelationshipLines(ghIssue.Body)

	// Insert new relationship lines before the todo comment
	if len(relLines) > 0 {
		block := strings.Join(relLines, "\n")
		if idx := strings.Index(newBody, "<!-- todo:"); idx >= 0 {
			newBody = newBody[:idx] + block + "\n\n" + newBody[idx:]
		} else {
			newBody = newBody + "\n\n" + block
		}
	}

	if newBody != ghIssue.Body {
		_, err := s.client.UpdateIssue(ctx, ghNumber, &UpdateIssueRequest{Body: &newBody})
		return err
	}

	return nil
}

// stripRelationshipLines removes all relationship-prefixed lines from a body string.
func stripRelationshipLines(body string) string {
	lines := strings.Split(body, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isRel := false
		for _, prefix := range relationshipPrefixes {
			if strings.HasPrefix(trimmed, prefix) {
				isRel = true
				break
			}
		}
		if !isRel {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}


