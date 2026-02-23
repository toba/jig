---
# pk6-su5
title: Extract shared transient-error helpers to syncutil
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:47:53Z
updated_at: 2026-02-21T21:03:36Z
parent: y9c-wny
sync:
    github:
        issue_number: "44"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`isTransientNetworkError` and `isTransientHTTPError` are duplicated between clickup and github integration clients. Extract to `syncutil/retry.go`.

## TODO
- [x] Extract `isTransientNetworkError` from `clickup/client.go:490` and `github/client.go:397` into `syncutil/retry.go`
- [x] Extract `isTransientHTTPError` from `clickup/client.go:515` and `github/client.go:418` into `syncutil/retry.go` (ClickUp version is superset)
- [x] Update both clients to call the shared functions
- [x] Run tests
