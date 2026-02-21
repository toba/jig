---
# mj0-cq6
title: Extract shared newJSONRequest to syncutil
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:47:54Z
updated_at: 2026-02-21T21:03:36Z
parent: y9c-wny
---

## Description
`newJSONRequest` is byte-for-byte identical in `clickup/client.go:78` and `github/client.go:79`.

## TODO
- [x] Extract `NewJSONRequest(ctx, method, url, payload) (*http.Request, error)` into `syncutil/http.go`
- [x] Update both clients to use the shared function
- [x] Run tests
