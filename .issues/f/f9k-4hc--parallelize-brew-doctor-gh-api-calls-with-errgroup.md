---
# f9k-4hc
title: Parallelize brew doctor gh API calls with errgroup
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:03Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
sync:
    github:
        issue_number: "60"
        synced_at: "2026-02-23T17:08:16Z"
---

## Description
`brew/doctor.go:43-76` makes 4 sequential `gh` calls. Checks 3, 4, and 5 are independent and can run concurrently. Check 6 depends on check 5.

## TODO
- [x] Run checks 3, 4, 5 concurrently using `errgroup.Group`
- [x] Run check 6 sequentially after check 5 completes
- [x] Collect results and display in original order
- [x] Run tests
