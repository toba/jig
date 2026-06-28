---
# z37-wsf
title: body-file - should read stdin like --body -
status: completed
type: bug
priority: normal
created_at: 2026-06-28T20:25:33Z
updated_at: 2026-06-28T20:25:33Z
sync:
    github:
        issue_number: "123"
        synced_at: "2026-06-28T20:26:22Z"
---

Agents pass `--body-file -` expecting stdin (the Unix convention) but resolveContent treated `-` as a literal filename, failing with `reading file: open -: no such file or directory`. Only `--body/--append-body/--replace-body -` read stdin.

## Summary of Changes

- resolveContent now treats `-` as stdin for the *file* argument too, so `--body-file -`, `--replace-body-file -` pipe stdin.
- Updated flag help on create/update and the agent prompt template to document `-` for the file flags.
- Added TestResolveContent subtests covering stdin for both value and file args.
