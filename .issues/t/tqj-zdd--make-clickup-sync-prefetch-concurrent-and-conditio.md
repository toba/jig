---
# tqj-zdd
title: Make ClickUp sync prefetch concurrent and conditional
status: completed
type: task
priority: normal
created_at: 2026-06-26T17:23:02Z
updated_at: 2026-06-26T17:27:40Z
sync:
    github:
        issue_number: "121"
        synced_at: "2026-06-26T17:28:32Z"
---

## Problem

ClickUp sync (`internal/todo/integration/clickup/sync.go` `SyncIssues`) fires three unconditional, **sequential** prefetch round-trips before processing any issue, whenever >=1 issue needs syncing:

- `GetAuthorizedUser` — only needed when CREATING tasks (for assignees)
- `GetList` — only needed to obtain spaceID, which is only used for tags
- `PopulateSpaceTagCache` (`GetSpaceTags`) — only needed if some syncing issue has tag changes

Diagnosed in `/Users/jason/Developer/pacer/core` (2109 issues, 2105 linked). The incremental pre-filter (`FilterIssuesNeedingSync`) works perfectly there: 2109 -> 6 issues actually sync. Issue count is NOT the cost driver. The perceived slowdown is this fixed ~1s prefetch tax (3 serial WAN round-trips) paid whenever anything at all changes. `jig todo list` (load+index 2109 issues) is 0.16s; sync dry-run is ~1.08s, of which ~0.9s is network wait.

## Plan

- [x] Run the prefetch calls concurrently (errgroup) instead of sequentially so ~3 RTT collapses to ~1
- [x] Make them conditional where safe (only fetch list+space tags when a syncing issue has tags)
- [x] Preserve ordering guarantees (spaceID needed before tag operations)
- [x] Add/adjust tests

## Goal

A typical few-issue sync drops from ~1s toward a few hundred ms and stays flat regardless of total project size.

## Summary of Changes

`internal/todo/integration/clickup/sync.go` `SyncIssues`: replaced the three unconditional, sequential prefetch round-trips with a conditional + concurrent prefetch.

- Skipped entirely on dry runs (they mutate nothing).
- Authorized-user lookup runs only when a task will be created AND no assignee is configured.
- List + space-tag-cache fetch runs only when a syncing issue carries tags.
- The two remaining prefetches run concurrently via errgroup, collapsing ~3 serial WAN round-trips to ~1. Prefetching the authorized user once also removes a data race on the client's cache across the parallel create passes (confirmed by `-race`).

Diagnosis recap: the incremental pre-filter was already correct (pacer/core: 2109 issues -> 6 actually sync). The perceived slowdown was this fixed prefetch tax, not issue-count scaling. A typical few-issue update sync no longer pays any prefetch; a create/tag sync pays one concurrent round-trip.

Added `TestSyncIssues_ConditionalPrefetch` (4 subtests: update-only/no prefetch, create/user-only, tags/space-cache, dry-run/none). Full integration suite, `-race`, `go vet`, and `scripts/lint.sh` all clean.
