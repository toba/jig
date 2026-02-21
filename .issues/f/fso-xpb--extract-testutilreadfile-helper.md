---
# fso-xpb
title: Extract testutil.ReadFile helper
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:27Z
updated_at: 2026-02-21T20:57:30Z
parent: y9c-wny
---

## Description
`readFile` test helper (read file, fatal on error) is duplicated in 2 test files.

## TODO
- [x] Create `internal/testutil/readfile.go` with `ReadFile(t *testing.T, path string) string`
- [x] Update `internal/update/update_test.go:263`
- [x] Update `internal/todo/refry/refry_test.go:229`
- [x] Run tests
