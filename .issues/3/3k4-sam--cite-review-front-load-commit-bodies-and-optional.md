---
# 3k4-sam
title: 'cite review: front-load commit bodies and optional diffs'
status: completed
type: feature
priority: normal
created_at: 2026-05-07T15:53:47Z
updated_at: 2026-05-07T16:04:04Z
sync:
    github:
        issue_number: "105"
        synced_at: "2026-05-07T16:09:00Z"
---

`jig cite review --json` currently returns only metadata per commit (`sha`, `message` subject, `author`, `date`, `level`) and a flat `files[]` with `path`/`level`. Agents consuming the output almost always need a follow-up `gh api` / clone step to read the actual commit body or diff before deciding if a change is relevant.

## Proposal

1. **Always include the full commit message body** (not just the subject line) in each commit object. This is one extra field, ~free to fetch from the GitHub API, and often enough on its own to triage relevance.

2. **Add a `--with-diffs` flag** that includes a unified diff per changed file inline in the JSON. Cap per-file size (e.g. 500 lines or 50 KB; skip + flag oversized files) so output stays bounded. Off by default to keep the default review path fast and small.

## Acceptance

- [x] `commits[].body` (or similar) field populated with full commit message body
- [x] `--with-diffs` flag adds `files[].diff` (unified diff text) with per-file size cap and a skipped/truncated indicator when exceeded
- [x] Default behavior unchanged in size/latency
- [x] Skill prompt at `~/.claude/skills/cite/SKILL.md` updated to mention the new fields/flag

## Context

Surfaced while running `/cite review` in the swiftiomatic repo on 2026-05-07 — three HIGH commits flagged across swift-syntax and SwiftLint, and the agent had to choose between filing follow-ups blind or making extra `gh` calls to read the diffs.



## Summary of Changes

- `internal/github/types.go`: split `commit.message` into `Message` (subject) + `Body` (rest, trimmed) via new `splitMessage`; added `File.Patch` and `Release.Body`.
- `internal/display/display.go`: added `CommitResult.Body`, `ReleaseInfo.Body`, and `FileResult.Diff`/`DiffTruncated`/`DiffSkipped` (all `omitempty`).
- `cmd/check.go`: added `--with-diffs` flag plus `capDiff` (500-line / 50 KB caps; truncate or skip), `patchesByFilename`, `buildFileResult`. Both `checkSource` and `checkSourceReleases` now propagate `body` per commit and per-file diffs when the flag is set; release-tracked sources surface release notes via `release.body`. CHANGELOG.md scraping intentionally deferred — would require extra fetches with low hit rate.
- Tests: `TestSplitMessage`, body assertion in `TestCommitNormalize`, plus new `cmd/check_test.go` covering `capDiff` (small / line-truncate / oversize-skip) and `buildFileResult` gating on `--with-diffs`.
- `~/.claude/skills/cite/SKILL.md`: documents the new `body`, `release.body`, `--with-diffs`, `diff_truncated`, `diff_skipped` fields.

Default output adds only the per-commit `body` field — no extra API calls — so the default review path stays the same size/latency. Diffs are off by default.
