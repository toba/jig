---
# 8dy-bi9
title: Add compound command splitting for independent segment checking
status: in-progress
type: feature
priority: high
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:56:17Z
---

nope has `CheckChained` to block `&&`/`||`/`;`, but if that rule is disabled, remaining checks don't analyze each sub-command independently. `echo hi && rm -rf /` should have both parts evaluated against all rules.

Split compound commands on `&&`, `||`, `;` operators and run each segment through all applicable rules independently.

Reference: https://github.com/leegonzales/claude-guardrails

- [ ] Split tokenized commands on chain/pipe operators into segments
- [ ] Run each segment through all built-in checks independently
- [ ] Ensure command-position logic resets per segment (already partially done in CheckNetwork)
- [ ] Add tests for dangerous commands hidden after innocuous ones
