---
# xdd-qvx
title: TUI list truncates titles incorrectly after Bubble Tea v2 migration
status: completed
type: bug
priority: normal
created_at: 2026-03-21T17:41:11Z
updated_at: 2026-03-21T17:58:03Z
sync:
    github:
        issue_number: "100"
        synced_at: "2026-03-21T18:02:31Z"
---

After migrating to Bubble Tea v2, Lipgloss v2, and Bubbles v2 (commit 0a473a0), the TUI list view miscalculates line lengths, causing issue titles to be truncated too aggressively. Titles that should fit within the terminal width are cut short with `...`.

## Evidence

Screenshot shows titles being truncated even though there is plenty of horizontal space remaining. For example:
- `remove_file removes files from all targets when multiple targets have files with th...` — cut off mid-word despite visible space
- `get_test_attachments parses manifest.json with wrong keys, returns Unnamed/unknown ...` — trailing ellipsis with room to spare

## Likely cause

Lipgloss v2 changed how string width is calculated (e.g. ANSI-aware width, or padding/margin accounting). The truncation logic in the TUI list renderer likely uses a stale or incorrect width calculation that doesn't match v2 behavior.

## Tasks

- [x] Identify the truncation/width calculation in `internal/todo/tui/`
- [x] Compare Lipgloss v1 vs v2 width APIs used
- [x] Fix the width calculation
- [x] Verify titles render without premature truncation


## Summary of Changes

The root cause: Lipgloss v2 changed `Width(N)` semantics. In v1, `Width(N)` set the content width (border added on top). In v2, `Width(N)` is the **total** width including borders — the library internally subtracts border size.

The TUI list border used `Width(m.width - 2)` and `SetSize(m.width - 2, ...)`. In v2, the border content area is `m.width - 4` (total minus 2 border chars), but the list thought it had `m.width - 2` cells — 2 cells too many. This caused titles to be truncated at a width that was narrower than the actual available space.

**Files changed:**
- `internal/todo/tui/list.go` — `SetSize(msg.Width-4, msg.Height-6)` and `CalculateResponsiveColumns(msg.Width-4, ...)`
- `internal/todo/tui/detail.go` — viewport width: `msg.Width - 6` (matching border content area)
- `internal/todo/tui/list_test.go` — added `TestListWidthAccountsForBorder` regression test
- `internal/todo/ui/styles_test.go` — added `TestRenderIssueRow_RowFitsInBorderContentArea`
- `internal/todo/tui/app_update_test.go` — updated test helpers to use correct sizes
