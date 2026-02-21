package syncutil

import (
	"time"

	"github.com/toba/jig/internal/todo/issue"
)

// SyncTimestampProvider is the minimal interface needed by FilterIssuesNeedingSync.
type SyncTimestampProvider interface {
	GetSyncedAt(issueID string) *time.Time
}

// FilterIssuesNeedingSync returns only issues that need to be synced based on timestamps.
// An issue needs sync if: force is true, it has no sync record, or it was updated after last sync.
func FilterIssuesNeedingSync(issues []*issue.Issue, store SyncTimestampProvider, force bool) []*issue.Issue {
	var needSync []*issue.Issue
	for _, b := range issues {
		if force {
			needSync = append(needSync, b)
			continue
		}
		syncedAt := store.GetSyncedAt(b.ID)
		if syncedAt == nil {
			needSync = append(needSync, b)
			continue
		}
		if b.UpdatedAt != nil && b.UpdatedAt.After(*syncedAt) {
			needSync = append(needSync, b)
		}
	}
	return needSync
}

// GetSyncString retrieves a string value from an issue's sync metadata.
func GetSyncString(b *issue.Issue, syncName, key string) string {
	if b.Sync == nil {
		return ""
	}
	extData, ok := b.Sync[syncName]
	if !ok {
		return ""
	}
	val, ok := extData[key]
	if !ok {
		return ""
	}
	s, _ := val.(string)
	return s
}

// GetSyncTime retrieves a time value from an issue's sync metadata.
func GetSyncTime(b *issue.Issue, syncName, key string) *time.Time {
	s := GetSyncString(b, syncName, key)
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}
