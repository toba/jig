---
# kl6-pbl
title: Consolidate commit scripts into ja commit subcommand
status: completed
type: feature
priority: normal
created_at: 2026-02-21T03:40:32Z
updated_at: 2026-02-21T03:42:00Z
sync:
    github:
        issue_number: "2"
        synced_at: "2026-02-21T03:47:20Z"
---

- [x] Create internal/commit/patterns.go with gitignore candidate patterns
- [x] Create internal/commit/commit.go with core logic
- [x] Create internal/commit/commit_test.go with unit tests
- [x] Create cmd/commit.go with cobra command
- [x] Verify build and tests pass


## Summary of Changes

Implemented `ja commit [push]` subcommand that consolidates duplicated commit script logic from ~18 repos. Created `internal/commit/` package with gitignore candidate pattern matching (33 patterns covering Go, Python, Node, Swift/iOS, and secrets), git staging, and background todo sync. Exported `FindKey`/`ReplaceKey` from `internal/config/` to support checking for `todo.sync` config section.
