---
# rjw-i10
title: scoop init looks for wrong architecture asset (arm64 instead of amd64)
status: completed
type: bug
priority: normal
created_at: 2026-03-05T19:39:32Z
updated_at: 2026-03-05T19:44:02Z
sync:
    github:
        issue_number: "75"
        synced_at: "2026-03-05T19:49:30Z"
---

`jig scoop init` tries to download `cupa_windows_arm64.zip` but the release only has `cupa_windows_amd64.zip`.

Scoop is Windows-only and the primary target is amd64. The init command should look for `_windows_amd64.zip` by default.

> Error: resolving SHA256 for cupa_windows_arm64.zip: downloading asset: no assets match the file pattern

## Summary of Changes

Made ARM64 architecture support optional in scoop init:

1. **init.go**: ARM64 SHA256 resolution no longer fails the command — if the arm64 asset is missing, it's silently skipped.
2. **manifest.go**: `GenerateManifest` only includes arm64 entries when `SHA256ARM64` is non-empty.
3. **workflow.go**: CI workflow conditionally includes arm64 architecture based on whether the checksum is found in the release.
