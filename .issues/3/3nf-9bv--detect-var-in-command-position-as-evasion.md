---
# 3nf-9bv
title: Detect $var in command position as evasion
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:37:35Z
sync:
    github:
        issue_number: "16"
        synced_at: "2026-02-23T17:08:13Z"
---

nope's CheckSubshell catches \$() and backticks but doesn't detect plain \$var or \${var} as a command name. If the guard can't know what a variable resolves to, it's safer to block.

Patterns to detect in command position:
- \$cmd args
- \${cmd} args
- \$(generate_cmd) as command name (partially covered)

Reference: https://github.com/leegonzales/claude-guardrails

- [x] Extend CheckSubshell or create new built-in for variable-as-command detection
- [x] Handle \$var and \${var} in command position
- [x] Add tests for variable expansion evasion


## Summary of Changes

Added `var-command` builtin check that detects `$var` and `${var}` in command position (first token after stripping wrappers and env assignments in each segment). Since the guard cannot know what a variable resolves to, this blocks potential evasion. Handles wrappers (sudo, env), pipe/chain segments, and excludes quoted variables and non-variable dollar references ($1, $?).
