---
# kid-fit
title: Block parent completion while children are active
status: completed
type: feature
priority: normal
created_at: 2026-06-25T03:00:25Z
updated_at: 2026-06-25T03:03:40Z
sync:
    github:
        issue_number: "119"
        synced_at: "2026-06-25T03:06:33Z"
---

Prevent a parent issue from being moved into a complete status (completed, scrapped, deferred) unless all of its child issues are also in a complete status.

## Tasks
- [x] Add config helper to identify complete statuses (completed, scrapped, deferred)
- [x] Add Core.Children public accessor
- [x] Enforce rule in resolver UpdateIssue (covers CLI + GraphQL)
- [x] Failing test first, then implementation

## Summary of Changes

A parent issue can no longer enter a complete status (completed, scrapped, deferred) while any child is still in an active (non-complete) status.

- `config.IsCompleteStatus(name)` — central definition of the complete set (completed/scrapped/deferred).
- `core.Core.Children(parentID)` — public, lock-safe accessor for an issue's children.
- `resolver.validateParentCompletion` runs in `UpdateIssue` before mutating the issue; only guards transitions *into* a complete status (no-op for non-status edits on already-complete parents). Covers CLI, GraphQL, and TUI (all route through the resolver).
- Error names the blocking children, e.g. `cannot set s71-t8d to "completed": 1 child issue(s) not complete: qxu-fun (ready)`.
- New table-driven test `TestMutationUpdateIssueParentCompletionGuard` (written failing first).
