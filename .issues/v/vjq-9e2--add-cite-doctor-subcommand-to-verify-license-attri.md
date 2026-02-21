---
# vjq-9e2
title: Add cite doctor subcommand to verify license attribution
status: in-progress
type: feature
priority: normal
created_at: 2026-02-21T19:30:00Z
updated_at: 2026-02-21T19:31:04Z
---

## Context

When a project cites external repos via `citations:` in `.jig.yaml`, those sources typically require attribution in `LICENSE`, `NOTICE`, or similar files. Currently there's no automated way to verify that cited repos are actually mentioned in your license/notice files.

The root `doctor` command already aggregates checks from `nope`, `brew`, and `zed` — `cite` should participate too.

## Proposed Behavior

`jig cite doctor` checks that each configured citation has corresponding attribution in the project's license-related files (`LICENSE`, `NOTICE`, `THIRD_PARTY`, etc.).

For each citation in `citations:`:
- [ ] Extract the repo name/owner (already available via `config.Source.Repo`)
- [ ] Search license-related files for mentions of the repo name, owner, or project name
- [ ] For GitHub repos: use `gh api` to fetch the repo's license type (e.g. MIT, Apache-2.0) — could help verify the attribution matches
- [ ] For non-GitHub repos (or when `gh` is unavailable): report what was found/not found and suggest the user verify manually, or hand off to an agent for deeper analysis
- [ ] Report pass/warn/fail per citation, styled with lipgloss (consistent with other doctor commands)

## Integration

- [ ] Add `doctor` subcommand under `cmd/cite.go` parent command
- [ ] Register cite doctor in the root `cmd/doctor.go` aggregator so `jig doctor` runs it alongside nope/brew/zed checks
- [ ] Follow existing doctor patterns (e.g. `internal/brew/doctor.go`, `internal/zed/doctor.go`)

## Edge Cases

- No `citations:` configured → skip silently or note "no citations configured"
- No LICENSE/NOTICE file found → warn
- Citation repo is a URL, not a GitHub slug → degrade gracefully (skip license API lookup, still search local files)
- Multiple license files → search all of them

## Open Questions

- Should it also check for SPDX identifiers?
- Should it suggest adding missing attributions, or just report?
