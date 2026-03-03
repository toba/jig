---
# etr-b9r
title: Improve sync configuration discovery for agents
status: completed
type: task
priority: normal
created_at: 2026-03-03T17:37:59Z
updated_at: 2026-03-03T17:44:53Z
sync:
    github:
        issue_number: "71"
        synced_at: "2026-03-03T17:45:38Z"
---

An agent attempting to enable GitHub issue sync struggled with multiple issues:

## Problems Observed

- [x] Agent tried `jig todo sync --enable` (non-existent flag) — no guided setup command exists
- [x] `sync check` says "No integration configured" but doesn't hint at the expected config format or location
- [x] Agent guessed wrong YAML nesting (`todo.sync.github` instead of `sync.github`)
- [x] Agent used `.jig.yml` instead of `.jig.yaml` — no error about wrong filename
- [x] No example config shown in `--help` output or `sync check` error message

## Suggested Improvements

- [x] ~~Add `jig todo sync init`~~ — Improved error messages and help text provide equivalent guidance
- [x] Make `sync check` output include an example config snippet when no integration is found
- [x] Add config file name validation (error when `.jig.yml` found instead of `.jig.yaml`)
- [x] Include a minimal config example in `jig todo sync --help`
- [x] `jig doctor` check warns if `.jig.yml` exists and runs offline sync validation


## Summary of Changes

- **Better error messages**: `sync` and `sync check` now show full YAML config examples (both GitHub and ClickUp) when no integration is configured
- **JSON hint field**: JSON output includes a `hint` field alongside the `error` field
- **Expanded --help**: `jig todo sync --help` shows both GitHub and ClickUp config examples with auth requirements
- **`.jig.yml` detection**: `FindConfig()` returns a clear error if `.jig.yml` exists but `.jig.yaml` is expected
- **Doctor check**: New `sync` entry in `jig doctor` checks for `.jig.yml` typo and runs offline sync validation
