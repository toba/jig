---
# mti-ldn
title: Add environment self-defense built-in check
status: ready
type: feature
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

Detect commands that attempt to subvert the guard or hijack the runtime environment:

- `LD_PRELOAD=...` — library injection
- `LD_LIBRARY_PATH=...` — library path hijacking
- Env vars attempting to disable guardrails

Could be a new `env-hijack` built-in check.

Reference: https://github.com/leegonzales/claude-guardrails

- [ ] Define list of dangerous environment variable prefixes/names
- [ ] Implement as new `env-hijack` built-in check
- [ ] Add tests for env var injection patterns
