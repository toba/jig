---
# bqt-abz
title: loadFromDisk doesn't skip dot-prefixed subdirectories in .issues/
status: completed
type: bug
priority: normal
created_at: 2026-03-12T19:03:45Z
updated_at: 2026-03-12T19:13:13Z
sync:
    github:
        issue_number: "93"
        synced_at: "2026-03-12T22:45:36Z"
---

In `internal/todo/core/core.go`, `loadFromDisk` uses `filepath.WalkDir` but returns `nil` for directories instead of `filepath.SkipDir` for dot-prefixed entries. Any `.md` files inside dot-prefixed subdirectories (e.g. `.issues/.git/`) would be erroneously loaded as issues.

Fix: add a guard before descending into directories:
```go
if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
    return filepath.SkipDir
}
```

## Summary of Changes

Added `filepath.SkipDir` return for dot-prefixed subdirectories in `loadFromDisk` (`internal/todo/core/core.go`). The root directory itself is excluded from the check. Added `TestLoadFromDiskSkipsDotPrefixedDirs` test.
