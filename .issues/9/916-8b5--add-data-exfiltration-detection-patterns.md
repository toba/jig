---
# 916-8b5
title: Add data exfiltration detection patterns
status: ready
type: feature
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

nope has `credential-read` and `network` built-ins separately, but nothing connecting the two. Add detection for exfiltration patterns that combine file access with network egress:

- `curl -F file=@.env` (file upload of secrets)
- `scp ~/.ssh/id_rsa host:` (SCP of credentials)
- `base64 ~/.aws/credentials | curl` (encoded exfil)
- DNS exfiltration via `dig`/`nslookup` with command substitution
- `wget --post-file` / `--post-data` with sensitive paths
- `/dev/tcp` and `/dev/udp` socket writes

Could be a new `exfiltration` built-in or starter regex rules.

Reference: https://github.com/leegonzales/claude-guardrails

- [ ] Define exfiltration patterns (file+network combos)
- [ ] Implement as built-in check or add to starter config regex rules
- [ ] Add tests for each exfiltration vector
