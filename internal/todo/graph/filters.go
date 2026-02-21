package graph

import (
	"cmp"
	"slices"
	"time"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/graph/model"
	"github.com/toba/jig/internal/todo/integration"
	"github.com/toba/jig/internal/todo/issue"
)

// ApplyFilter applies IssueFilter to a slice of issues and returns filtered results.
// This is used by both the top-level issues query and relationship field resolvers.
func ApplyFilter(issues []*issue.Issue, filter *model.IssueFilter, core *core.Core) []*issue.Issue {
	if filter == nil {
		return issues
	}

	result := issues

	// Status filters
	if len(filter.Status) > 0 {
		result = filterByField(result, filter.Status, func(b *issue.Issue) string { return b.Status })
	}
	if len(filter.ExcludeStatus) > 0 {
		result = excludeByField(result, filter.ExcludeStatus, func(b *issue.Issue) string { return b.Status })
	}

	// Type filters
	if len(filter.Type) > 0 {
		result = filterByField(result, filter.Type, func(b *issue.Issue) string { return b.Type })
	}
	if len(filter.ExcludeType) > 0 {
		result = excludeByField(result, filter.ExcludeType, func(b *issue.Issue) string { return b.Type })
	}

	// Priority filters (empty priority treated as "normal")
	if len(filter.Priority) > 0 {
		result = filterByPriority(result, filter.Priority)
	}
	if len(filter.ExcludePriority) > 0 {
		result = excludeByPriority(result, filter.ExcludePriority)
	}

	// Tag filters
	if len(filter.Tags) > 0 {
		result = filterByTags(result, filter.Tags)
	}
	if len(filter.ExcludeTags) > 0 {
		result = excludeByTags(result, filter.ExcludeTags)
	}

	// Parent filters
	if filter.HasParent != nil && *filter.HasParent {
		result = filterByHasParent(result)
	}
	if filter.NoParent != nil && *filter.NoParent {
		result = filterByNoParent(result)
	}
	if filter.ParentID != nil && *filter.ParentID != "" {
		result = filterByParentID(result, *filter.ParentID)
	}

	// Blocking filters
	if filter.HasBlocking != nil && *filter.HasBlocking {
		result = filterByHasBlocking(result)
	}
	if filter.BlockingID != nil && *filter.BlockingID != "" {
		result = filterByBlockingID(result, *filter.BlockingID)
	}
	if filter.NoBlocking != nil && *filter.NoBlocking {
		result = filterByNoBlocking(result)
	}
	if filter.IsBlocked != nil {
		if *filter.IsBlocked {
			result = filterByIsBlocked(result, core)
		} else {
			result = filterByNotBlocked(result, core)
		}
	}

	// Blocked-by filters (for direct blocked_by field)
	if filter.HasBlockedBy != nil && *filter.HasBlockedBy {
		result = filterByHasBlockedBy(result)
	}
	if filter.BlockedByID != nil && *filter.BlockedByID != "" {
		result = filterByBlockedByID(result, *filter.BlockedByID)
	}
	if filter.NoBlockedBy != nil && *filter.NoBlockedBy {
		result = filterByNoBlockedBy(result)
	}

	// Sync filters
	if filter.HasSync != nil && *filter.HasSync != "" {
		result = filterByHasSync(result, *filter.HasSync)
	}
	if filter.NoSync != nil && *filter.NoSync != "" {
		result = filterByNoSync(result, *filter.NoSync)
	}
	if filter.SyncStale != nil && *filter.SyncStale != "" {
		result = filterBySyncStale(result, *filter.SyncStale)
	}
	if filter.ChangedSince != nil {
		result = filterByChangedSince(result, *filter.ChangedSince)
	}

	return result
}

// stringSet builds a lookup set from a string slice.
func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

// filterIssues returns issues matching the predicate.
func filterIssues(issues []*issue.Issue, pred func(*issue.Issue) bool) []*issue.Issue {
	var result []*issue.Issue
	for _, b := range issues {
		if pred(b) {
			result = append(result, b)
		}
	}
	return result
}

// filterByField filters issues to include only those where getter returns a value in values (OR logic).
func filterByField(issues []*issue.Issue, values []string, getter func(*issue.Issue) string) []*issue.Issue {
	set := stringSet(values)
	return filterIssues(issues, func(b *issue.Issue) bool { return set[getter(b)] })
}

// excludeByField filters issues to exclude those where getter returns a value in values.
func excludeByField(issues []*issue.Issue, values []string, getter func(*issue.Issue) string) []*issue.Issue {
	set := stringSet(values)
	return filterIssues(issues, func(b *issue.Issue) bool { return !set[getter(b)] })
}

// filterByPriority filters issues to include only those with matching priorities (OR logic).
// Empty priority in the issue is treated as "normal" for matching purposes.
func filterByPriority(issues []*issue.Issue, priorities []string) []*issue.Issue {
	set := stringSet(priorities)
	return filterIssues(issues, func(b *issue.Issue) bool { return set[cmp.Or(b.Priority, config.PriorityNormal)] })
}

// excludeByPriority filters issues to exclude those with matching priorities.
// Empty priority in the issue is treated as "normal" for matching purposes.
func excludeByPriority(issues []*issue.Issue, priorities []string) []*issue.Issue {
	set := stringSet(priorities)
	return filterIssues(issues, func(b *issue.Issue) bool { return !set[cmp.Or(b.Priority, config.PriorityNormal)] })
}

// filterByTags filters issues to include only those with any of the given tags (OR logic).
func filterByTags(issues []*issue.Issue, tags []string) []*issue.Issue {
	set := stringSet(tags)
	return filterIssues(issues, func(b *issue.Issue) bool {
		for _, t := range b.Tags {
			if set[t] {
				return true
			}
		}
		return false
	})
}

// excludeByTags filters issues to exclude those with any of the given tags.
func excludeByTags(issues []*issue.Issue, tags []string) []*issue.Issue {
	set := stringSet(tags)
	return filterIssues(issues, func(b *issue.Issue) bool {
		for _, t := range b.Tags {
			if set[t] {
				return false
			}
		}
		return true
	})
}

func filterByHasParent(issues []*issue.Issue) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return b.Parent != "" })
}

func filterByNoParent(issues []*issue.Issue) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return b.Parent == "" })
}

func filterByParentID(issues []*issue.Issue, parentID string) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return b.Parent == parentID })
}

func filterByHasBlocking(issues []*issue.Issue) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return len(b.Blocking) > 0 })
}

func filterByBlockingID(issues []*issue.Issue, targetID string) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return slices.Contains(b.Blocking, targetID) })
}

func filterByNoBlocking(issues []*issue.Issue) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return len(b.Blocking) == 0 })
}

// filterByIsBlocked filters issues that are blocked by active (non-completed, non-scrapped) blockers.
func filterByIsBlocked(issues []*issue.Issue, core *core.Core) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return core.IsBlocked(b.ID) })
}

// filterByNotBlocked filters issues that are NOT blocked by active blockers.
func filterByNotBlocked(issues []*issue.Issue, core *core.Core) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return !core.IsBlocked(b.ID) })
}

func filterByHasBlockedBy(issues []*issue.Issue) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return len(b.BlockedBy) > 0 })
}

func filterByBlockedByID(issues []*issue.Issue, blockerID string) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return slices.Contains(b.BlockedBy, blockerID) })
}

func filterByNoBlockedBy(issues []*issue.Issue) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return len(b.BlockedBy) == 0 })
}

func filterByHasSync(issues []*issue.Issue, name string) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return b.HasSync(name) })
}

func filterByNoSync(issues []*issue.Issue, name string) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return !b.HasSync(name) })
}

// filterBySyncStale filters issues where updatedAt > sync[name]["synced_at"].
// If no synced_at or unparseable, the issue is treated as stale (conservative).
func filterBySyncStale(issues []*issue.Issue, name string) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return isSyncStale(b, name) })
}

// isSyncStale returns true if the issue's updatedAt is after the sync integration's synced_at.
func isSyncStale(b *issue.Issue, name string) bool {
	if b.UpdatedAt == nil {
		return false
	}

	if b.Sync == nil {
		return true
	}
	data, ok := b.Sync[name]
	if !ok {
		return true
	}
	syncedAtRaw, ok := data[integration.SyncKeySyncedAt]
	if !ok {
		return true
	}
	syncedAtStr, ok := syncedAtRaw.(string)
	if !ok {
		return true
	}
	syncedAt, err := time.Parse(time.RFC3339, syncedAtStr)
	if err != nil {
		return true
	}
	return b.UpdatedAt.After(syncedAt)
}

func filterByChangedSince(issues []*issue.Issue, since time.Time) []*issue.Issue {
	return filterIssues(issues, func(b *issue.Issue) bool { return b.UpdatedAt != nil && !b.UpdatedAt.Before(since) })
}
