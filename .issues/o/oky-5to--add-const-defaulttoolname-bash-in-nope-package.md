---
# oky-5to
title: Add const DefaultToolName = "Bash" in nope package
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:27Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
sync:
    github:
        issue_number: "36"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
The string `"Bash"` is used 5 times across `internal/nope/` without a named constant.

## TODO
- [x] Add `const DefaultToolName = "Bash"` in `internal/nope/config.go`
- [x] Replace `guard.go:78` — `return "Bash"` → `return DefaultToolName`
- [x] Replace `guard.go:87` — `hi.ToolName = "Bash"` → `hi.ToolName = DefaultToolName`
- [x] Replace `config.go:127` — `name == "Bash"` → `name == DefaultToolName`
- [x] Replace `config.go:160` — `t != "Bash"` → `t != DefaultToolName`
- [x] Replace `init.go:259` — `m["matcher"] != "Bash"` → `m["matcher"] != DefaultToolName`
- [x] Run tests
