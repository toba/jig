---
# u2n-k3p
title: '`todo init` rewrites .jig.yaml and drops fields it does not manage; lost sync.github.repo setting'
status: completed
type: bug
priority: normal
created_at: 2026-03-21T17:26:50Z
updated_at: 2026-03-21T17:40:44Z
sync:
    github:
        issue_number: "98"
        synced_at: "2026-03-21T18:02:31Z"
---

## Reproduction

Running `jig todo init` in a project with existing `.jig.yaml` that has `todo.sync.github.repo` causes the sync config to be dropped, because `todo init` creates a `Default()` config (no sync/tags) and `Save()` replaces the entire `todo:` YAML node.

## Tasks

- [x] Write failing test
- [x] Fix `todo init` to preserve existing config fields
- [x] Run tests and lint


## Summary of Changes

The bug was in `cmd/todo_init.go`: `todo init` created a `Default()` config (with only path, default_status, default_type) and called `Save()`, which replaced the entire `todo:` YAML node — dropping fields like `sync.github.repo`, `tags`, `editor`, etc.

**Fix**: Changed `todo init` to call `LoadFromDirectory()` first, which loads the existing config (preserving all fields including sync), then saves it back. If no config exists, `LoadFromDirectory` returns defaults as before.

**Files changed**:
- `cmd/todo_init.go` — use `LoadFromDirectory` instead of `Default()`
- `internal/todo/config/config_test.go` — added `TestInitPreservesExistingConfig` regression test
