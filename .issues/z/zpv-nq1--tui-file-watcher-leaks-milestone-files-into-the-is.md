---
# zpv-nq1
title: TUI file watcher leaks milestone files into the issue list
status: completed
type: bug
priority: normal
created_at: 2026-05-27T18:32:18Z
updated_at: 2026-05-27T18:32:18Z
sync:
    github:
        issue_number: "110"
        synced_at: "2026-05-27T18:33:46Z"
---

The file watcher's `handleChanges` and `pollForChanges` walk every `.md` file under `.issues/`, including the `milestones/` subdirectory. When a milestone file is created or rewritten while the TUI is running (e.g. GitHub sync stamping `synced_at`), the watcher loaded it via `loadIssue` and inserted it into `c.issues` — appearing as an empty-title task (default type) with just a due-date hourglass. The non-watching CLI path (`loadFromDisk`) already skips the milestones dir, so this only surfaced in the long-running TUI.

## Summary of Changes

- Added `Core.isMilestonePath` helper (internal/todo/core/watcher.go).
- `handleChanges` now routes milestone-file events to the `c.milestones` map (reload on create/write, delete on remove) and never inserts them into `c.issues`.
- Added `TestHandleChangesIgnoresMilestoneFiles` regression test reproducing the leak.
