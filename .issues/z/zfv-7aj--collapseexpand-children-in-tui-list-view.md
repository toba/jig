---
# zfv-7aj
title: Collapse/expand children in TUI list view
status: completed
type: feature
priority: normal
created_at: 2026-03-09T16:21:38Z
updated_at: 2026-03-09T16:25:18Z
sync:
    github:
        issue_number: "89"
        synced_at: "2026-03-09T16:29:07Z"
---

Add z (toggle single) and Z (toggle all) keybindings to collapse/expand top-level issues in the TUI tree view, hiding descendants. Collapsed items show a leaf-count badge.

- [x] Add LeafCounts() and countLeaves() to tree.go
- [x] Add RootID field to FlatItem, propagate in flattenNodes
- [x] Add collapse state (collapsed map, leafCounts) to listModel
- [x] Add leafCount field to issueItem
- [x] Update issuesLoadedMsg and loadIssues with leafCounts
- [x] Apply collapse filtering in issuesLoadedMsg handler
- [x] Add z and Z key handlers in Update()
- [x] Render leaf count badge in RenderIssueRow
- [x] Update help overlay and footer
- [x] Tests for LeafCounts and RootID propagation
- [x] All tests pass, lint clean


## Summary of Changes

Added collapse/expand functionality to the TUI list view. Press `z` to toggle collapse on the root ancestor of the selected item, `Z` to toggle all roots. Collapsed items show a leaf-count badge. Changes span tree.go (LeafCounts, RootID on FlatItem), styles.go (LeafCount badge in RenderIssueRow), list.go (collapse state, key handlers, applyCollapse filtering), and help.go (shortcut docs).
