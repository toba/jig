---
# wiy-t4z
title: changelog --commits 1 returns empty when commit has no linked completed issues
status: completed
type: bug
priority: normal
created_at: 2026-03-07T21:41:07Z
updated_at: 2026-03-07T21:46:41Z
sync:
    github:
        issue_number: "84"
        synced_at: "2026-03-07T21:47:21Z"
---

## Problem

The commit skill's changelog phase runs `jig changelog --json --commits 1` for `per-commit` changelog config. When the current commit doesn't close/complete any tracked issues, the response has `issues.completed: []`, and the skill skips changelog generation entirely.

This is wrong — the commit clearly has meaningful code changes that belong in the changelog (e.g. a fix to `detect_unused_code` output size). The changelog should reflect what changed in the commit, not just which issues were completed.

## Expected

`jig changelog --commits 1` should produce usable changelog content even when no issues were explicitly completed. Options:

1. Generate entries from the commit message/diff when no issues are completed
2. Add a `changes` or `commits` field to the JSON output so the skill can synthesize an entry
3. Allow the skill to fall back to the commit message as a changelog entry

## Reproduction

```bash
# Make changes, don't complete any issues
jig changelog --json --commits 1
# Returns: { "issues": { "completed": [] } }
# Skill skips changelog entirely despite real changes
```

## Summary of Changes

Two fixes applied:

1. **`CommitTimeRange` zero-width range** (`internal/changelog/changelog.go`): When `--commits 1` is used, `since` and `until` are the same timestamp, creating an empty range where no issues can match. Fixed by extending `until` by 1 second when `since >= until`.

2. **Auto-include git commits** (`cmd/changelog.go`): When `--commits N` is specified, git commits are now automatically included in the output (previously required explicit `--git` flag). This ensures the skill always has commit subject lines to generate changelog entries from, even when no issues were completed.

Added test `TestGather_SingleCommitRange` confirming the fix.
