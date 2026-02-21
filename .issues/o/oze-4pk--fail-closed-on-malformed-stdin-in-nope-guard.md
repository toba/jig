---
# oze-4pk
title: Fail closed on malformed stdin in nope guard
status: ready
type: feature
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

nope returns exit 0 (allow) when stdin JSON is unparseable. Malformed input could be an evasion attempt and should block (exit 2) rather than allow.

Missing config is a deliberate design choice for global installs (exit 0 is correct there), but bad stdin is different — it means the hook was invoked but the input is wrong.

Reference: https://github.com/leegonzales/claude-guardrails (fail-closed design)

- [ ] Change ReadHookInput error path from exit 0 to exit 2 for malformed JSON
- [ ] Keep exit 0 for missing config (intentional)
- [ ] Add tests for malformed stdin scenarios
