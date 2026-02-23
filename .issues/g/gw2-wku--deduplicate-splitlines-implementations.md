---
# gw2-wku
title: Deduplicate splitLines implementations
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:48:12Z
updated_at: 2026-02-21T20:49:49Z
parent: y9c-wny
sync:
    github:
        issue_number: "54"
        synced_at: "2026-02-23T17:08:16Z"
---

## Description
`splitLines` is duplicated in 3 files with slightly different implementations. Use `strings.Split` or extract a shared helper.

## TODO
- [x] `update/update.go:202` — already uses `strings.Split`, keep as-is
- [x] `nope/init.go:165` — manual byte-loop → replace with `strings.Split(s, "\n")`
- [x] `cmd/init.go:70` — manual byte-loop → replace with `strings.Split(s, "\n")`
- [x] Run tests
