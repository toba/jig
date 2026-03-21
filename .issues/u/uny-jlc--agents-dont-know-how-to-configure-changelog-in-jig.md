---
# uny-jlc
title: Agents don't know how to configure changelog in .jig.yaml
status: completed
type: bug
priority: normal
created_at: 2026-03-21T17:55:11Z
updated_at: 2026-03-21T17:57:46Z
sync:
    github:
        issue_number: "99"
        synced_at: "2026-03-21T18:02:31Z"
---

Agents (Claude Code) don't know that changelog is configured via the `changelog:` key in `.jig.yaml`. When asked to set up changelog, they respond:

> jig changelog doesn't have a config file section — it only takes CLI flags. There's nothing to add to .jig.yaml for it.

This is incorrect — the changelog config is a simple top-level key in `.jig.yaml` (e.g. `changelog: weekly`). The issue is likely that:

1. The agent instructions / CLAUDE.md doesn't document the `changelog:` config key
2. The `jig help-all` or `jig prime` output doesn't mention `changelog:` as a `.jig.yaml` section

## Tasks

- [x] Verify `jig prime` / `jig help-all` output mentions changelog config
- [x] Add changelog config documentation to agent-facing output if missing


## Summary of Changes

Added a `## Changelog` section to `cmd/todo_prompt.tmpl` (the agent prompt template used by `jig prime`) documenting:
- The `changelog:` config key in `.jig.yaml` with valid values (`weekly`, `per-commit`)
- How it integrates with the `/commit` skill
- CLI usage examples for `jig changelog`
