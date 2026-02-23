---
# ahd-26y
title: Replace slicesEqual/labelsMatch with slices.Equal
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:27Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
sync:
    github:
        issue_number: "49"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
Two test files have hand-rolled ordered string-slice comparison that `slices.Equal` (Go 1.21+) already provides.

## TODO
- [x] `internal/todo/integration/clickup/sync_test.go:783` — replace `slicesEqual` with `slices.Equal`
- [x] `internal/todo/integration/github/sync_test.go:136` — replace `labelsMatch` with `slices.Equal`
- [x] Remove the local helper functions
- [x] Run tests
