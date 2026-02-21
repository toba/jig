---
# ivf-lah
title: Add const defaultHost = "github.com" in cite/add.go
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:40Z
updated_at: 2026-02-21T20:57:26Z
parent: y9c-wny
---

## Description
`"github.com"` appears 3 times in `internal/cite/add.go` (lines 38, 41, 44) without a constant.

## TODO
- [ ] Add `const defaultHost = "github.com"` at top of `internal/cite/add.go`
- [ ] Replace all 3 occurrences
- [ ] Run tests
