---
# wzp-ixx
title: jig commit leaves dirty files after sync metadata updates
status: completed
type: bug
priority: high
created_at: 2026-02-26T00:27:20Z
updated_at: 2026-02-26T00:47:20Z
sync:
    github:
        issue_number: "70"
        synced_at: "2026-02-26T00:47:51Z"
---

## Problem

`jig commit gather` syncs issues to GitHub (updating metadata like `synced_at`, `issue_number`), then stages everything. `jig commit apply` commits the staged files, then syncs again — but the post-commit sync metadata changes are NOT committed, leaving dirty `.issues/` files.

## Expected Behavior

After `jig commit apply [--push]`, the working tree should be clean. The sync-commit-push cycle should not leave leftover modified issue files.

## Current Workaround

Manual `git add .issues/ && git commit -m "sync issue metadata after GitHub push"` after every commit.

## Suggested Fix

Either:
1. After the post-commit sync, auto-commit the metadata changes (e.g. `git add .issues/ && git commit --amend` or a second commit)
2. Or restructure to: sync → stage all → commit → push → sync → stage → commit (two commits)
3. Or only sync once (during gather), not again during apply

## TODO

- [x] Fix the commit workflow so no dirty files remain after apply


## Summary of Changes

Moved todo sync to happen **before** staging/committing instead of after, so sync metadata changes are naturally included in the commit:

- **gather**: moved `syncTodoIfConfigured` before `StageAll()` so metadata updates from sync are picked up by `git add -A`
- **apply**: sync runs before commit, then `RestageIssues()` re-stages any `.issues/` changes, so they're included in the commit. Removed the post-commit and post-push syncs that were the source of the dirty files.
- **`RestageIssues()`**: new helper in `internal/commit/` that stages `.issues/` if the directory exists, handling the case where sync runs between gather and apply.
