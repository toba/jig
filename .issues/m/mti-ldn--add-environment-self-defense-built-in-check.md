---
# mti-ldn
title: Add environment self-defense built-in check
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:28:21Z
sync:
    github:
        issue_number: "27"
        synced_at: "2026-02-23T17:08:13Z"
---

Detect commands that attempt to subvert the guard or hijack the runtime environment:

- `LD_PRELOAD=...` — library injection
- `LD_LIBRARY_PATH=...` — library path hijacking
- Env vars attempting to disable guardrails

Could be a new `env-hijack` built-in check.

Reference: https://github.com/leegonzales/claude-guardrails

- [x] Define list of dangerous environment variable prefixes/names
- [x] Implement as new `env-hijack` built-in check
- [x] Add tests for env var injection patterns


## Summary of Changes

Added `env-hijack` builtin check to the nope guard that detects dangerous environment variable assignments in command position. Covers library injection (LD_PRELOAD, LD_LIBRARY_PATH, DYLD_INSERT_LIBRARIES, DYLD_LIBRARY_PATH), runtime hijack (NODE_OPTIONS, PYTHONPATH, PYTHONSTARTUP, PERL5OPT, PERL5LIB, RUBYOPT, RUBYLIB), and explicit `env`/`export` commands. Includes unit tests and compound-segment integration test.
