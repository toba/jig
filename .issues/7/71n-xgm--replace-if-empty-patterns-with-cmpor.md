---
# 71n-xgm
title: Replace if-empty patterns with cmp.Or
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:47:53Z
updated_at: 2026-02-21T20:54:31Z
parent: y9c-wny
---

## Description
Replace 6 if-empty-then-default patterns with `cmp.Or` (Go 1.22+).

## TODO
- [x] `cmd/root.go:43` — `configPath()` if/return → `cmp.Or(cfgPath, ".jig.yaml")`
- [x] `cmd/todo_create.go:37` — `if title == "" { title = "Untitled" }` → `title = cmp.Or(title, "Untitled")`
- [x] `cmd/help_all.go:39` — `desc := cmd.Short; if desc == ""` → `desc := cmp.Or(cmd.Short, cmd.Long)`
- [x] `cmd/todo_roadmap.go:328` — color default fallback → `color := cmp.Or(colors[b.Type], "gray")`
- [x] `internal/cite/add.go:65` — branch default → `branch := cmp.Or(info.DefaultBranch, "main")`
- [x] `internal/todo/tui/tui.go:645` — 3-step editor fallback → `editor := cmp.Or(cfg.GetEditor(), os.Getenv("VISUAL"), os.Getenv("EDITOR"))`
