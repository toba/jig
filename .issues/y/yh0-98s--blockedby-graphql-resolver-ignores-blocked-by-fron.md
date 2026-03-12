---
# yh0-98s
title: blockedBy GraphQL resolver ignores blocked_by frontmatter links
status: completed
type: bug
priority: normal
created_at: 2026-03-12T19:03:45Z
updated_at: 2026-03-12T19:09:18Z
sync:
    github:
        issue_number: "94"
        synced_at: "2026-03-12T22:45:36Z"
---

The `BlockedBy` resolver in `internal/todo/graph/schema.resolvers.go` only checks `LinkTypeBlocking` and drops `LinkTypeBlockedBy` links. If issue B has `blocked_by: [A]`, querying A's `blockedBy` field won't include B.

The fix is to include both link types in the filter:
```go
if link.LinkType == issue.LinkTypeBlocking || link.LinkType == issue.LinkTypeBlockedBy {
```

Note: `BlockedByIds` resolver is fine — it returns `obj.BlockedBy` directly.

## Summary of Changes

Fixed `BlockedBy` resolver in `internal/todo/graph/schema.resolvers.go` to combine both sources: issues that declare `blocking: [obj]` and issues listed in `obj.blocked_by`. Added deduplication via `seen` map. Added test case for the `blocked_by` frontmatter path.
