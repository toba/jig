---
# 96d-51b
title: Extract testutil.Chdir helper
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:57:30Z
parent: y9c-wny
---

## Description
The chdir save/restore pattern (`origDir := os.Getwd(); os.Chdir(dir); defer os.Chdir(origDir)`) is duplicated across 10+ test files.

## TODO
- [x] Create `internal/testutil/chdir.go` with `Chdir(t *testing.T, dir string)` using `t.Cleanup`
- [x] Update `internal/cite/doctor_test.go` (9 occurrences)
- [x] Update `internal/nope/doctor_test.go` (3 occurrences)
- [x] Update `internal/nope/init_test.go` (2 occurrences)
- [x] Update `internal/update/update_test.go`
- [x] Update `internal/update/cite_test.go`
- [x] Update `internal/update/commit_test.go`
- [x] Update `internal/brew/language_test.go`
- [x] Run tests
