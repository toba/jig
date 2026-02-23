---
# zht-wbf
title: Extract shared goreleaser existence check
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:55:40Z
parent: y9c-wny
sync:
    github:
        issue_number: "35"
        synced_at: "2026-02-23T17:08:15Z"
---

## Description
`brew/doctor.go:213` and `zed/doctor.go:218` both check for `.goreleaser.yaml`/`.goreleaser.yml` existence.

## TODO
- [x] Extract `CheckGoreleaserExists() ([]byte, bool)` into `internal/companion/`
- [x] Update brew and zed doctor to use the shared function
- [x] Run tests
