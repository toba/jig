---
# scy-dno
title: Show milestone prefix on child/linked issues in TUI detail view
status: completed
type: bug
priority: normal
created_at: 2026-05-27T18:22:33Z
updated_at: 2026-05-27T18:23:26Z
sync:
    github:
        issue_number: "112"
        synced_at: "2026-05-27T18:33:46Z"
---

The list of child (and other linked) issues within an issue's detail view should include the milestone short-name prefix, same as the main TUI list rows. Currently linkDelegate.Render calls RenderIssueRow without setting MilestoneShort.


## Summary of Changes

Added a milestone ID→short lookup to `detailModel` (`loadMilestoneShorts`), populated in `newDetailModel` and `refreshIssue`. Threaded it through `linkDelegate` so `RenderIssueRow` sets `MilestoneShort` for each linked/child issue — matching the main TUI list's gray `<short>:` ID prefix.
