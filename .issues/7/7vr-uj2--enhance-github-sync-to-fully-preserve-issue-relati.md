---
# 7vr-uj2
title: Enhance GitHub sync to fully preserve issue relationships
status: completed
type: feature
priority: normal
created_at: 2026-02-22T17:46:06Z
updated_at: 2026-02-22T17:57:26Z
sync:
    github:
        issue_number: "20"
        synced_at: "2026-02-23T17:08:13Z"
---

Enhance the GitHub issue sync to fully preserve parent/child and blocking/blocked-by relationships using both the sub-issues API and footer links.

## Current State

The sync already has a 3-pass approach (`internal/todo/integration/github/sync.go`):
- **Pass 1**: Sync parent issues (no parent, or parent not in batch)
- **Pass 2**: Sync child issues, call `AddSubIssue` API to link parent↔child
- **Pass 3**: Append `**Blocks:** #123, #456` lines to issue bodies

## Gaps

- [x] Sub-issue linking only happens on **creation** — if a parent is added/changed after initial sync, the sub-issue relationship is never updated on GitHub
- [x] Sub-issue **unlinking** never happens — removing a parent locally doesn't remove the sub-issue link on GitHub
- [x] `blocked_by` relations are **not synced** — only the `blocking` direction is written to GitHub bodies
- [x] No **parent footer link** — parent relationship relies solely on the sub-issues API; a visible `**Parent:** #42` line in the body would make it scannable
- [x] No **blocked-by footer links** — a `**Blocked by:** #10, #11` line would complement the existing `**Blocks:**` line
- [x] When a blocking/blocked-by target **wasn't synced** (no GitHub number), the reference is silently dropped with no indication

## Proposed Changes

### 1. Update sub-issue links on every sync (not just creation)
In `syncIssue()`, after updating an existing issue, check if the parent relationship matches the current sub-issue state on GitHub and call `AddSubIssue`/`RemoveSubIssue` as needed.

### 2. Add `RemoveSubIssue` to the GitHub client
GitHub sub-issues API supports `DELETE /repos/{owner}/{repo}/issues/{parent}/sub_issues` to unlink a child.

### 3. Expand `syncRelationships()` to handle all relation types
Currently only writes `**Blocks:**` lines. Extend to also write:
- `**Blocked by:** #10, #11`
- `**Parent:** #42`
- `**Children:** #43, #44`

These footer links (`#N` format) are natively clickable in GitHub and provide a quick-glance view of the issue graph even without the sub-issues UI.

### 4. Clean up stale relationship lines on update
When relationships change locally (e.g., a blocking link is removed), the corresponding line in the GitHub issue body should be removed on next sync.

### 5. Handle partially-synced relationship targets gracefully
When a relationship target hasn't been synced to GitHub, either:
- Skip the reference (current behavior) but log a warning, or
- Include the local issue title as plain text: `**Blocks:** #123, *Unsynced: "Add widget support"*`

## Files to Modify

- `internal/todo/integration/github/sync.go` — `syncIssue()`, `syncRelationships()`
- `internal/todo/integration/github/client.go` — add `RemoveSubIssue()`, possibly `ListSubIssues()`
- `internal/todo/integration/github/sync_test.go` — cover new relationship sync paths
- `internal/todo/integration/github/client_test.go` — cover new client methods

## References

- GitHub sub-issues API: `POST /repos/{owner}/{repo}/issues/{parent}/sub_issues` (add), `DELETE` (remove)
- GitHub sub-issues GA announcement: https://github.com/orgs/community/discussions/154148
- Sub-issues support up to 50 children per parent, 8 levels of nesting

## Summary of Changes

### client.go
- Updated `AddSubIssue` to accept `replaceParent` bool for re-parenting
- Added `RemoveSubIssue` (DELETE /repos/{owner}/{repo}/issues/{issue_number}/sub_issue)
- Added `GetParentIssue` (GET /repos/{owner}/{repo}/issues/{issue_number}/parent)
- Added `RemoveSubIssueRequest` type

### sync.go
- Added `syncSubIssueLink()` — manages sub-issue API links on both create and update, handles add/remove/re-parent
- Added `childrenOf` map to Syncer, built during SyncIssues for children index
- Rewrote `syncRelationships()` to write all 4 relation types as footer links: **Parent**, **Children**, **Blocks**, **Blocked by**
- Added `stripRelationshipLines()` to clean stale relationship lines before writing new ones
- Relationship lines are inserted before the `<!-- todo:id -->` comment using `#N` format for native GitHub linking

### sync_test.go
- `TestStripRelationshipLines` — 4 cases covering strip/preserve behavior
- `TestSyncRelationships_AllTypes` — verifies all 4 relation types appear in body
- `TestSyncRelationships_CleansStaleLines` — verifies old lines removed when relations cleared
- `TestSyncSubIssueLink_AddParent` — verifies AddSubIssue called for new parent
- `TestSyncSubIssueLink_RemoveParent` — verifies RemoveSubIssue called when parent removed
- `TestSyncSubIssueLink_NoRelationships` — verifies NoRelationships flag skips API calls
