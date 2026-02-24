---
# 7t5-opd
title: Add 'review' status for code-complete issues awaiting evaluation
status: completed
type: feature
priority: normal
created_at: 2026-02-24T01:09:01Z
updated_at: 2026-02-24T01:14:58Z
---

Add a new `review` status that indicates an issue is code-complete but needs evaluation (code review, testing, QA, etc.) before it can be marked completed.

## Motivation

Currently the only path from `in-progress` is straight to `completed` or `scrapped`. There's no way to signal "work is done, but it needs someone to look at it." This is useful for:
- Code review gates
- QA/testing phases
- Acceptance evaluation
- Any "done but not done-done" state

## Implementation

Files to modify:

- [ ] `internal/todo/config/config.go` — add `StatusReview = "review"` constant and entry in `DefaultStatuses` (between `in-progress` and `draft` in sort order, or after `in-progress`)
- [ ] `internal/todo/ui/styles.go` — add icon (e.g. "◉" or "⊙") and color (e.g. cyan or magenta) for `review` status
- [ ] `internal/todo/core/links.go` — no changes needed (`review` is not a resolved status, so it still blocks dependents — correct behavior)
- [ ] `internal/todo/integration/github/config.go` — add `"review": "open"` to `DefaultStatusMapping`
- [ ] `internal/todo/integration/clickup/config.go` — add `"review": "in review"` (or similar) to `DefaultStatusMapping`
- [ ] `internal/todo/graph/schema.graphqls` — update status comment to include `review`
- [ ] `CLAUDE.md` — add `review` to the Statuses section in the agent prompt
- [ ] `cmd/prime.go` — update status documentation in agent instructions

## Design Decisions

- `review` should NOT be an archive status (issues in review are still active)
- `review` should sort after `in-progress` but before `draft` (active work flows: in-progress → review → completed)
- `review` issues should still block dependents (not resolved)
- The `--ready` filter should exclude `review` (it's not available for pickup)

## Summary of Changes

Added `review` status (code complete, awaiting evaluation) between `in-progress` and `ready` in the status workflow.

- Added `StatusReview` constant and `DefaultStatuses` entry with cyan color
- Added `◈` icon for review status, rendered with in-progress (active work) styling
- Excluded review from `--ready` filter (review issues aren't available for pickup)
- Added GitHub sync mapping (`review → open`) and ClickUp mapping (`review → review`)
- Updated GraphQL schema status docstring
- Updated all tests for new status count (5 → 6)
