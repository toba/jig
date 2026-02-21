---
# 6cf-f7v
title: Add wrapper command unwrapping to nope guard
status: completed
type: feature
priority: high
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:09:57Z
---

nope's built-in checks operate on shell tokens but don't strip wrapper prefixes. A command like `sudo timeout 30 curl example.com` won't trigger `CheckNetwork` because `curl` isn't in command position.

Implement recursive stripping of wrapper commands before running checks. Handle at minimum: `sudo`, `timeout`, `env`, `nice`, `nohup`, `xargs`, `ionice`, `strace`, `time`, `watch`, `caffeinate`, `doas`. Each wrapper needs option-aware skipping (e.g., `sudo -u root -E` consumes flags before the real command).

Reference: https://github.com/leegonzales/claude-guardrails (wrapper.rs)

- [ ] Add wrapper command list with per-wrapper option parsing
- [ ] Implement recursive unwrapping in shell tokenizer or as preprocessing step
- [ ] Apply unwrapping before all built-in checks
- [ ] Add tests for nested wrappers (`sudo timeout 30 nice -n 10 curl`)
- [ ] Add tests for wrappers with flags that take arguments


## Summary of Changes

Implemented wrapper command unwrapping for the nope guard so that `CheckNetwork` correctly identifies network tools hidden behind wrappers like `sudo`, `timeout`, `env`, `nice`, `nohup`, etc.

- Added `wrapperDef` struct, `wrappers` map, and `SkipWrappers()` function in `internal/nope/shell.go`
- Rewrote `CheckNetwork` in `internal/nope/builtins.go` to use `SkipWrappers` for each segment
- Added `TestSkipWrappers` (12 cases) in `shell_test.go`
- Added 7 wrapper test cases to `TestCheckNetwork` and 1 compound+wrapper integration test in `builtins_test.go`
