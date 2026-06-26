---
# mq4-27x
title: Make ClickUp sync prefetch concurrent and conditional
status: scrapped
type: task
priority: normal
created_at: 2026-06-26T17:22:57Z
updated_at: 2026-06-26T17:28:29Z
sync:
    github:
        issue_number: "122"
        synced_at: "2026-06-26T17:29:00Z"
---

## Problem

ClickUp sync (`internal/todo/integration/clickup/sync.go` `SyncIssues`) fires three unconditional, **sequential** prefetch round-trips before processing any issue, whenever ≥1 issue needs syncing:

- `GetAuthorizedUser` — only needed when CREATING tasks (for assignees)
- `GetList` — only needed to obtain spaceID, which is only used for tags
- `PopulateSpaceTagCache` (`GetSpaceTags`) — only needed if some syncing issue has tag changes

Diagnosed in `/Users/jason/Developer/pacer/core` (2109 issues, 2105 linked). The incremental pre-filter (`FilterIssuesNeedingSync`) works perfectly there: 2109 → 6 issues actually sync. Issue count is NOT the cost driver. The perceived slowdown is this fixed ~1s prefetch tax (3 serial WAN round-trips) paid whenever anything at all changes. `jig todo list` (load+index 2109 issues) is 0.16s; sync dry-run is ~1.08s, of which ~0.9s is network wait.

## Plan

- [ ] Run the prefetch calls concurrently (errgroup) instead of sequentially so ~3 RTT collapses to ~1
- [ ] Make them conditional where safe:
  - [ ] Only fetch authorized user when at least one issue will be created (or keep it but lazy)
  - [ ] Only fetch list + space tags when at least one syncing issue has tags
- [ ] Preserve existing behavior/ordering guarantees (spaceID needed before tag operations)
- [ ] Add/adjust tests

## Goal

A typical few-issue sync drops from ~1s toward a few hundred ms and stays flat regardless of total project size.



## Reasons for Scrapping

Accidental duplicate of tqj-zdd, created when the first `jig todo create` call's JSON parse appeared to fail but the issue was actually written. tqj-zdd carries the real work and summary.
