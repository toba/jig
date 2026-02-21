---
# ms5-8lj
title: Add const DefaultBranch = "main"
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:40Z
updated_at: 2026-02-21T20:57:26Z
parent: y9c-wny
---

## Description
`"main"` is used as a default branch name in 3 places without a constant.

## TODO
- [ ] Define `const DefaultBranch = "main"` in an appropriate shared location
- [ ] Replace `internal/cite/add.go:67`
- [ ] Replace `internal/cite/add.go:105`
- [ ] Replace `internal/config/config.go:74`
- [ ] Run tests
