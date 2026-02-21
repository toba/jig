---
# whc-ynp
title: 'brew doctor: detect project language and adjust diagnostics'
status: completed
type: feature
priority: normal
created_at: 2026-02-21T06:19:26Z
updated_at: 2026-02-21T19:19:47Z
sync:
    github:
        issue_number: "4"
        synced_at: "2026-02-21T16:40:37Z"
---

\`ja brew doctor\` currently assumes a Go/goreleaser-based release pipeline and fails on non-Go projects that have valid custom release workflows.

## Current behavior

Hardcoded checks for:
- `.goreleaser.yaml` exists
- `goreleaser/goreleaser-action` step in workflow
- Asset named `<tool>_darwin_arm64.tar.gz` (goreleaser convention)
- `checksums.txt` (goreleaser checksum output)

## Expected behavior

Detect the project language/build system and adjust diagnostics accordingly:

### Go projects (goreleaser)
- Keep existing checks: `.goreleaser.yaml`, goreleaser-action, `<tool>_darwin_arm64.tar.gz`, `checksums.txt`

### Swift projects
- Detect via `Package.swift`
- Check for `swift build -c release` in workflow
- Accept asset patterns like `<tool>-<tag>-arm64.tar.gz` (or read from workflow)
- Check for `.sha256` sidecar file instead of `checksums.txt`
- Don't require `.goreleaser.yaml` or goreleaser-action

### Rust projects
- Detect via `Cargo.toml`
- Check for `cargo build --release` or `cross` in workflow
- Adjust expected asset naming accordingly

## Example failure (Swift project: toba/xc-mcp)

```
OK:   companions.brew configured: toba/homebrew-xc-mcp
FAIL: .goreleaser.yaml not found
OK:   tap repo exists: toba/homebrew-xc-mcp
OK:   formula exists: Formula/xc-mcp.rb
OK:   latest release: v0.19.0
FAIL: release v0.19.0 missing asset xc-mcp_darwin_arm64.tar.gz
FAIL: release v0.19.0 missing checksums.txt
OK:   workflow exists: .github/workflows/release.yml
FAIL: workflow missing goreleaser/goreleaser-action step
OK:   workflow has update-homebrew job
OK:   workflow references homebrew-xc-mcp
FAIL: workflow does not reference asset xc-mcp_darwin_arm64.tar.gz
```

The actual release workflow uses `swift build -c release`, a custom tar archive (`xc-mcp-v0.19.0-arm64.tar.gz`), and a `.sha256` sidecar — all valid, but doctor flags them as failures.



## Summary of Changes

Duplicate of pt4-51i which implemented the language detection. Marking as completed.
