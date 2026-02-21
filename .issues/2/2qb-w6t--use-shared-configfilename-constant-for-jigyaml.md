---
# 2qb-w6t
title: Use shared ConfigFileName constant for .jig.yaml
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:40Z
updated_at: 2026-02-21T20:57:26Z
parent: y9c-wny
---

## Description
`ConfigFileName = ".jig.yaml"` is already defined in `todo/config/config.go:16` but 3 files use the raw string.

## TODO
- [ ] `cmd/root.go:47` — use config constant or define shared one
- [ ] `internal/nope/config.go:92` — use constant
- [ ] `internal/nope/init.go:116` — use constant
- [ ] Run tests

Note: May need to define the constant in a lower-level package to avoid circular imports.
