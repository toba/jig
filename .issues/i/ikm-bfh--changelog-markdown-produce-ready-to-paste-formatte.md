---
# ikm-bfh
title: 'changelog --markdown: produce ready-to-paste formatted output'
status: completed
type: feature
priority: normal
created_at: 2026-03-07T23:49:35Z
updated_at: 2026-03-07T23:56:41Z
sync:
    github:
        issue_number: "86"
        synced_at: "2026-03-07T23:58:25Z"
---

The `jig changelog` command currently outputs raw JSON that the calling agent has to manually parse, diff against existing CHANGELOG.md, categorize by issue type, format with GitHub links, and generate section headers. This is the exact work the skill instructions describe but the CLI should do it.

Add a `--markdown` flag that:

- [ ] Groups completed issues by type (feature→Features, bug→Fixes, task/epic/milestone→Tweaks)
- [ ] Formats entries with GitHub issue links when `github_issue` is set
- [ ] Wraps tool/code names in backticks (reuse title as-is per skill)
- [ ] Generates proper section header (weekly/daily/since)
- [ ] Excludes issues whose IDs already appear in existing CHANGELOG.md
- [ ] Outputs ready-to-paste markdown that can be prepended to CHANGELOG.md
- [ ] Update the changelog skill to use `--markdown` instead of `--json`


## Summary of Changes

Added `--markdown` flag to `jig changelog` that produces ready-to-paste formatted output:
- Groups completed issues by type (Features / Fixes / Tweaks)
- Formats GitHub issue links when synced
- Generates proper section header (weekly/since/daily/append modes)
- Excludes issues already in existing CHANGELOG.md (by issue ID)
- Outputs empty + stderr message when nothing new

Updated both changelog and commit skills to use `--markdown` instead of `--json` + manual categorization.
