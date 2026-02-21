---
# pt4-51i
title: 'brew doctor: detect project language and adjust diagnostics'
status: completed
type: feature
priority: normal
created_at: 2026-02-21T16:36:54Z
updated_at: 2026-02-21T16:38:11Z
sync:
    github:
        issue_number: "3"
        synced_at: "2026-02-21T16:40:37Z"
---

- [x] Create internal/brew/language.go with Language type, DetectLanguage, and per-language methods
- [x] Modify internal/brew/doctor.go to use Language for checks 2, 6, 7, 9, 12
- [x] Create internal/brew/language_test.go
- [x] Verify build and tests pass


## Summary of Changes

Added language detection to `brew doctor` so it adjusts diagnostics for Go, Swift, and Rust projects instead of hardcoding Go/goreleaser assumptions.
