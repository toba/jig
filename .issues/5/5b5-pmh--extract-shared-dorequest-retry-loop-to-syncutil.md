---
# 5b5-pmh
title: Extract shared doRequest retry loop to syncutil
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:47:54Z
updated_at: 2026-02-21T21:03:36Z
parent: y9c-wny
sync:
    github:
        issue_number: "53"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`doRequest` with retry/exponential-backoff logic is nearly identical in `clickup/client.go:398` and `github/client.go:293`. Only auth header injection and rate-limit handling differ.

## TODO
- [x] Design shared `DoWithRetry(httpClient, req, retryConfig, authFunc, isTransient) ([]byte, error)` in `syncutil/`
- [x] Refactor clickup client to use shared retry loop
- [x] Refactor github client to use shared retry loop
- [x] Run tests
