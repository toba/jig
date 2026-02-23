---
# q12-koq
title: Add tests for todo TUI key handlers and navigation
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:04Z
updated_at: 2026-02-21T21:01:15Z
parent: y9c-wny
sync:
    github:
        issue_number: "43"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`internal/todo/tui` has only 5.8% test coverage. Key handlers, filtering, and navigation logic are untested.

## TODO
- [x] Add tests for key event handling (navigation, selection, filtering)
- [x] Add tests for view state transitions
- [x] Add tests for filter/search logic
- [x] Target >30% coverage
