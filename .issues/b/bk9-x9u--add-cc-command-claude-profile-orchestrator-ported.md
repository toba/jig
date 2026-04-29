---
# bk9-x9u
title: Add cc command (Claude profile orchestrator) ported from aimux
status: completed
type: feature
priority: normal
created_at: 2026-04-29T01:51:40Z
updated_at: 2026-04-29T02:05:52Z
sync:
    github:
        issue_number: "103"
        synced_at: "2026-04-29T02:06:50Z"
---

Port the capabilities of [aimux](https://github.com/digital-threads/aimux) (`/Users/jason/Developer/temp/aimux/`) into jig as a new `cc` command. aimux is a local AI workspace orchestrator that manages multiple Claude Code subscriptions with a shared knowledge layer and isolated authentication — symlinking shared resources (agents, skills, commands, memory, settings) from `~/.claude` while keeping credentials and session state private per profile.

## Goals

- Add `jig cc` parent command with subcommands modeled on aimux
- Use the term **alias** instead of **profile** (simpler, fits the `jig cc <alias>` invocation pattern)
- Reuse jig conventions: config in `.jig.yaml` (or `~/.jig.yaml` global), Cobra subcommands under `cmd/cc/`, internal logic in `internal/cc/`

## Proposed Commands

- `jig cc init` — auto-detect existing `~/.claude*` dirs, create config, migrate aliases
- `jig cc add <alias>` — create a new alias (replaces aimux's `profile add`)
- `jig cc list` — list aliases
- `jig cc update <alias>` — update model / CLI settings
- `jig cc remove <alias>` — remove alias, clean up symlinks
- `jig cc clone <src> <alias>` — clone alias with private files
- `jig cc rebuild [alias]` — sync symlinks, surface conflicts
- `jig cc doctor` — health check (broken symlinks, missing shared entries, conflicts)
- `jig cc auth login <alias>` — launch OAuth flow
- `jig cc auth status` — show auth file status per alias
- `jig cc <alias>` — **launch claude with that alias** (no `run` subcommand needed; the alias name itself is the verb)
- `jig cc` (no args) — interactive picker (↑↓ + Enter), history pre-selects last used
- Prefix matching on alias names (e.g. `jig cc w` → `work`)
- Model override flag passes through: `jig cc w -m claude-sonnet-4-6`
- Other unknown flags pass through to the `claude` CLI (e.g. `--resume`)

## Architecture Notes

- Source of truth stays at `~/.claude/`
- Aliases live at `~/.jig/cc/<alias>/` (or similar) with symlinks to shared dirs (agents, skills, commands, memory, settings.json) and real files for private items (`.credentials.json`, `.claude.json`, `settings.local.json`, etc.)
- Launches set `CLAUDE_CONFIG_DIR=~/.jig/cc/<alias>` before exec'ing `claude`
- Config under a `cc:` section in `.jig.yaml` (or a dedicated global config file — TBD; aimux uses `~/.aimux/config.yaml`, so a global `~/.jig/cc.yaml` may fit better since aliases are user-global, not project-scoped)
- Add `cc` to the top-level `jig doctor` aggregator
- Shell completion for alias names

## Open Questions

- Global vs. project config? Aliases are inherently user-global (one `~/.claude` source of truth), so likely `~/.jig/cc.yaml` rather than `.jig.yaml`
- Naming collision: top-level `jig cc <alias>` must not shadow subcommands. Cobra can handle this with a custom dispatch in `RunE` on the parent (similar to how `nope` guard mode works per CLAUDE.md)
- Should we also support non-claude CLIs (aimux has a `cli:` field per profile)? Probably yes for parity

## Reference

- aimux source: `/Users/jason/Developer/temp/aimux/` (TypeScript/Node)
- aimux README documents commands, config schema, and symlink layout


## Summary of Changes

Added `jig cc` parent command — Claude Code profile orchestrator — with the following capabilities:

### Commands
- `jig cc init` — auto-detect `~/.claude*` dirs, write `~/.jig/cc.yaml`, copy private files, build symlinks
- `jig cc add <alias>` — create new alias under `~/.jig/cc/<alias>` with shared symlinks
- `jig cc list` — list aliases (text + `--json`)
- `jig cc remove <alias> [--keep-dir]` — remove (refuses to remove source)
- `jig cc clone <src> <alias>` — clone alias including private files
- `jig cc rebuild [alias]` — sync symlinks; reports created / skipped / repaired / conflicts
- `jig cc doctor` — health check (broken / missing / conflict / orphaned), wired into top-level `jig doctor`
- `jig cc auth login <alias>` — execs `<cli> /login` with `CLAUDE_CONFIG_DIR` set
- `jig cc auth status` — shows per-alias presence of credential files
- `jig cc <alias> [claude flags...]` — launches the CLI with prefix matching, all unknown args forwarded
- `jig cc` (no args) — Bubble Tea v2 picker, last-used pre-selected via per-cwd history

### Files added
- `internal/cc/{paths,config,symlinks,init,run,history,picker,output}.go`
- `internal/cc/{config,symlinks,init}_test.go`
- `cmd/cc.go` + `cmd/cc_{init,add,list,remove,clone,rebuild,doctor,auth,auth_login,auth_status}.go`

### Files modified
- `cmd/doctor.go` — added `cc` to the aggregator

### Notes
- Config is global at `~/.jig/cc.yaml`, history at `~/.jig/cc-history.yaml`
- Per-alias dirs at `~/.jig/cc/<name>/` with absolute symlinks to entries in `shared_source` and real per-alias copies for the 9 default private items
- Top-level `jig cc <alias>` uses `DisableFlagParsing` + manual subcommand routing in `RunE`, mirroring how `git <alias>` works — all flags after the alias name pass through to `claude` untouched
- No special `--model` handling: native `claude` flag passes through
- No `update` or `status` commands per scope decision
- All tests pass; `scripts/lint.sh` clean
