---
# 3xk-ik9
title: brew/scoop doctor should not require goreleaser for projects with manual builds
status: completed
type: bug
priority: normal
created_at: 2026-03-05T19:39:32Z
updated_at: 2026-03-05T19:46:12Z
sync:
    github:
        issue_number: "76"
        synced_at: "2026-03-05T19:49:30Z"
---

`jig brew doctor` and `jig scoop doctor` fail with:

> FAIL: .goreleaser.yaml not found
> FAIL: workflow missing goreleaser-action

Projects that use manual cross-compilation in their release workflow (without goreleaser) should not fail these checks. The doctor should either:
- Accept manual build workflows as valid when assets follow the expected naming convention
- Make goreleaser checks a warning instead of a failure

## Summary of Changes

Made goreleaser checks non-fatal when goreleaser is absent:

1. **brew/doctor.go**: `checkGoreleaser` now prints WARN and returns true (pass) when `.goreleaser.yaml` is missing, instead of FAIL.
2. **brew/language.go**: Go workflow build markers now accept `go build` and `GOOS=` in addition to `goreleaser/goreleaser-action`, so manual build workflows pass validation.
3. **scoop/doctor.go**: Goreleaser check is now conditional — only runs when `.goreleaser.yaml` exists. Otherwise prints WARN.
4. **brew/language_test.go**: Updated test expectations for expanded Go build markers.
