---
# sw2-cvl
title: Replace errors.As with errors.AsType generic
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:47:53Z
updated_at: 2026-02-21T20:54:32Z
parent: y9c-wny
---

## Description
Use Go 1.26 `errors.AsType[T]()` to eliminate the manual var declaration.

## TODO
- [x] `cmd/root.go:35` — `var exitErr nope.ExitError; errors.As(err, &exitErr)` → `if exitErr, ok := errors.AsType[nope.ExitError](err); ok {`
