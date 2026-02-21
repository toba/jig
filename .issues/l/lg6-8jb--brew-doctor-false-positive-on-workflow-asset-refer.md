---
# lg6-8jb
title: 'brew doctor: false positive on workflow asset reference check'
status: ready
type: bug
created_at: 2026-02-21T20:20:19Z
updated_at: 2026-02-21T20:20:19Z
---

## Description

\`jig doctor\` / \`brew doctor\` reports a false positive:

```
FAIL: workflow does not reference asset xc-mcp-v0.19.0-arm64.tar.gz
```

The check in \`internal/brew/doctor.go:179\` uses \`strings.Contains(workflowStr, expectedAsset)\` to look for the **literal** versioned asset name (e.g. \`xc-mcp-v0.19.0-arm64.tar.gz\`) inside the workflow YAML. However, workflows use GitHub Actions variables like \`\${{ github.ref_name }}\`, so the actual string in the file is:

```
xc-mcp-${{ github.ref_name }}-arm64.tar.gz
```

The literal tag value never appears in the workflow, causing the check to always fail despite the workflow correctly producing the asset (confirmed by the release having the expected artifact).

## Reproduction

```bash
cd ../xc-mcp && jig doctor
```

## Fix

Replace the literal asset name check with a pattern match that substitutes the tag portion with the expected variable reference or a wildcard. For example, check for the asset name template (e.g. \`xc-mcp-\${{ github.ref_name }}-arm64.tar.gz\`) instead of the resolved name.

## Files

- \`internal/brew/doctor.go\` — the \`strings.Contains\` check (~line 179)
- \`internal/brew/language.go\` — \`ExpectedAssetName()\` generates the literal name

## Tasks

- [ ] Update asset reference check to match template pattern instead of literal tag
- [ ] Add test case for workflow with variable-based asset names
