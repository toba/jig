---
# nw0-yr7
title: Add 'jig todo comment' command for appending to issue body
status: completed
type: feature
priority: normal
created_at: 2026-06-20T19:03:47Z
updated_at: 2026-06-20T19:06:12Z
---

Agents instinctively reach for `jig todo comment <id>` (by analogy with gh/git) to add content to an issue, find it doesn't exist, and fall back to editing the .issues/*.md file directly — which bypasses etag concurrency checks, updated timestamps, and sync.

## Plan
- [x] Failing test for a comment command that appends to an issue body
- [x] Implement `jig todo comment <id> <text>` as a thin, discoverable alias for `update --append-body` (supports `-` stdin, `--json`)
- [x] Update agent guide (todo_prompt.tmpl) to surface the command and warn against editing issue files directly
- [x] Update CLAUDE.md command inventory and CHANGELOG

## Verification

End-to-end test of the new `comment` command — this text was appended via `jig todo comment`, not by editing the file.

## Stdin note

Appended from stdin.

## Summary of Changes

- New `cmd/todo_comment.go`: `jig todo comment <id> <text>` appends to an issue body through the GraphQL `UpdateIssue` `BodyMod.Append` path (same as `update --append-body`), preserving etag checks, the `updated` timestamp, and sync. Supports `-` for stdin and `--json`.
- `cmd/todo_comment_test.go`: failing-first tests covering append/separation, empty-text rejection, and missing-issue error.
- `cmd/todo_prompt.tmpl` (agent guide): surfaced `jig todo comment` in the CLI Reference and Body Modifications, plus a **never edit `.issues/*.md` directly** warning.
- `CLAUDE.md`, `CHANGELOG.md`: documented the command.

Verified end-to-end (direct arg + stdin), `go test ./cmd/`, `go vet`, and `scripts/lint.sh` all clean.
