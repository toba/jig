---
# sfn-zlh
title: Remove relationship field from citations config
status: completed
type: task
priority: normal
created_at: 2026-02-21T19:19:31Z
updated_at: 2026-02-21T19:19:31Z
sync:
    github:
        issue_number: "6"
        synced_at: "2026-02-23T17:08:13Z"
---

The `relationship` field on citation sources (derived, dependency, watch) was purely cosmetic — only displayed in the `jig cite check` output header. Remove it to simplify the config.

- [x] Remove `Relationship` field from `config.Source` struct
- [x] Remove `Relationship` from `citationSource` (migration code)
- [x] Remove `relationship` from `repoEntry` (skill parser)
- [x] Remove `normalizeRelationship` function and its test
- [x] Remove relationship from display header
- [x] Remove from starter config template in `cite init`
- [x] Remove from `schema.json`
- [x] Update all test fixtures and assertions
- [x] Build and tests pass

## Summary of Changes

Removed the `relationship` field from citation sources. It was display-only (shown as e.g. `(derived)` in check output) and didn't affect any logic. Deleted the field from structs, schema, display code, migration parser, and all tests.
