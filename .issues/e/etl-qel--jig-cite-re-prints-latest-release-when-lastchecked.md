---
# etl-qel
title: jig cite re-prints latest release when LastCheckedTag matches release.tag_name
status: completed
type: bug
priority: normal
created_at: 2026-06-13T19:27:27Z
updated_at: 2026-06-13T19:29:27Z
sync:
    github:
        issue_number: "117"
        synced_at: "2026-06-13T19:32:53Z"
---

## Problem

For `track: releases` sources, `jig cite review` re-prints the latest release on every run, even when `LastCheckedTag` already equals `release.tag_name`. The SHA-based "new commits" filter doesn't gate the release output.

## Evidence

Observed today on two sources in JSON output:
- GRDB v7.11.0
- swift-subprocess 0.5

Both showed `LastCheckedTag` exactly matching `release.tag_name` yet the release was emitted again.

## Expected

When `LastCheckedTag == release.tag_name`, the release should be suppressed (or at minimum not surface as a new change) the same way already-seen commits are filtered out.

## Tasks

- [x] Add failing test reproducing repeat release emission when `LastCheckedTag == release.tag_name`
- [x] Gate release output on tag comparison for `track: releases` sources
- [x] Verify `last_checked_*` fields still update correctly on no-op runs



## Summary of Changes

- `cmd/check.go` (`checkSourceReleases`): when `LastCheckedTag == latest.TagName`, stop populating `result.Release`. The headSHA / tag are still returned so `MarkSourceRelease` refreshes `last_checked_sha` and `last_checked_date`, but the JSON `release` field is omitted and the text renderer falls through to the existing "No new releases since <tag>" branch instead of re-emitting a "Tracking releases from …" banner.
- `cmd/check_test.go`: added `TestCheckSourceReleases_NoNewRelease` with a fake `github.Client` that returns a release whose `tag_name` equals `LastCheckedTag`; asserts `result.Release == nil`, no commits, and that headSHA/tag still flow through for the timestamp update.

Verified with `go test ./...`, `go vet ./...`, and `scripts/lint.sh` (0 issues).
