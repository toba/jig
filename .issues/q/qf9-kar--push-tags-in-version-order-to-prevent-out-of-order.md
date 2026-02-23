---
# qf9-kar
title: Push tags in version order to prevent out-of-order GitHub releases
status: completed
type: bug
priority: high
created_at: 2026-02-23T00:39:52Z
updated_at: 2026-02-23T00:43:52Z
sync:
    github:
        issue_number: "17"
        synced_at: "2026-02-23T17:08:13Z"
---

## Problem

When `jig commit apply -v <version> --push` pushes, it sends all unpushed tags to the remote at once via `git push`. If an older tag existed locally but was never pushed (e.g. v1.10.3 from a previous session without `--push`), it gets pushed simultaneously with the new tag (e.g. v1.11.0). GitHub's release automation then creates both releases near-simultaneously, and the older version can end up with a *higher* release number than the newer version.

Observed: GitHub release #41 = v1.11.0, release #42 = v1.10.3. The later release has a lower version number.

## Root Cause

`git push` (or `git push --follow-tags`) sends all reachable tags in one operation. GitHub Actions triggers are non-deterministic in ordering when multiple tags arrive in one push.

## Proposed Fix

Before pushing a new tag, `jig commit apply` should:

1. Detect any local tags not present on the remote (`git tag -l` vs `git ls-remote --tags`)
2. If unpushed tags exist that are *older* than the new tag being created:
   - Push them individually first, in semver order (`git push origin <tag>`)
   - Or at minimum, warn the user that stale unpushed tags exist
3. Then push the new commit and tag

This ensures GitHub releases are created in version order.

## Reproduction

```bash
# Session 1: commit with tag but don't push
jig commit apply -m "fix something" -v v1.10.3

# Session 2: commit with tag and push
jig commit apply -m "add feature" -v v1.11.0 --push
# Both v1.10.3 and v1.11.0 tags are pushed simultaneously
# GitHub may create v1.10.3 release AFTER v1.11.0 release
```



## Summary of Changes

Modified `Push()` in `internal/commit/commit.go` to push unpushed version tags individually in semver order instead of using `git push --tags` (which sends all tags at once).

**New behavior:**
1. `git push` pushes the branch first
2. `unpushedVersionTags()` compares local `v*` tags against `git ls-remote --tags origin` to find unpushed ones
3. Each unpushed tag is pushed individually via `git push origin <tag>`, in version-sorted order (oldest first)
4. If the remote is unavailable for listing, falls back to `git push --tags`

This ensures GitHub Actions release triggers fire in version order, preventing older releases from getting higher release numbers.

**Files changed:**
- `internal/commit/commit.go` — replaced `Push()`, added `unpushedVersionTags()`
- `internal/commit/commit_test.go` — added `TestUnpushedVersionTags` (4 sub-tests), `TestPushOrdersTags`, and `setupGitRepoWithRemote` helper
