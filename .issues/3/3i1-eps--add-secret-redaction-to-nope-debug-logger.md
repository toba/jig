---
# 3i1-eps
title: Add secret redaction to nope debug logger
status: draft
type: task
priority: normal
created_at: 2026-02-21T19:50:20Z
updated_at: 2026-02-21T19:50:20Z
sync:
    github:
        issue_number: "22"
        synced_at: "2026-02-23T17:08:12Z"
---

nope's debug logger writes raw command content to JSONL. The guardrails repo redacts secrets from log entries before writing. Adding redaction before logging would harden the debug output.

Reference: https://github.com/leegonzales/claude-guardrails (common::redact_secrets)

- [ ] Define redaction patterns (reuse inline-secrets patterns if that issue is done)
- [ ] Apply redaction to command content before writing to debug log
- [ ] Add tests for redacted log output
