---
# 3qi-e30
title: Update jig scoop init to support shared bucket repos
status: completed
type: feature
created_at: 2026-03-20T19:55:05Z
updated_at: 2026-03-20T19:55:05Z
---

Mirror the brew shared tap changes for scoop. Use `owner/scoop-bucket` convention (like charmbracelet/scoop-bucket). Manifests at repo root (not `bucket/` subdir). No per-project repo creation. Auto-save companions.scoop.

## Requirements

- [x] Convention changed to `owner/scoop-bucket`
- [x] Manifests placed at repo root (not `bucket/` subdir), matching charmbracelet model
- [x] `scoop init` pushes manifest to existing shared bucket repo
- [x] Remove per-project repo creation, README generation
- [x] `scoop doctor` derives tool name from source repo
- [x] Workflow job clones shared bucket repo using `Bucket` field
- [x] Auto-save `companions.scoop` to `.jig.yaml` after init
- [x] Add tests for workflow generation with shared bucket

## Summary of Changes

- Removed per-project bucket creation (`createBucketRepo`, `pushInitialContent`, `generateReadme`)
- `scoop init` clones existing shared bucket, adds manifest at repo root, commits and pushes
- Manifests at repo root (e.g. `jig.json`) instead of `bucket/jig.json`, matching charmbracelet/scoop-bucket model
- Workflow job uses `Bucket` field for clone URL instead of deriving `scoop-<tool>`
- Convention changed from `owner/scoop-<tool>` to `owner/scoop-bucket`
- `scoop doctor` derives tool name from source repo (not bucket name)
- `scoop init` auto-saves `companions.scoop` to `.jig.yaml`
- CI commit message changed to `bump <tool> to` for shared bucket clarity
- Added workflow, manifest, and injection tests
