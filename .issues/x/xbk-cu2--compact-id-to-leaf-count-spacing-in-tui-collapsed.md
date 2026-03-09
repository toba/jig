---
# xbk-cu2
title: Compact ID-to-leaf-count spacing in TUI collapsed view
status: completed
type: bug
priority: normal
created_at: 2026-03-09T17:38:38Z
updated_at: 2026-03-09T17:38:38Z
sync:
    github:
        issue_number: "90"
        synced_at: "2026-03-09T17:38:58Z"
---

When the TUI list is in collapsed view, there's excessive whitespace between the ID column and the leaf count column. The idColWidth is calculated once from all issues including nested children at various tree depths, adding maxDepth*3 chars for tree indentation. But when items are collapsed, all visible items are at depth 0 with no tree prefix — that indentation space is wasted.

- [x] Add fullIDColWidth field to store original calculated width
- [x] Recalculate idColWidth from visible items after collapse filtering
- [x] Use fullIDColWidth in rebuildVisibleItems for correct recalculation on toggle
- [x] Build, test, lint clean

## Summary of Changes

Added visible-item-aware ID column width recalculation in the TUI. When items are collapsed, idColWidth is computed from the actual visible items (which have no tree prefix at depth 0) instead of using the pre-collapse value that includes deep tree indentation padding. Stored the original full width in fullIDColWidth so expand/collapse toggles recalculate correctly each time.
