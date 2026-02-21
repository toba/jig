---
# ax1-kx0
title: Implement skill brew update command
status: in-progress
type: feature
priority: normal
created_at: 2026-02-21T02:54:41Z
updated_at: 2026-02-21T02:54:41Z
sync:
    github:
        issue_number: "1"
        synced_at: "2026-02-21T03:40:28Z"
---

Implement `skill brew update` CLI command to automate Homebrew tap formula updates after a release.

## Tasks
- [x] Create `internal/brew/formula.go` — Parse and rewrite `.rb` formula files
- [x] Create `internal/brew/sha.go` — SHA extraction (3 strategies: sidecar, checksums.txt, compute)
- [x] Create `internal/brew/update.go` — Orchestration: clone tap, update formula, commit+push
- [x] Create `cmd/brew.go` — Cobra parent command
- [x] Create `cmd/brew_update.go` — `skill brew update` subcommand with flags
- [x] Add `brew` to PersistentPreRunE skip list in cmd/root.go
- [x] Write tests for formula parsing, SHA extraction, and URL building
- [x] Verify build, tests, and vet pass
- [x] Dry-run against real tap
- [ ] Update brew SKILL.md (user-level skill, separate from this repo)
