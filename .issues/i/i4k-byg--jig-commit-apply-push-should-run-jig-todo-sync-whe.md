---
# i4k-byg
title: jig commit apply --push should run jig todo sync when sync is configured
status: completed
type: bug
priority: high
created_at: 2026-02-23T05:40:03Z
updated_at: 2026-02-23T16:14:27Z
---

When `jig commit apply --push` pushes to the remote, it should automatically run `jig todo sync` if `.jig.yaml` has a `sync:` section configured. Currently issues only sync when the user manually runs `jig todo sync`, which means the external tracker (GitHub Issues, etc.) gets out of date after every push.

## Expected behavior

After a successful push in `jig commit apply --push`, check if sync is configured and if so, run `jig todo sync` automatically.

## Current behavior

`jig commit apply --push` commits, tags, and pushes but does not sync issues. The user must remember to run `jig todo sync` separately.

## Reproduction

1. Configure `.jig.yaml` with `sync: github: repo: ...`
2. Create or update issues with `jig todo`
3. Run `jig commit apply -m "msg" --push`
4. Check GitHub — issues are not synced


## Summary of Changes

Fixed `TodoSync()` in `internal/commit/commit.go` — was calling `exec.Command("todo", "sync")` (nonexistent binary) instead of `exec.Command("jig", "todo", "sync")`. The sync hookup in `cmd/commit.go` was already correct; only the executable name was wrong.
