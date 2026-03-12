---
# 77y-3ss
title: File watcher silently drops events and discards errors
status: completed
type: bug
priority: normal
created_at: 2026-03-12T19:03:45Z
updated_at: 2026-03-12T19:11:16Z
sync:
    github:
        issue_number: "92"
        synced_at: "2026-03-12T22:45:36Z"
---

In `internal/todo/core/watcher.go`:

- `watcher.Errors` channel is consumed but discarded with no logging (`_ = err` and a comment saying "you might want to log this")
- `fanOut` default branch silently drops entire event batches when a subscriber's channel (buffered at 16) is full — no log or notification to the subscriber

If the OS drops kernel-level events (e.g. kqueue overflow), the watcher appears healthy while missing real changes.

## Summary of Changes

- `fanOut` now logs a warning with event count and subscriber ID when dropping events for slow subscribers
- `watcher.Errors` now logged via `logWarn` instead of silently discarded
- Added `TestFanOutLogsDroppedEvents` test
