---
# lgf-bxl
title: Add tests for internal/commit gather and apply phases
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:04Z
updated_at: 2026-02-21T21:05:09Z
parent: y9c-wny
sync:
    github:
        issue_number: "38"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`internal/commit` has only 6.3% test coverage. The core gather and apply phases are untested.

## TODO
- [x] Add tests for the gather phase (file classification, diff generation)
- [x] Add tests for the apply phase (commit creation, message formatting)
- [x] Add tests for error paths
- [x] Target >50% coverage (achieved 76.2%)
