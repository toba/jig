---
# d64-j6r
title: 'Rename project: skill/ja → jig'
status: completed
type: task
priority: normal
created_at: 2026-02-21T16:49:47Z
updated_at: 2026-02-21T16:52:32Z
sync:
    github:
        issue_number: "5"
        synced_at: "2026-02-21T16:53:27Z"
---

Rename Go module github.com/toba/skill → github.com/toba/jig, binary ja → jig, update all references across codebase.

- [x] Phase 1: Go module path rename
- [x] Phase 2: Binary name rename (ja → jig)
- [x] Phase 3: Release workflow and config
- [x] Phase 4: Documentation
- [x] Phase 5: Test files
- [x] Phase 6: Git remote and GitHub repo rename
- [x] Verify: go build, go test, go vet

## Summary of Changes

Renamed project from skill/ja to jig:
- Go module: github.com/toba/skill → github.com/toba/jig (go.mod + all imports)
- Binary: ja → jig (.goreleaser.yaml, cmd/root.go Use field)
- Hook command: ja nope → jig nope, with ja nope added to legacy migration chain
- Release workflow: updated asset names, formula, tap repo references
- Config: .jig.yaml repo and companions updated
- Schema:  and descriptions updated
- Docs: README.md and CLAUDE.md fully updated
- Tests: all assertions updated to expect jig
- GitHub repos renamed: toba/skill → toba/jig, toba/homebrew-ja → toba/homebrew-jig
- Git remote updated to toba/jig
