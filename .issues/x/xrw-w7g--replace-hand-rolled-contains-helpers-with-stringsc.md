---
# xrw-w7g
title: Replace hand-rolled contains helpers with strings.Contains
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
sync:
    github:
        issue_number: "56"
        synced_at: "2026-02-23T17:08:16Z"
---

## Description
Six test files have hand-rolled `contains(s, substr string) bool` helpers that duplicate `strings.Contains`.

## TODO
- [x] `internal/cite/add_test.go:136` — replace with `strings.Contains`
- [x] `internal/config/companions_test.go:149` — replace with `strings.Contains`
- [x] `internal/todo/refry/refry_test.go:271` — replace with `strings.Contains`
- [x] `internal/brew/formula_test.go:151` — replace with `strings.Contains`
- [x] `internal/todo/integration/clickup_adapter_test.go:177` — replace with `strings.Contains`
- [x] `internal/todo/integration/syncutil/images_test.go:234` — replace with `strings.Contains`
- [x] Remove the local helper functions
- [x] Run tests
