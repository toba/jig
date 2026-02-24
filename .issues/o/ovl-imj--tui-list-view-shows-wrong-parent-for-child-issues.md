---
# ovl-imj
title: TUI list view shows wrong parent for child issues
status: completed
type: bug
priority: normal
created_at: 2026-02-24T01:10:00Z
updated_at: 2026-02-24T01:34:11Z
sync:
    github:
        issue_number: "64"
        synced_at: "2026-02-24T01:35:02Z"
---

On the TUI list page, a set of LOS issues (core-jbag, core-gqyt, core-iwgi, core-94dd, core-1a4s, core-372t, core-wvht) are displayed as children of `core-87p9` ("Merge amenity→description + PriceLabs gap closes"), but when viewing an individual issue (e.g. core-jbag "LOS PMS interface abstraction"), the Linked Issues section correctly shows its parent as `core-r6y1` ("Length-of-Stay Rules Redesign" epic).

The tree rendering in the list view is grouping children under the wrong parent node.

Observed in pacer/core project (`../../pacer/core`).

## TODO

- [x] Investigate tree-building logic in `internal/todo/tui/` list view
- [x] Determine why parent association is incorrect in the rendered tree
- [x] Fix: strip tree prefixes when BubbleTea filter is active
- [x] Verify fix against pacer/core dataset


## Root Cause

BubbleTea list filter operates on the flat item list independently of the tree hierarchy. When a search term (e.g. "LOS") matches children but not their parent:

1. Parent `core-r6y1` ("Length-of-Stay Rules Redesign") is filtered out — no "los" substring
2. `core-87p9` ("Merge amenity→description + PriceLabs gap c**los**es") remains — "closes" contains "los"
3. LOS child issues remain — titles contain "LOS"
4. Children retain pre-computed tree prefixes (├─, └─) and visually appear as children of `core-87p9`

## Fix

- **`internal/todo/tui/list.go`**: Tree-aware filter function that preserves ancestor hierarchy. When a child matches the search term, walks backward through the flat list to include parent nodes at progressively lower depths. Ancestors appear dimmed (nil MatchedIndexes) while direct matches are highlighted normally.
- **`internal/todo/tui/list.go`**: Render method dims filter-ancestors (items with no MatchedIndexes when filter is active) so they're visually distinct from actual matches.
- **`internal/todo/ui/tree.go`**: Fix `strings.Repeat` panic with negative count — header used "STATUS" (6 chars) but `ColWidthStatus` is only 3 (icon-width). Abbreviated to "ST" and added `max(0, ...)` guards on all header padding.
