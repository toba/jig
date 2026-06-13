---
# frc-wf0
title: TUI filter should match partial issue ID
status: completed
type: feature
priority: normal
created_at: 2026-06-13T15:27:54Z
updated_at: 2026-06-13T15:34:18Z
sync:
    github:
        issue_number: "114"
        synced_at: "2026-06-13T15:35:25Z"
---

When typing a filter in the TUI, treat the input as a partial issue ID match in addition to the existing text search. For example, typing `vfj` should surface issue `vfj-jop` as a match.

## Acceptance Criteria

- [x] Filter input matches against issue ID prefix/substring
- [x] Existing title/body text matching continues to work
- [x] Partial ID matches are ranked sensibly alongside text matches



## Summary of Changes

The TUI filter already supports partial issue ID matching. `issueItem.FilterValue()` returns `Title + " " + ID` (internal/todo/tui/list.go:121), and `substringFilter` performs case-insensitive substring matching across the whole value. So typing `vfj` matches an issue with ID `vfj-jop`, in any case, with the suffix alone (`jop`), or across the dash (`fj-j`).

Added `TestPartialIssueIDFilter` in `internal/todo/tui/list_test.go` as a regression test covering: prefix match, full-ID match, case-insensitive, mid-ID substring including the dash, suffix match, and isolating distinct IDs.
