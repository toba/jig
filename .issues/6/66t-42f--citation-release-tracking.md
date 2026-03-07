---
# 66t-42f
title: Citation release tracking
status: completed
type: feature
priority: normal
created_at: 2026-03-07T18:20:23Z
updated_at: 2026-03-07T18:23:46Z
sync:
    github:
        issue_number: "80"
        synced_at: "2026-03-07T18:40:10Z"
---

Track releases instead of commits for cited repos. Add `track: releases` and `last_checked_tag` fields to Source, GitHub release API methods, release-aware review logic, and display changes.

- [x] Add Track, LastCheckedTag fields + helpers to config
- [x] Add config tests for new fields
- [x] Add Release struct to github/types.go
- [x] Add GetLatestRelease, ListReleases to github client
- [x] Add ReleaseInfo to display, update RenderText
- [x] Add checkSourceReleases to cmd/check.go
- [x] Add --releases flag to cite add
- [x] Render track field in FormatSourceYAML
- [x] Add cite/add tests for Track field
- [x] All tests pass, lint clean


## Summary of Changes

Implemented citation release tracking across 9 files. Sources with `track: releases` now compare between GitHub releases instead of branch commits, with full classify/display support and a `--releases` flag on `jig cite add`.
