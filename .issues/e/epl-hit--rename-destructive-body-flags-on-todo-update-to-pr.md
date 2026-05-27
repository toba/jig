---
# epl-hit
title: Rename destructive body flags on todo update to prevent accidental body clobbering
status: completed
type: task
priority: normal
created_at: 2026-05-27T19:15:08Z
updated_at: 2026-05-27T19:17:45Z
sync:
    github:
        issue_number: "113"
        synced_at: "2026-05-27T19:18:56Z"
---

Agents repeatedly wipe an issue's existing body by reaching for `jig todo update --body/--body-file`, which does a full replacement.

## Behavior
- Add `--replace-body` / `--replace-body-file` for the explicit destructive whole-body write.
- Add `--append-body` as the documented safe add (keep `--body-append` working as a hidden alias).
- Remove `--body`/`--body-file`/`-d` from `update`: if passed, error with guidance pointing to `--append-body`, `--body-replace-old/new`, or `--replace-body`.
- `create` keeps `--body`/`--body-file` (nothing to clobber).
- Substring flags `--body-replace-old/new` unchanged.
- Update prompt template and help text.

## Tasks
- [x] Failing test for new flags + guidance error
- [x] Implement flag rename + guidance
- [x] Update cmd/todo_prompt.tmpl
- [x] Lint, test, vet

## Summary of Changes

- Added `--replace-body`/`--replace-body-file` (destructive whole-body overwrite) and `--append-body` to `jig todo update`.
- Retired `--body`/`--body-file`/`-d` on `update`: hidden, and passing them returns a guidance error pointing to `--append-body`, `--body-replace-old/new`, or `--replace-body`. `create` keeps `--body`.
- Kept `--body-append` as a hidden working alias; substring flags `--body-replace-old/new` unchanged.
- Refactored flag registration into `registerUpdateFlags` for isolated testing; added `TestBuildUpdateInputBody`.
- Updated `cmd/todo_prompt.tmpl` (agent guide) to teach the new verbs and warn that `update` has no `--body`.
