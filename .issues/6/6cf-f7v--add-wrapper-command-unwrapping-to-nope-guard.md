---
# 6cf-f7v
title: Add wrapper command unwrapping to nope guard
status: ready
type: feature
priority: high
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

nope's built-in checks operate on shell tokens but don't strip wrapper prefixes. A command like `sudo timeout 30 curl example.com` won't trigger `CheckNetwork` because `curl` isn't in command position.

Implement recursive stripping of wrapper commands before running checks. Handle at minimum: `sudo`, `timeout`, `env`, `nice`, `nohup`, `xargs`, `ionice`, `strace`, `time`, `watch`, `caffeinate`, `doas`. Each wrapper needs option-aware skipping (e.g., `sudo -u root -E` consumes flags before the real command).

Reference: https://github.com/leegonzales/claude-guardrails (wrapper.rs)

- [ ] Add wrapper command list with per-wrapper option parsing
- [ ] Implement recursive unwrapping in shell tokenizer or as preprocessing step
- [ ] Apply unwrapping before all built-in checks
- [ ] Add tests for nested wrappers (`sudo timeout 30 nice -n 10 curl`)
- [ ] Add tests for wrappers with flags that take arguments
