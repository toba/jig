---
# vjq-9e2
title: Add cite doctor subcommand to verify license attribution
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:30:00Z
updated_at: 2026-02-21T19:35:39Z
sync:
    github:
        issue_number: "29"
        synced_at: "2026-02-23T17:08:13Z"
---

## Context

When a project cites external repos via `citations:` in `.jig.yaml`, those sources typically require attribution in `LICENSE`, `NOTICE`, or similar files. Currently there's no automated way to verify that cited repos are actually mentioned in your license/notice files.

The root `doctor` command already aggregates checks from `nope`, `brew`, and `zed` — `cite` should participate too.

## Proposed Behavior

`jig cite doctor` checks that each configured citation has corresponding attribution in the project's license-related files (`LICENSE`, `NOTICE`, `THIRD_PARTY`, etc.).

For each citation in `citations:`:
- [x] Extract the repo name/owner (already available via `config.Source.Repo`)
- [x] Search license-related files for mentions of the repo name, owner, or project name
- [x] For GitHub repos: use `gh api` to fetch the repo's license type (e.g. MIT, Apache-2.0) — could help verify the attribution matches
- [x] For non-GitHub repos (or when `gh` is unavailable): report what was found/not found and suggest the user verify manually, or hand off to an agent for deeper analysis
- [x] Report pass/warn/fail per citation, styled with lipgloss (consistent with other doctor commands)

## Integration

- [x] Add `doctor` subcommand under `cmd/cite.go` parent command
- [x] Register cite doctor in the root `cmd/doctor.go` aggregator so `jig doctor` runs it alongside nope/brew/zed checks
- [x] Follow existing doctor patterns (e.g. `internal/brew/doctor.go`, `internal/zed/doctor.go`)

## Edge Cases

- No `citations:` configured → skip silently or note "no citations configured"
- No LICENSE/NOTICE file found → warn
- Citation repo is a URL, not a GitHub slug → degrade gracefully (skip license API lookup, still search local files)
- Multiple license files → search all of them

## Open Questions

- Should it also check for SPDX identifiers?
- Should it suggest adding missing attributions, or just report?


## Summary of Changes

Implemented `jig cite doctor` subcommand:

- `internal/cite/doctor.go` — core logic: discovers license files (LICENSE, NOTICE, THIRD_PARTY, COPYING, etc.), checks each citation for case-insensitive attribution matches, optionally fetches upstream license type via `gh api repos/{owner}/{repo}/license`
- `internal/cite/doctor_test.go` — 9 tests covering: no sources, no license files, found/missing attribution, partial matches, case insensitivity, NOTICE files, GitHub license enrichment
- `cmd/cite_doctor.go` — Cobra command wiring under `citeCmd`
- `cmd/doctor.go` — registered in root doctor aggregator
- `cmd/cite.go` — added "doctor" to PersistentPreRunE skip list
- `internal/github/client.go` — added `GetLicense` method to Client interface and GHClient
- `internal/github/types.go` — added `LicenseInfo` type
