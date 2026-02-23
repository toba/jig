---
# lva-bnf
title: Parallelize zed doctor gh API calls with errgroup
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:53Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
sync:
    github:
        issue_number: "58"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`zed/doctor.go:41-83` makes 5 sequential `gh api` calls for independent checks. Running them concurrently would cut doctor wall time.

## TODO
- [x] Run checks 2-6 concurrently using `errgroup.Group`
- [x] Collect results and display in original order
- [x] Run tests
