---
# wca-711
title: Changelog skill misses entries from commits that land after the initial changelog update
status: completed
type: bug
priority: high
created_at: 2026-03-07T20:07:36Z
updated_at: 2026-03-07T20:17:47Z
sync:
    github:
        issue_number: "83"
        synced_at: "2026-03-07T20:18:31Z"
---

## Problem

The `/changelog` skill (and per-commit changelog updates in the `/commit` skill) fails to capture entries for commits that land **after** the initial weekly changelog update. In the xc-mcp project this week, 5 significant changes were completely absent from `CHANGELOG.md`:

### Missing entries (all committed after the weekly changelog was generated)

1. **`detect_unused_code` tool** — major new feature wrapping Periphery CLI for dead code detection (commit `5e7cb09`, issue #171)
2. **Subprocess orphan process fix** — `teardownSequence` (SIGTERM → 5s → SIGKILL) preventing cancelled builds from holding SPM lock (same commit)
3. **`validate_project` crash fix** — `PBXBuildFile` Hashable violation, commit `d1a1ef4`
4. **Session defaults PPID isolation** — parallel MCP clients no longer clobber each other, commit `92b4834`
5. **Test plan error hints with scheme awareness** — commit `92b4834`

All of these were committed between Mar 5–7 in the same weekly window (Mar 1–7), but after the initial `48dc12b update changelog for week of Mar 1` commit landed on ~Mar 4.

## Expected behavior

Late-week commits should be appended to the existing weekly changelog section, either:
- Automatically during `/commit` if a changelog section for the current week already exists
- Flagged as missing when `/changelog` is run, with an option to append

## Current behavior

Once the weekly changelog section is written, subsequent commits in the same week are silently ignored. The changelog becomes stale by mid-week.

## Possible fixes

- [ ] `/commit` skill: detect existing weekly section in CHANGELOG.md and append new entries
- [ ] `/changelog` skill: diff committed changes against existing changelog entries to find gaps
- [ ] Add a `jig changelog check` command that compares issue completions against changelog mentions

## Summary of Changes

- `internal/changelog/changelog.go`: Include `review` status alongside `completed` in the `isCompleted` check so review-status issues appear in the changelog's Completed bucket
- `internal/changelog/changelog_test.go`: Added a `review`-status issue test case to `TestGather`, verifying it lands in `Issues.Completed`
