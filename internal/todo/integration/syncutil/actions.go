package syncutil

// Sync action constants identify the outcome of syncing a single issue.
const (
	ActionCreated     = "created"
	ActionUpdated     = "updated"
	ActionSkipped     = "skipped"
	ActionError       = "error"
	ActionUnchanged   = "unchanged"
	ActionWouldCreate = "would create"
	ActionWouldUpdate = "would update"
)

// Link/unlink action constants.
const (
	ActionLinked        = "linked"
	ActionAlreadyLinked = "already_linked"
	ActionUnlinked      = "unlinked"
	ActionNotLinked     = "not_linked"
)

// SyncKeySyncedAt is the common key used in sync metadata for the last sync timestamp.
const SyncKeySyncedAt = "synced_at"
