---
# 8dy-bi9
title: Add compound command splitting for independent segment checking
status: completed
type: feature
priority: high
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:03:59Z
sync:
    github:
        issue_number: "26"
        synced_at: "2026-02-23T17:08:13Z"
---

nope has `CheckChained` to block `&&`/`||`/`;`, but if that rule is disabled, remaining checks don't analyze each sub-command independently. `echo hi && rm -rf /` should have both parts evaluated against all rules.

Split compound commands on `&&`, `||`, `;` operators and run each segment through all applicable rules independently.

Reference: https://github.com/leegonzales/claude-guardrails

- [x] Split tokenized commands on chain/pipe operators into segments
- [x] Run each segment through all built-in checks independently
- [x] Ensure command-position logic resets per segment (already partially done in CheckNetwork)
- [x] Add tests for dangerous commands hidden after innocuous ones


## Summary of Changes

Added `SplitSegments()` in `shell.go` to split compound commands on `&&`/`||`/`;` (preserving pipes). Modified `CheckRules()` in `check.go` to check each segment independently via `syntheticInput()`. Added tests in `shell_test.go`, `builtins_test.go`, and `check_test.go` covering segment splitting, builtin detection across segments, and pattern rule matching on individual segments.
