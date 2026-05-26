---
# of8-gkt
title: Make deferred status indicator more subtle (pink)
status: completed
type: task
priority: normal
created_at: 2026-05-26T05:12:58Z
updated_at: 2026-05-26T05:13:59Z
sync:
    github:
        issue_number: "107"
        synced_at: "2026-05-26T05:29:06Z"
---

The deferred status icon (⏸) currently uses orange, which is attention-grabbing. Deferred items are parked and shouldn't draw attention. Change the default color to a subtle pink.

## Summary of Changes

- Added a muted dusty pink color (`#D6A2C0`) as `ColorPink` in `internal/todo/ui/styles.go` and registered `pink` in `NamedColors`.
- Changed the default `deferred` status color from `orange` to `pink` in `internal/todo/config/config.go` so the ⏸ indicator is subtle rather than attention-grabbing.
