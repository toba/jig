---
# 7yn-mof
title: Consolidate truncate implementations into display.Truncate
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:47:54Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
---

## Description
Three separate `truncate`/`truncateTitle` implementations exist. Consolidate into one exported `display.Truncate`.

## TODO
- [x] Create `display.Truncate(s string, maxLen int) string` in `display/display.go`
- [x] Update `cmd/todo_list.go:247` to call `display.Truncate`
- [x] Update `cmd/todo_sync.go:161` to call `display.Truncate`
- [x] Remove the duplicate local functions
- [x] Run tests
