---
# mha-zl8
title: 'fix cite add: skip duplicate when URL already cited'
status: completed
type: bug
priority: normal
created_at: 2026-03-07T17:59:53Z
updated_at: 2026-03-07T17:59:53Z
sync:
    github:
        issue_number: "79"
        synced_at: "2026-03-07T18:00:30Z"
---

When `jig cite add` is called with a URL that already exists in the citations list, it should skip adding a duplicate entry instead of creating one.

- [x] Detect existing URL before adding
- [x] Skip with message when duplicate found

## Summary of Changes

Implemented in commit f3f0103. `cite add` now checks existing citations and skips the add when the URL is already present.
