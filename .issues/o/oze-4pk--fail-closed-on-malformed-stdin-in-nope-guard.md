---
# oze-4pk
title: Fail closed on malformed stdin in nope guard
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:39:26Z
---

nope returns exit 0 (allow) when stdin JSON is unparseable. Malformed input could be an evasion attempt and should block (exit 2) rather than allow.

Missing config is a deliberate design choice for global installs (exit 0 is correct there), but bad stdin is different — it means the hook was invoked but the input is wrong.

Reference: https://github.com/leegonzales/claude-guardrails (fail-closed design)

- [x] Change ReadHookInput error path from exit 0 to exit 2 for malformed JSON
- [x] Keep exit 0 for missing config (intentional)
- [x] Add tests for malformed stdin scenarios


## Summary of Changes

Changed `RunGuard` to fail closed (exit 2) when `ReadHookInput` returns an error from malformed stdin JSON. Previously returned exit 0 (allow), which could let evasion attempts through. Missing config still returns exit 0 (intentional for global installs). Added `TestReadHookInput` with 7 cases covering valid JSON, empty stdin, null tool_input, missing tool_name, malformed JSON, truncated JSON, and plain text.
