---
# f8y-nd4
title: brew/scoop init should not fail when companion repo already exists
status: completed
type: bug
priority: normal
created_at: 2026-03-05T19:39:32Z
updated_at: 2026-03-05T19:41:10Z
sync:
    github:
        issue_number: "77"
        synced_at: "2026-03-05T19:49:30Z"
---

`jig brew init` and `jig scoop init` fail with:

> Error: creating tap repo: GraphQL: Name already exists on this account (createRepository)

When the companion repo already exists (e.g. created manually or by a prior partial init), the init command should detect it exists and skip creation, proceeding to push the formula/manifest.

## Reproduction

1. Create companion repos manually (e.g. `homebrew-cupa`, `scoop-cupa`)
2. Run `jig brew init` → fails trying to create repo that already exists

## Summary of Changes

Added `companion.RepoExists()` helper that checks via `gh repo view` whether a repo already exists. Updated `createTapRepo` (brew), `createBucketRepo` (scoop), and `createExtRepo` (zed) to skip `gh repo create` when the companion repo already exists, allowing init to proceed to push content instead of failing.
