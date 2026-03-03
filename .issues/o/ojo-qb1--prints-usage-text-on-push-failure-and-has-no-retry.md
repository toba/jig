---
# ojo-qb1
title: ' prints usage text on push failure and has no retry for transient errors'
status: completed
type: bug
priority: normal
created_at: 2026-03-03T18:52:54Z
updated_at: 2026-03-03T18:56:34Z
---

## Problem

When `jig commit apply --push` encounters a transient network error during `git push`, two things go wrong:

### 1. Usage text printed on push failure

The commit apply cobra command doesn't set `SilenceUsage = true`, so any error (including transient network failures) dumps the full usage text:

```
Committed.
Tagged v1.22.0.
Error: git push: exit status 128
Usage:
  jig commit apply [flags]

Flags:
  -h, --help             help for apply
  -m, --message string   commit message (required)
      --push             push commits and tags after committing
  -v, --version string   version tag to create
...
```

This is misleading — the user didn't make a usage error, the network failed. The `nope` command already sets `SilenceUsage: true` and `SilenceErrors: true` correctly. The commit subcommands should do the same.

### 2. No retry for transient push failures

`commit.Push()` in `internal/commit/commit.go` has no retry logic. When GitHub has transient TLS/connection issues, the push fails immediately. In a real session, the agent had to retry `git push` 4 times manually before it succeeded:

```
git push && git push --tags          → TLS_error, Connection reset by peer
git push && git push --tags          → Connection refused
git push origin main && git push origin v1.22.0  → main succeeded, tag Connection refused
git push origin v1.22.0             → remote: fatal error in commit_refs
git push origin v1.22.0             → success
```

The commit and tag were created successfully — only the push failed. The agent had to know to retry just the push, not re-run the whole apply.

## Suggested Fix

### SilenceUsage

Add `SilenceUsage: true` to commit subcommands, or set it in a `PersistentPreRun` on the root/commit parent command. Alternatively, set `cmd.SilenceUsage = true` in each RunE before returning errors.

### Push retry

Add retry with exponential backoff for transient errors in `commit.Push()`:
- Retry on exit codes that indicate network failure (exit 128 from git)
- 3 attempts with 1s/3s/5s backoff
- Only retry the failed operation (if branch push succeeded, don't re-push it)
- Log each retry attempt so the user knows what's happening

## Observed in

xc-mcp commit session, 2026-03-03. Agent used the `/commit push` skill which calls `jig commit apply --push`.


## Summary of Changes

### 1. SilenceUsage on commit commands (`cmd/commit.go`)
Added `SilenceUsage: true` and `SilenceErrors: true` to the `commitCmd` parent command so transient push errors no longer dump misleading usage text.

### 2. Push retry with exponential backoff (`internal/commit/commit.go`)
Refactored `Push()` to use `pushWithRetry()` which retries on transient git errors (exit code 128 — used for network/TLS/connection failures):
- 3 retry attempts with 1s/3s/5s backoff delays
- Each sub-operation (branch push, individual tag pushes) is retried independently
- Non-transient errors (exit code ≠ 128) fail immediately without retry
- Retry attempts are logged to stderr so the user sees what's happening

### 3. Tests
- `TestCommitCmdSilenceFlags` — verifies SilenceUsage/SilenceErrors on commitCmd
- `TestIsTransientGitError` — exit 128 is transient, exit 1 is not, nil is not
- `TestPushWithRetry` — succeeds first try, non-transient fails fast, transient exhausts retries, transient succeeds on retry
