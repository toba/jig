---
# 9ri-4h7
title: Rename .toba.yaml to .jig.yaml
status: completed
type: task
priority: normal
created_at: 2026-02-21T16:55:39Z
updated_at: 2026-02-21T16:57:06Z
sync:
    github:
        issue_number: "18"
        synced_at: "2026-02-23T17:08:13Z"
---

- [ ] Find all references to .toba.yaml in code
- [ ] Update all references to .jig.yaml
- [ ] Ensure tests pass

## Summary of Changes

Renamed all references from `.toba.yaml` to `.jig.yaml` across the entire codebase:
- 28 source files updated (Go, tests, markdown, JSON, issue files)
- `tobaPath` variable renamed to `jigPath` in 6 Go files
- `schema.json` title and description updated
- Actual config file renamed from `.toba.yaml` to `.jig.yaml`
- All tests pass, build and vet clean
