---
# 49o-xvo
title: jig todo show emits ANSI codes when piped, forcing agents to re-read raw issue files
status: completed
type: bug
priority: high
created_at: 2026-06-25T17:00:10Z
updated_at: 2026-06-25T17:00:10Z
sync:
    github:
        issue_number: "120"
        synced_at: "2026-06-25T17:13:45Z"
---

`jig todo show <id>` emitted truecolor ANSI escape sequences even when stdout was piped (non-TTY). Agents reading the output via `| cat`/`| head` saw garbled escape codes and had to fall back to reading the raw `.issues/*.md` file.

## Fix

- Route styled `show` output through `colorprofile.NewWriter(os.Stdout, os.Environ())`, which detects the destination and strips/downsamples ANSI for non-TTY pipes (and respects `NO_COLOR`/`CLICOLOR`).
- Use glamour's clean `notty` style for the body when color is disabled (avoids background-padding noise); keep the configured terminal style otherwise.
- Refactored `showStyledIssue` into writer-based `writeStyledIssue(w, b, color)` plus a testable `renderIssue(b, color)` helper.

## Summary of Changes

- `cmd/todo_show.go`: writer/color-profile-aware rendering.
- `cmd/todo_show_test.go`: `TestRenderIssuePlainNoANSI` asserts plain output has zero ANSI escapes and still contains ID/title/body; color mode keeps ANSI.
- `cmd/cmd_test.go`: updated `TestShowStyledIssue*` to use `renderIssue`.

Verified: piped output has 0 ANSI codes; interactive/forced-color output keeps color.
