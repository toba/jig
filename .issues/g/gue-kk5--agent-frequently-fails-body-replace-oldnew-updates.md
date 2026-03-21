---
# gue-kk5
title: Agent frequently fails body-replace-old/new updates due to backtick/escaping mismatches
status: completed
type: bug
priority: high
created_at: 2026-03-21T00:31:17Z
updated_at: 2026-03-21T00:37:13Z
sync:
    github:
        issue_number: "97"
        synced_at: "2026-03-21T00:39:03Z"
---

## Problem

The agent constantly fails when using `jig todo update` with `--body-replace-old` / `--body-replace-new` flags. The replacement text doesn't match what's actually in the issue body, causing `text not found in body` errors.

Common failure modes:
- [x] Backticks in markdown (`` ` ``) get escaped or mangled by the shell
- [x] Agent guesses at the exact body text instead of reading the issue first
- [x] Whitespace or newline differences between what the agent passes and what's in the file

## Evidence

```
Error: replacement 0 failed: text not found in body
```

This happens constantly and blocks routine issue management.

## Possible Solutions

- [ ] Investigate: make `--body-replace-old` matching more forgiving (trim whitespace, normalize)
- [x] Investigate: add a `--body-check-item` flag that toggles a specific checkbox by index or substring match
- [ ] Investigate: improve agent instructions to always `jig todo show` before attempting replacements
- [ ] Investigate: support regex or fuzzy matching in replace operations

## Summary of Changes

Added `--body-check` and `--body-uncheck` flags to `jig todo update` that toggle checkbox items by case-insensitive substring match. This eliminates the need for exact text matching with `--body-replace-old`/`--body-replace-new` when toggling checkboxes, which was the primary failure mode for agents.

### Files changed
- `internal/todo/issue/content.go` — `CheckItem()`, `UncheckItem()`, `toggleCheckbox()` functions
- `internal/todo/issue/content_test.go` — unit tests for check/uncheck
- `internal/todo/graph/schema.graphqls` — `check`/`uncheck` fields on `BodyModification` input
- `internal/todo/graph/schema.resolvers.go` — resolver logic for check/uncheck
- `internal/todo/graph/schema.resolvers_test.go` — integration tests
- `internal/todo/graph/generated.go` — regenerated gqlgen code
- `internal/todo/graph/model/models_gen.go` — regenerated model
- `cmd/todo_update.go` — `--body-check`/`--body-uncheck` CLI flags
- `cmd/cmd_test.go` — flag existence test

### Usage
```
jig todo update <id> --body-check "substring"    # checks matching - [ ] item
jig todo update <id> --body-uncheck "substring"  # unchecks matching - [x] item
```

Also works via GraphQL: `bodyMod: { check: ["substring"] }`
