---
# ao2-rrk
title: Replace ptrEqual variants with generic ptrEqual[T comparable]
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:27Z
updated_at: 2026-02-21T20:54:27Z
parent: y9c-wny
---

## Description
`stringPtrEqual`, `intPtrEqual`, and `int64PtrEqual` in `clickup/sync.go` are structurally identical — classic generics candidate.

## TODO
- [x] Add `func ptrEqual[T comparable](a, b *T) bool` in `clickup/sync.go`
- [x] Replace `stringPtrEqual` (line 525) with `ptrEqual`
- [x] Replace `intPtrEqual` (line 536) with `ptrEqual`
- [x] Replace `int64PtrEqual` (line 772) with `ptrEqual`
- [x] Remove the three type-specific functions
- [x] Run tests
