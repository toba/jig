---
# g4v-mqe
title: cite review updates last_checked_sha to oldest commit instead of newest
status: completed
type: bug
priority: normal
created_at: 2026-03-03T18:25:19Z
updated_at: 2026-03-03T18:32:55Z
sync:
    github:
        issue_number: "72"
        synced_at: "2026-03-03T18:33:34Z"
---

## Bug

`jig cite review` sets `last_checked_sha` to the **oldest** new commit instead of the **newest**. This causes subsequent reviews to re-show already-reviewed commits.

## Root Cause

`cmd/check.go:144`:
```go
headSHA := commits[0].SHA
```

When `last_checked_sha` exists (line 115), the code uses the GitHub **Compare API** (`/repos/{owner}/{repo}/compare/{base}...{head}`). The Compare API returns commits in **chronological order** (oldest first), unlike the List Commits API which returns newest first. So `commits[0]` is the oldest commit in the range.

## Evidence

Observed in toba/xc-mcp citations:

**tuist/xcodeproj** (5 new commits):
- `last_checked_sha` updated to `01fbdd7` (Feb 26, first/oldest in results)
- Should have been `b62b255` (Mar 3, last/newest in results)

**getsentry/XcodeBuildMCP** (60 new commits):
- Updated to `b90c0a6` (Feb 12, first/oldest)
- Should have been `7379a39` (Mar 3, last/newest)

## Fix

In `cmd/check.go`, change line 144 from:
```go
headSHA := commits[0].SHA
```
to:
```go
headSHA := commits[len(commits)-1].SHA
```

The Compare API always returns oldest-first. The `GetCommits` path (first run, no `last_checked_sha`) returns newest-first from the List Commits API, so it needs `commits[0]` — meaning the two paths need different indexing. The simplest fix is to branch:

```go
var headSHA string
if src.LastCheckedSHA == "" {
    headSHA = commits[0].SHA           // List API: newest first
} else {
    headSHA = commits[len(commits)-1].SHA  // Compare API: oldest first
}
```

Alternatively, always use `commits[len(commits)-1]` and reverse the first-run path, but that changes more behavior.

## Impact

Every `jig cite review` that finds new commits via Compare will re-show most of the same commits on the next review. Only one commit is "consumed" per review cycle.

## Summary of Changes

Fixed `cmd/check.go` to track which GitHub API returned the commits (`newestFirst` flag). The Compare API returns commits oldest-first, while the List Commits API returns newest-first. Previously, `commits[0]` was always used, which picked the oldest commit from Compare results. Now the code indexes from the correct end based on which API was called.
