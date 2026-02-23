---
# qu1-f0t
title: Rename cite check → cite review; add cite add command
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:25:41Z
updated_at: 2026-02-21T19:28:57Z
sync:
    github:
        issue_number: "23"
        synced_at: "2026-02-23T17:08:13Z"
---

- [x] Rename checkCmd → reviewCmd in cmd/check.go, add check as alias
- [x] Create cmd/cite_add.go with add subcommand
- [x] Create internal/cite/add.go with inspection + path suggestion logic
- [x] Add GetRepo, GetTree methods to internal/github/client.go
- [x] Add RepoInfo, TreeResponse types to internal/github/types.go
- [x] Add AppendSource helper to internal/config/config.go
- [x] Verify: go build, go test, go vet


## Summary of Changes

Renamed `cite check` to `cite review` (with `check` as a backward-compatible alias). Added `cite add <url>` command that inspects a repository via GitHub API or git clone, detects its primary language, and suggests path classification globs. Supports GitHub shorthand, HTTPS, SSH URLs, and non-GitHub git URLs. The `-w` flag appends directly to `.jig.yaml`.
