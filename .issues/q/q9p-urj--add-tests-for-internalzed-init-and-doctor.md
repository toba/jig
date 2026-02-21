---
# q9p-urj
title: Add tests for internal/zed init and doctor
status: ready
type: task
priority: normal
created_at: 2026-02-21T20:49:14Z
updated_at: 2026-02-21T20:49:50Z
parent: y9c-wny
---

## Description
`internal/zed` has only 17.0% test coverage. Init and doctor logic need tests.

## TODO
- [ ] Add tests for zed init (extension.toml generation, Cargo.toml scaffolding)
- [ ] Add tests for zed doctor checks (extension exists, required files present)
- [ ] Add tests for workflow injection
- [ ] Target >40% coverage
