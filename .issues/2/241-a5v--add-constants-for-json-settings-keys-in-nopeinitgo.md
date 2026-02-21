---
# 241-a5v
title: Add constants for JSON settings keys in nope/init.go
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:52Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
---

## Description
`"hooks"`, `"PreToolUse"`, and `"matcher"` are used 13+ times as map keys in `nope/init.go` without named constants.

## TODO
- [x] Add local constants `settingsKeyHooks`, `settingsKeyPreToolUse`, `settingsKeyMatcher` in `nope/init.go`
- [x] Replace all 5 `"hooks"` occurrences
- [x] Replace all 5 `"PreToolUse"` occurrences
- [x] Replace all 3 `"matcher"` occurrences
- [x] Run tests
