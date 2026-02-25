---
# 50n-3of
title: 'Fix TUI layout bug: dimmed parent rows wrap type column'
status: completed
type: bug
priority: normal
created_at: 2026-02-25T00:40:47Z
updated_at: 2026-02-25T00:42:25Z
---

## Problem

When filters are active that include children, parent issues are shown dimmed (grey) for context. The dimmed rendering path in `RenderIssueRow` passes the full type name (e.g. "feature") to `typeStyle.Render(Muted.Render(typeName))` instead of using `TypeAbbrev(typeName)` like the non-dimmed path.

Since `typeStyle` has `Width(2)`, lipgloss wraps the longer string across multiple lines instead of truncating, causing the row to take multiple lines and breaking the entire TUI layout.

## Root Cause

`internal/todo/ui/styles.go:RenderIssueRow` line ~513: dimmed branch uses raw `typeName` instead of `TypeAbbrev(typeName)`.

## Fix

- [x] Add test in `internal/todo/ui/styles_test.go` to verify dimmed rows produce single-line output with abbreviation
- [x] Change dimmed type rendering to use `TypeAbbrev(typeName)`


## Summary of Changes

Fixed `RenderIssueRow` in `internal/todo/ui/styles.go`: the dimmed code path was passing the full type name (e.g. "feature") to `typeStyle.Render(Muted.Render(typeName))`. Since `typeStyle` has `Width(2)`, lipgloss word-wrapped the string across multiple lines instead of truncating, breaking the entire row layout. Changed to use `TypeAbbrev(typeName)` to match the non-dimmed path.
