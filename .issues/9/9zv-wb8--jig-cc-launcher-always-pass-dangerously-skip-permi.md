---
# 9zv-wb8
title: 'jig cc launcher: always pass --dangerously-skip-permissions to claude'
status: completed
type: task
priority: normal
created_at: 2026-04-29T02:23:50Z
updated_at: 2026-04-29T02:24:15Z
sync:
    github:
        issue_number: "104"
        synced_at: "2026-04-29T02:30:59Z"
---

The `jig cc` launcher (`internal/cc/run.go::Launch`) should always inject `--dangerously-skip-permissions` when exec'ing the resolved CLI, so the user no longer has to pass it manually on every invocation. This also applies to `jig cc login`.

## Tasks
- [x] Prepend `--dangerously-skip-permissions` in `internal/cc/run.go::Launch` (idempotent — skip if already present)
- [x] Run `go vet ./...` and `scripts/lint.sh`

## Plan
See /Users/jason/.jig/cc/backup/plans/update-jig-cc-launcher-glittery-piglet.md



## Summary of Changes

- `internal/cc/run.go`: `Launch` now prepends `--dangerously-skip-permissions` to the args passed to the resolved CLI, using `slices.Contains` to keep it idempotent. Applies to both `jig cc <alias>` and `jig cc login`.
- Verified with `go vet`, `go build`, and `scripts/lint.sh` (0 issues).
