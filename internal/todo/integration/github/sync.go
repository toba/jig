package github

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strconv"
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
	issueToGHID     map[string]int      // local issue ID -> GitHub issue ID (for sub-issues API)
	childrenOf      map[string][]string // parent ID -> child IDs (built during sync)

	// Milestone tracking
	issueToMilestoneNumber map[string]int    // local issue ID -> GitHub milestone number
	issueTypes             map[string]string // local issue ID -> issue type
}

// NewSyncer creates a new syncer with the given client and options.
func NewSyncer(client *Client, cfg *Config, opts SyncOptions, c *core.Core, syncStore SyncStateProvider) *Syncer {
	return &Syncer{
		client:                 client,
		config:                 cfg,
		opts:                   opts,
		core:                   c,
		syncStore:              syncStore,
		issueToGHNumber:        make(map[string]int),
		issueToGHID:            make(map[string]int),
		childrenOf:             make(map[string][]string),
		issueToMilestoneNumber: make(map[string]int),
		issueTypes:             make(map[string]string),
	}
}

// SyncIssues syncs a list of issues to GitHub issues.
// Uses a multi-pass approach:
// 0. Create/update milestones for milestone-type issues
// 1. Create/update parent issues (issues without parents, or parents not in this sync)
// 2. Create/update child issues with sub-issue relationships
// 3. Sync native blocking/blocked-by relationships via GitHub dependencies API
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

	// Build type index for parent lookups
	for _, b := range issues {
		s.issueTypes[b.ID] = b.Type
	}

	// Pre-populate mapping with already-synced issues from sync store
	for _, b := range issues {
		issueNumber := s.syncStore.GetIssueNumber(b.ID)
		if issueNumber != nil && *issueNumber != 0 {
			s.issueToGHNumber[b.ID] = *issueNumber
		}
		milestoneNumber := s.syncStore.GetMilestoneNumber(b.ID)
		if milestoneNumber != nil && *milestoneNumber != 0 {
			s.issueToMilestoneNumber[b.ID] = *milestoneNumber
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

	// Separate milestone-type issues from regular issues
	var milestones []*issue.Issue
	var regularIssues []*issue.Issue
	for _, b := range issues {
		if b.Type == "milestone" {
			milestones = append(milestones, b)
		} else {
			regularIssues = append(regularIssues, b)
		}
	}

	// Separate regular issues into layers: parents first, then children
	var parents, children []*issue.Issue
	for _, b := range regularIssues {
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

	var completed int
	g := new(errgroup.Group)
	g.SetLimit(10)

	reportProgress := func(result SyncResult) {
		if s.opts.OnProgress != nil {
			s.mu.Lock()
			completed++
			current := completed
			s.mu.Unlock()
			s.opts.OnProgress(result, current, total)
		}
	}

	syncAndTrack := func(b *issue.Issue, result SyncResult) {
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
	}

	// Pass 0: Create/update milestones
	for _, b := range milestones {
		result := s.syncMilestone(ctx, b)
		idx := issueIndex[b.ID]
		results[idx] = result
		reportProgress(result)
	}

	// Pass 1: Create/update parent issues in parallel
	for _, b := range parents {
		g.Go(func() error {
			syncAndTrack(b, s.syncIssue(ctx, b))
			return nil
		})
	}
	_ = g.Wait()

	// Pass 2: Create/update child issues in parallel (parents now exist)
	for _, b := range children {
		g.Go(func() error {
			syncAndTrack(b, s.syncIssue(ctx, b))
			return nil
		})
	}
	_ = g.Wait()

	// Pass 3: Sync native blocking relationships (if not disabled)
	if !s.opts.NoRelationships && !s.opts.DryRun {
		for _, b := range regularIssues {
			if len(b.Blocking) == 0 && len(b.BlockedBy) == 0 {
				continue
			}
			g.Go(func() error {
				s.syncBlockingRelationships(ctx, b)
				return nil
			})
		}
		_ = g.Wait()
	}

	return results, nil
}

// syncMilestone syncs a milestone-type issue to a GitHub milestone.
func (s *Syncer) syncMilestone(ctx context.Context, b *issue.Issue) SyncResult {
	result := SyncResult{
		IssueID:    b.ID,
		IssueTitle: b.Title,
	}

	state := s.getGitHubState(b.Status)
	// GitHub milestones use "open"/"closed" just like issues
	milestoneState := state

	var dueOn string
	if b.Due != nil {
		dueOn = b.Due.Format(time.RFC3339)
	}

	// Check if already linked
	milestoneNumber := s.syncStore.GetMilestoneNumber(b.ID)
	if milestoneNumber != nil && *milestoneNumber != 0 {
		result.ExternalID = fmt.Sprintf("milestone:%d", *milestoneNumber)

		if !s.opts.Force && !s.needsSync(b) {
			result.Action = syncutil.ActionSkipped
			return result
		}

		// Verify milestone still exists
		ghMilestone, err := s.client.GetMilestone(ctx, *milestoneNumber)
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
				s.syncStore.Clear(b.ID)
				// Fall through to create
			} else {
				result.Action = syncutil.ActionError
				result.Error = fmt.Errorf("fetching milestone #%d: %w", *milestoneNumber, err)
				return result
			}
		} else {
			result.ExternalURL = ghMilestone.HTMLURL
			s.issueToMilestoneNumber[b.ID] = ghMilestone.Number

			if s.opts.DryRun {
				result.Action = syncutil.ActionWouldUpdate
				return result
			}

			// Build update request
			update := &UpdateMilestoneRequest{}
			if ghMilestone.Title != b.Title {
				update.Title = &b.Title
			}
			body := s.buildMilestoneDescription(b)
			if ghMilestone.Description != body {
				update.Description = &body
			}
			if ghMilestone.State != milestoneState {
				update.State = &milestoneState
			}
			if dueOn != "" && ghMilestone.DueOn != dueOn {
				update.DueOn = &dueOn
			}

			if update.hasChanges() {
				updated, err := s.client.UpdateMilestone(ctx, *milestoneNumber, update)
				if err != nil {
					result.Action = syncutil.ActionError
					result.Error = fmt.Errorf("updating milestone: %w", err)
					return result
				}
				result.ExternalURL = updated.HTMLURL
				result.Action = syncutil.ActionUpdated
			} else {
				result.Action = syncutil.ActionUnchanged
			}

			s.syncStore.SetSyncedAt(b.ID, time.Now().UTC())
			return result
		}
	}

	// Create new milestone
	if s.opts.DryRun {
		result.Action = syncutil.ActionWouldCreate
		return result
	}

	createReq := &CreateMilestoneRequest{
		Title:       b.Title,
		Description: s.buildMilestoneDescription(b),
		State:       milestoneState,
		DueOn:       dueOn,
	}

	ghMilestone, err := s.client.CreateMilestone(ctx, createReq)
	if err != nil {
		result.Action = syncutil.ActionError
		result.Error = fmt.Errorf("creating milestone: %w", err)
		return result
	}

	result.ExternalID = fmt.Sprintf("milestone:%d", ghMilestone.Number)
	result.ExternalURL = ghMilestone.HTMLURL
	s.issueToMilestoneNumber[b.ID] = ghMilestone.Number

	s.syncStore.SetMilestoneNumber(b.ID, ghMilestone.Number)
	s.syncStore.SetSyncedAt(b.ID, time.Now().UTC())

	result.Action = syncutil.ActionCreated
	return result
}

// buildMilestoneDescription builds the GitHub milestone description from a local issue.
func (s *Syncer) buildMilestoneDescription(b *issue.Issue) string {
	var parts []string
	if b.Body != "" {
		parts = append(parts, b.Body)
	}
	parts = append(parts, fmt.Sprintf(TodoCommentFormat, b.ID))
	return strings.Join(parts, "\n\n")
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
	milestoneNumber := s.getMilestoneForIssue(b)

	// Check if already linked (from sync store)
	issueNumber := s.syncStore.GetIssueNumber(b.ID)
	if issueNumber != nil && *issueNumber != 0 {
		result.ExternalID = strconv.Itoa(*issueNumber)

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

			// Track GitHub issue ID for sub-issues API
			s.mu.Lock()
			s.issueToGHID[b.ID] = ghIssue.ID
			s.mu.Unlock()

			if s.opts.DryRun {
				result.Action = syncutil.ActionWouldUpdate
				return result
			}

			update := s.buildUpdateRequest(ghIssue, b, body, state, ghType, labels, milestoneNumber)

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
		Milestone: milestoneNumber,
	}

	ghIssue, err := s.client.CreateIssue(ctx, createReq)
	if err != nil {
		result.Action = syncutil.ActionError
		result.Error = fmt.Errorf("creating issue: %w", err)
		return result
	}

	result.ExternalID = strconv.Itoa(ghIssue.Number)
	result.ExternalURL = ghIssue.HTMLURL
	s.mu.Lock()
	s.issueToGHNumber[b.ID] = ghIssue.Number
	s.issueToGHID[b.ID] = ghIssue.ID
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
	parts = append(parts, syncutil.SyncFooter, fmt.Sprintf(TodoCommentFormat, b.ID))
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

// getMilestoneForIssue returns the GitHub milestone number if the issue's parent is a milestone-type issue.
func (s *Syncer) getMilestoneForIssue(b *issue.Issue) *int {
	if b.Parent == "" {
		return nil
	}
	// Check if parent is a milestone-type issue
	if parentType, ok := s.issueTypes[b.Parent]; ok && parentType == "milestone" {
		if milestoneNum, ok := s.issueToMilestoneNumber[b.Parent]; ok {
			return &milestoneNum
		}
	}
	return nil
}

// relationshipPrefixes are the bold-label prefixes used for legacy relationship lines in issue bodies.
var relationshipPrefixes = []string{
	"**Parent:**",
	"**Children:**",
	"**Blocks:**",
	"**Blocked by:**",
}

// buildUpdateRequest builds an UpdateIssueRequest containing only fields that differ from current.
func (s *Syncer) buildUpdateRequest(current *Issue, b *issue.Issue, body, state, ghType string, labels []string, milestone *int) *UpdateIssueRequest {
	update := &UpdateIssueRequest{}

	// Only include title if changed
	if current.Title != b.Title {
		update.Title = &b.Title
	}

	// Strip stale relationship lines from current body before comparing
	currentBody := stripRelationshipLines(current.Body)
	if currentBody != body {
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

	// Only include milestone if changed
	currentMilestone := 0
	if current.Milestone != nil {
		currentMilestone = current.Milestone.Number
	}
	wantMilestone := 0
	if milestone != nil {
		wantMilestone = *milestone
	}
	if currentMilestone != wantMilestone {
		update.Milestone = nullableInt(wantMilestone) // 0 serializes as null to clear
	}

	return update
}

// syncSubIssueLink ensures the GitHub sub-issue relationship matches the local parent field.
// It handles adding, removing, and re-parenting sub-issues.
// Skips if parent is a milestone-type issue (those use milestone assignment instead).
// The GitHub sub-issues API requires issue IDs (not numbers) for the sub_issue_id field.
func (s *Syncer) syncSubIssueLink(ctx context.Context, b *issue.Issue, ghNumber int) {
	if s.opts.NoRelationships {
		return
	}

	// Look up desired parent's GitHub number
	wantParentNumber := 0
	if b.Parent != "" {
		// Skip sub-issue linking if parent is a milestone-type (use milestone assignment instead)
		if parentType, ok := s.issueTypes[b.Parent]; ok && parentType == "milestone" {
			return
		}
		s.mu.RLock()
		if n, ok := s.issueToGHNumber[b.Parent]; ok {
			wantParentNumber = n
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

	if currentParentNumber == wantParentNumber {
		return // Already correct
	}

	if wantParentNumber != 0 {
		// The sub-issues API requires the child's issue ID (not number)
		s.mu.RLock()
		childID, hasChildID := s.issueToGHID[b.ID]
		s.mu.RUnlock()
		if !hasChildID {
			return // Can't link without the child's GitHub ID
		}
		_ = s.client.AddSubIssue(ctx, wantParentNumber, childID, true)
	} else if currentParentNumber != 0 {
		// Parent removed locally — unlink (also needs issue ID)
		s.mu.RLock()
		childID, hasChildID := s.issueToGHID[b.ID]
		s.mu.RUnlock()
		if !hasChildID {
			return
		}
		_ = s.client.RemoveSubIssue(ctx, currentParentNumber, childID)
	}
}

// syncBlockingRelationships syncs native blocking/blocked-by relationships for an issue.
// It diffs the current GitHub state against the desired state and adds/removes as needed.
func (s *Syncer) syncBlockingRelationships(ctx context.Context, b *issue.Issue) {
	s.mu.RLock()
	ghNumber, ok := s.issueToGHNumber[b.ID]
	s.mu.RUnlock()
	if !ok {
		return
	}

	// Build desired blocked-by set from both BlockedBy and inverse of Blocking
	// For this issue: it is blocked by items in b.BlockedBy
	// We sync blocked-by on *this* issue for b.BlockedBy entries
	wantBlockedBy := make(map[int]bool)
	for _, blockerID := range b.BlockedBy {
		s.mu.RLock()
		if blockerGHID, ok := s.issueToGHID[blockerID]; ok {
			wantBlockedBy[blockerGHID] = true
		}
		s.mu.RUnlock()
	}

	// For b.Blocking entries: this issue blocks those, so we add blocked-by on the *other* issues
	for _, blockedID := range b.Blocking {
		s.mu.RLock()
		blockedGHNumber, hasNumber := s.issueToGHNumber[blockedID]
		thisGHID, hasThisID := s.issueToGHID[b.ID]
		s.mu.RUnlock()
		if !hasNumber || !hasThisID {
			continue
		}
		// Add this issue as a blocker on the blocked issue
		currentBlockedBy, err := s.client.ListBlockedBy(ctx, blockedGHNumber)
		if err != nil {
			continue
		}
		alreadyBlocked := false
		for _, dep := range currentBlockedBy {
			if dep.ID == thisGHID {
				alreadyBlocked = true
				break
			}
		}
		if !alreadyBlocked {
			_ = s.client.AddBlockedBy(ctx, blockedGHNumber, thisGHID)
		}
	}

	// Sync blocked-by on this issue
	currentBlockedBy, err := s.client.ListBlockedBy(ctx, ghNumber)
	if err != nil {
		return
	}

	// Build current set
	currentSet := make(map[int]bool)
	for _, dep := range currentBlockedBy {
		currentSet[dep.ID] = true
	}

	// Add missing
	for ghID := range wantBlockedBy {
		if !currentSet[ghID] {
			_ = s.client.AddBlockedBy(ctx, ghNumber, ghID)
		}
	}

	// Remove stale
	for _, dep := range currentBlockedBy {
		if !wantBlockedBy[dep.ID] {
			_ = s.client.RemoveBlockedBy(ctx, ghNumber, dep.ID)
		}
	}
}

// stripRelationshipLines removes all legacy relationship-prefixed lines from a body string.
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
