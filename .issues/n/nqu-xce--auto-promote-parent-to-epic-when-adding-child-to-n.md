---
# nqu-xce
title: Auto-promote parent to epic when adding child to non-container type
status: completed
type: feature
priority: normal
created_at: 2026-03-07T18:59:29Z
updated_at: 2026-03-07T19:01:53Z
sync:
    github:
        issue_number: "82"
        synced_at: "2026-03-07T19:15:27Z"
---

When user sets a parent on an issue, and the parent's type (e.g. task, bug, feature) doesn't allow children of that type, automatically promote the parent to epic instead of returning an error.

- [x] Write failing test
- [x] Modify ValidateParent to auto-promote parent to epic
- [x] Ensure parent is saved after promotion
- [x] Run tests and lint

## Summary of Changes

Modified `ValidateParent` in `internal/todo/core/links.go` to auto-promote a parent issue to epic when its current type doesn't allow children of the requested type. Promotion only happens for types below epic in the hierarchy (feature, task, bug) — milestones are never demoted. Updated the existing GraphQL resolver test that expected an error for task-parenting-task to instead verify the auto-promotion behavior.
