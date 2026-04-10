---
# lx3-9p9
title: Add cite update command for modifying existing citations
status: completed
type: feature
priority: normal
created_at: 2026-04-10T22:03:02Z
updated_at: 2026-04-10T22:08:08Z
sync:
    github:
        issue_number: "101"
        synced_at: "2026-04-10T22:10:03Z"
---

Add `jig cite update <source>` command to modify fields of an existing citation in `.jig.yaml`.

## Updatable fields
- `--branch` — change tracked branch
- `--track` — change tracking mode (e.g. "releases" or empty to clear)
- `--scope` — change scope description
- `--notes` — change notes
- `--paths-high` — replace high-priority path globs
- `--paths-medium` — replace medium-priority path globs
- `--paths-low` — replace low-priority path globs
- `--repo` — change repo identifier (rename)

## Tasks
- [x] Write failing test for `UpdateSource` in `internal/config/`
- [x] Implement `UpdateSource` in `internal/config/config.go`
- [x] Create `cmd/cite_update.go` with Cobra command
- [x] Write test for cite update command
- [x] Update CLAUDE.md architecture docs
- [x] Run lint and tests


## Summary of Changes

Added `jig cite update <source>` command that modifies fields of an existing citation in `.jig.yaml`.

### Files changed
- `cmd/cite_update.go` — new Cobra command with flags for all updatable fields
- `cmd/cite.go` — skip PersistentPreRunE config load for `update` subcommand
- `cmd/cmd_test.go` — subcommand registration, flags, and integration tests
- `internal/config/config_test.go` — round-trip update and rename tests
- `CLAUDE.md` — architecture docs updated
