---
# 916-8b5
title: Add data exfiltration detection patterns
status: completed
type: feature
priority: normal
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T20:22:07Z
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

- [x] Define exfiltration patterns (file+network combos)
- [x] Implement as built-in check or add to starter config regex rules
- [x] Add tests for each exfiltration vector


## Summary of Changes

Added `exfiltration` builtin to nope guard that detects data flowing from sensitive files to network egress:

- **curl/wget uploads**: `-d @.env`, `--data-binary @key`, `-F file=@creds`, `--upload-file`, `-T`, `--post-file`
- **scp of sensitive files**: detects credential files in scp source arguments
- **Bash socket writes**: `/dev/tcp/` and `/dev/udp/` token detection
- **Piped credential access**: sensitive file before `|` followed by network tool (e.g. `cat .env | nc host`)

Files modified:
- `internal/nope/builtins.go` — `CheckExfiltration()` with helpers for curl, wget, scp
- `internal/nope/config.go` — `BuiltinExfiltration` constant, wired into `CompileRules`
- `internal/nope/init.go` — `data-exfiltration` rule in `StarterConfig`
- `internal/nope/help.go` — added exfiltration to BUILTINS section
- `internal/nope/builtins_test.go` — 24 test cases covering all vectors plus integration test
