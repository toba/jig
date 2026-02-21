---
# ohk-cer
title: Extract parseIntMapping in clickup config
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:54:15Z
parent: y9c-wny
---

## Description
`clickup/config.go:100-115` and `clickup/config.go:118-134` have two identical blocks parsing `map[string]any` into `map[string]int`.

## TODO
- [x] Extract `parseIntMapping(m map[string]any) map[string]int` local helper
- [x] Replace both blocks with calls to the helper
- [x] Run tests
