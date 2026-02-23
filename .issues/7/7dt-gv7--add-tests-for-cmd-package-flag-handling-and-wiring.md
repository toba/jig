---
# 7dt-gv7
title: Add tests for cmd package flag handling and wiring
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:18Z
updated_at: 2026-02-21T21:05:17Z
parent: y9c-wny
sync:
    github:
        issue_number: "48"
        synced_at: "2026-02-23T17:08:16Z"
---

## Description
`cmd` has only 20.6% test coverage. Command wiring and flag handling need more tests.

## TODO
- [x] Add tests for flag parsing and validation across key commands
- [x] Add tests for command error paths
- [x] Add tests for output formatting
- [x] Target >35% coverage (achieved 35.5%)
