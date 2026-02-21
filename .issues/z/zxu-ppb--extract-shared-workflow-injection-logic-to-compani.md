---
# zxu-ppb
title: Extract shared workflow injection logic to companion
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:55:32Z
parent: y9c-wny
---

## Description
`brew/workflow.go:93` (`InjectWorkflowJob`) and `zed/workflow.go:38` (`InjectSyncExtensionJob`) follow the same pattern: check for existing job, detect needs, generate YAML, ensure trailing newline, append.

## TODO
- [x] Extract generic `InjectJob(content, jobMarker, needsField, generateFn) (string, error)` into `internal/companion/`
- [x] Refactor brew `InjectWorkflowJob` to use shared helper
- [x] Refactor zed `InjectSyncExtensionJob` to use shared helper
- [x] Run tests
