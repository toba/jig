---
# klk-o6u
title: Use issue.DueDateFormat instead of raw date format string
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:52Z
updated_at: 2026-02-21T20:57:26Z
parent: y9c-wny
sync:
    github:
        issue_number: "52"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`cmd/check.go:179` uses `"2006-01-02"` directly instead of the existing `issue.DueDateFormat` constant.

## TODO
- [ ] Replace `"2006-01-02"` at `cmd/check.go:179` with `issue.DueDateFormat`
- [ ] Run tests
