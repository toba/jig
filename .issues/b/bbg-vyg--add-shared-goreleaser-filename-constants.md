---
# bbg-vyg
title: Add shared goreleaser filename constants
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:53Z
updated_at: 2026-02-21T20:57:26Z
parent: y9c-wny
---

## Description
`".goreleaser.yaml"` and `".goreleaser.yml"` are repeated across `brew/language.go:17`, `brew/doctor.go:216-219`, and `zed/doctor.go:219-223`.

## TODO
- [ ] Define `GoreleaserYAML` and `GoreleaserYML` constants in a shared location
- [ ] Update all 3 files to use the constants
- [ ] Run tests
