---
# mia-0l7
title: Add inline secret detection built-in
status: ready
type: feature
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

Detect secrets appearing inside command text, e.g. `echo "API_KEY=sk_live_abc123"`. Orthogonal to file-based credential checks.

Patterns to detect:
- AWS access key IDs (`AKIA...`)
- GitHub tokens (`ghp_...`, `github_pat_...`)
- Generic `api_key=`/`secret_key=`/`access_token=` assignments with values
- Inline `password='...'` assignments

Reference: https://github.com/leegonzales/claude-guardrails (common.rs)

- [ ] Define regex patterns for common secret formats
- [ ] Implement as new `inline-secrets` built-in check
- [ ] Add tests for each secret pattern
- [ ] Ensure patterns don't false-positive on placeholder values
