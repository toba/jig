---
# w8v-nyx
title: TUI sort by updated should bubble parent date from descendants
status: scrapped
type: bug
priority: normal
created_at: 2026-05-28T15:07:26Z
updated_at: 2026-05-28T15:07:26Z
sync:
    github:
        issue_number: "116"
        synced_at: "2026-06-13T15:35:25Z"
---

TUI sort-by-updated already treats parents as effectively updated by the newest of self or any descendant (`internal/todo/issue/sort.go:178 ComputeEffectiveDates`, applied at both root and child levels via `ui.BuildTree` -> `SortByEffectiveDate` in `internal/todo/tui/list.go:351-360`). Recursion covers grandchildren and deeper (`TestComputeEffectiveDates/propagates_through_grandchildren`).

Reported symptom in ../thesis: epic `thesis-9ahh` sorted far down despite recent child updates. Investigation showed all 13 issues in the subtree (epic + 9 children + 3 grandchildren) share `updated_at: 2026-05-27T17:50:20Z` exactly. Effective date therefore correctly equals 17:50:20Z; nothing newer exists to bubble. Code is behaving as specified.

## Reasons for Scrapping

The requested behavior already exists and is correct. Root cause of the user-visible symptom is upstream of the sort:

1. No code path in jig cascades a field assignment (e.g. milestone) from a parent to its descendants. Verified across `Mutation.UpdateIssue`, `inheritMilestoneFromParent`, `MigrateMilestoneTypeIssues`, `propagateStatusLocked`, `RemoveLinksTo`/`FixBrokenLinks`, `SaveSyncOnly`, GitHub `Syncer.syncIssue`, watcher, and TUI `milestoneSelectedMsg`. The 13 identical timestamps must have come from individual updates batched fast enough to land in the same second-truncation bucket.
2. `core.go:333,373` and `propagate.go:134` truncate `time.Now()` to `time.Second`. Rapid sequential updates collapse to the same `updated_at`; `SortByEffectiveDate` then breaks ties by title with no recency signal. This is the real ordering hazard, separable from the parent-bubbling behavior the original report asked about.

Follow-up work (to be filed as separate issues if desired):
- Increase `updated_at` precision past seconds so rapid bulk updates retain ordering.
- Add first-class subtree assignment (e.g. `--cascade` on `update --milestone`, or `jig todo milestone assign <ms> <root> --recursive`) so agents don't need to walk the tree manually.
