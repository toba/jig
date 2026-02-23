---
# 87a-ark
title: jig commit apply hangs indefinitely when pre-commit hooks spawn child processes
status: completed
type: bug
priority: high
created_at: 2026-02-23T00:13:43Z
updated_at: 2026-02-23T00:16:30Z
---

## Problem

`jig commit apply -m "..." -v v1.10.3 --push` hangs indefinitely when the target repo has a pre-commit hook that runs long-lived or child-spawning commands (e.g., `swift build` + `swift test`).

Observed in `/Users/jason/Developer/toba/xc-mcp` where the pre-commit hook runs:
1. `swiftformat Sources Tests`
2. `swiftlint`
3. `swift build`
4. `swift test` (500+ tests, 14s–300s)

The command was run as a background task (via Claude Code Bash tool with `run_in_background`). It produced zero output and never completed. The same commit succeeded when run manually via `git add && git commit`.

## Root cause

`Commit()` in `internal/commit/commit.go` uses `cmd.CombinedOutput()`:

```go
func Commit(message string) error {
    cmd := exec.Command("git", "commit", "-m", message)
    out, err := cmd.CombinedOutput()
    ...
}
```

`CombinedOutput()` creates OS pipes for stdout/stderr and reads from them until EOF. EOF occurs only when **all** processes holding the pipe write-end file descriptors close them.

When `git commit` runs the pre-commit hook, the hook's child processes (`swift build`, `swift test`, and their sub-processes like `swift-frontend`, `swift-build-tool`, `xctest`) inherit the pipe write ends via standard Unix fork+exec fd inheritance. If any grandchild process outlives `git commit` (daemon processes, orphaned test runners, etc.), the pipe write end stays open and `CombinedOutput()` blocks forever waiting for EOF.

This is a documented failure mode in Go's `os/exec` package — Go 1.20 added `Cmd.WaitDelay` specifically to address it:

> "a child process that exits but leaves its I/O pipes unclosed"

## Fix

Set `cmd.WaitDelay` on the `exec.Command` in `Commit()` (and ideally all long-running exec calls in `internal/commit/commit.go`):

```go
func Commit(message string) error {
    cmd := exec.Command("git", "commit", "-m", message)
    cmd.WaitDelay = 10 * time.Second
    out, err := cmd.CombinedOutput()
    ...
}
```

When git exits but the pipes remain open (because a grandchild holds them), `WaitDelay` force-closes the pipes after the specified duration, unblocking `CombinedOutput()`.

The `Push()` function has a similar pattern (`exec.Command("git", "push").Run()`) but `Run()` with nil Stdout/Stderr doesn't create pipes, so it's not affected. Still worth auditing all `Output()`/`CombinedOutput()` calls in the package.

## Additional consideration

The `TodoSync()` function uses fire-and-forget `cmd.Start()` without setting Stdout/Stderr, so orphaned `todo sync` processes don't inherit jig's pipes and are not part of this bug.



## Summary of Changes

Added `cmd.WaitDelay = 10 * time.Second` to the `Commit()` function in `internal/commit/commit.go`. When git exits but pipe file descriptors remain open (because a pre-commit hook's grandchild process inherited them), `WaitDelay` force-closes the pipes after 10 seconds, unblocking `CombinedOutput()`.

Other `exec.Command` calls in the package were audited:
- `Output()` calls (git log, git diff, git status, git tag, git ls-files) don't trigger hooks — not affected
- `Run()` calls (git add, git push, git tag) don't create pipes — not affected
- `TodoSync()` uses `Start()` with nil Stdout/Stderr — not affected
