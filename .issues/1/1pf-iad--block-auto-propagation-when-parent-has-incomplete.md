---
# 1pf-iad
title: Block auto-propagation when parent has incomplete checklist
status: completed
type: feature
priority: normal
created_at: 2026-04-12T23:29:40Z
updated_at: 2026-04-12T23:31:13Z
sync:
    github:
        issue_number: "102"
        synced_at: "2026-04-12T23:31:48Z"
---

When all child issues are completed/scrapped, the parent is auto-updated to match. This is wrong when the parent has its own incomplete checklist items.

- [x] Add `HasIncompleteChecklist` helper to `internal/todo/issue/content.go`
- [x] Add early return in `computeParentStatus` in `internal/todo/core/propagate.go`
- [x] Add tests for `HasIncompleteChecklist` in `content_test.go`
- [x] Add propagation tests in `propagate_test.go`
- [x] Run tests and lint


## Summary of Changes

Added `HasIncompleteChecklist()` to `internal/todo/issue/content.go` and an early return in `computeParentStatus()` in `internal/todo/core/propagate.go`. Parent issues with any `- [ ]` items in their body are now excluded from automatic status propagation. 19 new test cases added across unit and integration tests.
