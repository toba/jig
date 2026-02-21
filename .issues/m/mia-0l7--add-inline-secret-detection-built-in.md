---
# mia-0l7
title: Add inline secret detection built-in
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:34:01Z
---

Detect secrets appearing inside command text, e.g. `echo "API_KEY=sk_live_abc123"`. Orthogonal to file-based credential checks.

Patterns to detect:
- AWS access key IDs (`AKIA...`)
- GitHub tokens (`ghp_...`, `github_pat_...`)
- Generic `api_key=`/`secret_key=`/`access_token=` assignments with values
- Inline `password='...'` assignments

Reference: https://github.com/leegonzales/claude-guardrails (common.rs)

- [x] Define regex patterns for common secret formats
- [x] Implement as new `inline-secrets` built-in check
- [x] Add tests for each secret pattern
- [x] Ensure patterns don't false-positive on placeholder values


## Summary of Changes

Added `inline-secrets` builtin check to the nope guard that detects secrets embedded in command text. Patterns cover: AWS access key IDs (AKIA...), AWS secret access keys, GitHub tokens (ghp_/ghs_/gho_/ghu_/ghr_), GitHub fine-grained PATs (github_pat_), generic api_key/secret_key/access_token assignments (16+ char values), and password/passwd/pwd assignments with quoted values. Placeholder values (YOUR_*, xxx, changeme, example, test_key, dummy, sample, fake) are excluded to prevent false positives.
