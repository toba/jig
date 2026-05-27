---
# q2b-yqc
title: Move milestone short to first column in TUI rows
status: completed
type: feature
priority: normal
created_at: 2026-05-27T18:02:45Z
updated_at: 2026-05-27T18:11:42Z
sync:
    github:
        issue_number: "109"
        synced_at: "2026-05-27T18:11:56Z"
---

Render the milestone short name as the very first column (before the issue ID) instead of a bracketed badge column after status. Glue the milestone short to the front of the ID as a gray '<short>:' prefix (ID stays purple), e.g. 'b1:issue-749'; just '<id>' when unassigned. Applies in tree view too.

## Summary of Changes

- styles.go: milestone column now renders as the first column (after the cursor, before the ID). Short name right-aligned within `MilestoneColWidth` then a single trailing space; padding-only when an issue has no milestone. Removed the old bracketed badge between status and priority.
- list.go: column width is now max short length (dropped the +2 for brackets).
- Updated doc comments.
