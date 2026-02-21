---
# zht-wbf
title: Extract shared goreleaser existence check
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:55:40Z
parent: y9c-wny
---

## Description
`brew/doctor.go:213` and `zed/doctor.go:218` both check for `.goreleaser.yaml`/`.goreleaser.yml` existence.

## TODO
- [x] Extract `CheckGoreleaserExists() ([]byte, bool)` into `internal/companion/`
- [x] Update brew and zed doctor to use the shared function
- [x] Run tests
