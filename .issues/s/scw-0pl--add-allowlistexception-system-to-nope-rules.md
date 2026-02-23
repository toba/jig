---
# scw-0pl
title: Add allowlist/exception system to nope rules
status: draft
type: feature
priority: normal
created_at: 2026-02-21T19:50:20Z
updated_at: 2026-02-21T19:50:20Z
sync:
    github:
        issue_number: "21"
        synced_at: "2026-02-23T17:08:12Z"
---

Allow users to define exceptions that bypass specific rules for known-safe patterns (e.g., `rm -rf ./node_modules`). The guardrails repo uses an allowlist file checked before security rules as an early exit.

This may conflict with nope's design where users simply don't add rules for things they want to allow. Needs design thought on whether this adds value given user-defined rules.

Reference: https://github.com/leegonzales/claude-guardrails (allowlist.toml)

- [ ] Evaluate whether allowlists add value vs. just not defining rules
- [ ] Design allowlist format if proceeding
- [ ] Implement allowlist checking before rule evaluation
