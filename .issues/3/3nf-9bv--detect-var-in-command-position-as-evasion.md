---
# 3nf-9bv
title: Detect $var in command position as evasion
status: ready
type: feature
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

nope's CheckSubshell catches \$() and backticks but doesn't detect plain \$var or \${var} as a command name. If the guard can't know what a variable resolves to, it's safer to block.

Patterns to detect in command position:
- \$cmd args
- \${cmd} args
- \$(generate_cmd) as command name (partially covered)

Reference: https://github.com/leegonzales/claude-guardrails

- [ ] Extend CheckSubshell or create new built-in for variable-as-command detection
- [ ] Handle \$var and \${var} in command position
- [ ] Add tests for variable expansion evasion
