---
# pix-e52
title: Use existing integration.Action* constants instead of raw strings
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:40Z
updated_at: 2026-02-21T20:58:57Z
parent: y9c-wny
---

## Description
The `integration` package defines `ActionCreated`, `ActionUpdated`, `ActionSkipped`, `ActionError`, `ActionUnchanged`, `ActionWouldCreate`, and `ActionWouldUpdate` constants. But clickup/sync.go (11 occurrences) and github/sync.go (11 occurrences) use raw string literals instead.

## TODO
- [x] Replace all 11 raw action strings in `clickup/sync.go` with `integration.Action*` constants
- [x] Replace all 11 raw action strings in `github/sync.go` with `integration.Action*` constants
- [x] Run tests
