---
# tpq-g0k
title: 'Flatten citations config: remove sources nesting'
status: completed
type: task
priority: normal
created_at: 2026-02-21T19:19:18Z
updated_at: 2026-02-21T19:19:18Z
sync:
    github:
        issue_number: "13"
        synced_at: "2026-02-23T17:08:13Z"
---

Remove the unnecessary `sources:` key nesting in the citations config. The YAML format changes from:

```yaml
citations:
  sources:
    - repo: owner/repo
```

to:

```yaml
citations:
  - repo: owner/repo
```

- [x] Change `Config` from `struct { Sources []Source }` to `type Config []Source`
- [x] Update all `cfg.Sources` references to `*cfg` / `(*cfg)[i]`
- [x] Update migration code (`citationConfig`, `citationSource`)
- [x] Update `schema.json` — citations is now directly an array
- [x] Update all test fixtures and assertions
- [x] Update starter config in `cite init`
- [x] Build and tests pass

## Summary of Changes

Flattened the citations config by removing the intermediate `sources:` key. `Config` is now a named slice type (`type Config []Source`) instead of a struct wrapper. Updated all code, tests, migration logic, and schema.
