---
# tre-zyz
title: Remove --markdown changelog output; agents should use --json
status: completed
type: task
priority: normal
created_at: 2026-03-08T04:10:07Z
updated_at: 2026-03-08T04:13:35Z
sync:
    github:
        issue_number: "87"
        synced_at: "2026-03-08T04:13:57Z"
---

The --markdown flag causes agent churn — agents burn tokens agonizing over merging formatted markdown into existing CHANGELOG.md. The --json output already includes github_issue numbers and the repo URL, which is all agents need to construct links themselves.

- [x] Remove --markdown and --changelog-file flags from cmd/changelog.go
- [x] Remove FormatMarkdown, MarkdownOptions, sectionHeader, shortDate, shortDateWithYear, formatEntry, excludeExisting from internal/changelog/changelog.go
- [x] Remove markdown-related tests from changelog_test.go
- [x] No skill files referenced --markdown; nothing to update
- [x] Run tests and lint

## Summary of Changes

Removed `--markdown` and `--changelog-file` flags from `jig changelog`. Agents should use `--json` which already includes `github` repo URL and `github_issue` number per issue — everything needed to construct changelog entries without the merge-two-markdown-files problem.
