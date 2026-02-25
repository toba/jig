---
# 8yw-lvi
title: Add github_url and github_issue_number to changelog JSON output
status: completed
type: feature
priority: normal
created_at: 2026-02-25T20:22:38Z
updated_at: 2026-02-25T20:26:13Z
sync:
    github:
        issue_number: "69"
        synced_at: "2026-02-25T20:28:08Z"
---

## Context

The `jig changelog --json` output includes `sync.github.issue_number` nested inside each issue, but consuming agents need to construct full GitHub URLs manually. This is error-prone and requires knowing the repo URL separately.

## Proposed Changes

1. Add a top-level `github` field to the changelog JSON output with the project's GitHub repo URL (from `.jig.yaml` sync config):
   ```json
   {
     "github": "https://github.com/toba/xc-mcp",
     "range": { ... },
     "issues": { ... }
   }
   ```

2. For each issue that has a GitHub sync, promote the issue number to a top-level `github_issue` integer field (instead of only `sync.github.issue_number` as a string):
   ```json
   {
     "id": "f6b-ert",
     "title": "Add validate_project tool",
     "github_issue": 134,
     ...
   }
   ```
   Issues without GitHub sync get `"github_issue": null`.

## Why

- The top-level `github` URL lets consumers build issue links without parsing `.jig.yaml` or git remotes
- `github_issue` as an integer at the top level is easier to consume than `sync.github.issue_number` as a nested string
- Together they make changelog formatting trivial: `[#${issue.github_issue}](${github}/issues/${issue.github_issue})`
- The `sync` field remains unchanged for backward compatibility


## Summary of Changes

- Added top-level `github` field to `changelog.Result` struct, populated from `.jig.yaml` sync config
- Added `github_issue` integer field to issue JSON output via `MarshalJSON`, extracted from `sync.github.issue_number`; issues without GitHub sync get `null`
- Added `GithubIssueNumber()` helper method on `Issue`
- Existing `sync` field unchanged for backward compatibility
