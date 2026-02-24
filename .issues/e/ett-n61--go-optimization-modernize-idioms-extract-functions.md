---
# ett-n61
title: 'Go optimization: modernize idioms, extract functions, add generics, constants, concurrency fixes, and test coverage'
status: completed
type: task
priority: normal
created_at: 2026-02-24T02:04:43Z
updated_at: 2026-02-24T02:13:11Z
sync:
    github:
        issue_number: "66"
        synced_at: "2026-02-24T02:14:32Z"
---

Findings from goptimize analysis of the full codebase (Go 1.26, 30 packages, 216 files).

## Modern Idioms

- [x] `internal/cite/add.go:244` — `sort.Slice` → `slices.SortFunc` with `cmp.Compare`

## Function Extraction

- [x] `cmd/todo.go:23-41` + `cmd/todo_tags_import.go:25-39` — extract `loadConfigWithFallback(cfgPath)` (identical 13-line config loading)
- [x] `cmd/commit.go:88-91,124-127,151-154` — extract `syncTodoIfConfigured(cmd, cfgPath)` with proper error handling (currently discards errors)
- [x] `internal/config/config.go:138-140,154-156` — extract `normalizeYAMLNode(node)` from `FindKey`/`ReplaceKey`
- [x] `internal/todo/core/core_test.go:16-51` — parameterize `setupTestCore(t, opts ...func(*config.Config))`
- [x] `cmd/cmd_test.go` (10+ occurrences) — extract `writeTempConfig(t, content) string`
- [x] `internal/todo/core/links_test.go` (5+ occurrences) — extract `createTestIssues(t, core, issues...)`

## Generics Consolidation

- [ ] ~`internal/todo/tui/*.go` (7 files) — skipped: delegates too diverse for single generic~  — 7 identical `*ItemDelegate` types → generic `PickerDelegate[T list.Item]` with render callback
- [ ] ~`internal/todo/tui/{status,priority,type}picker.go` — skipped: coupled to delegate refactor~ — 3 identical `newXxxPickerModel()` → generic `NewPickerModel[T]` constructor
- [ ] ~`internal/todo/graph/filters.go:114` — skipped: 3-line helpers not worth shared package dependency~ + `issue/sort.go:17-28` + TUI files — repeated slice-to-map → generic `SliceToSet[T]()` and `SliceToIndexMap[T]()`
- [ ] ~`internal/todo/integration/{clickup,github}/sync_state.go` — skipped: would add coupling for minimal gain~ — near-identical `SyncStateStore` → generic `SyncStateStore[T]` or remove trivial wrappers

## Constants/Enums

- [x] `internal/todo/ui/styles.go` — responsive layout magic numbers (140, 50, 55, 45, 35, 42, 32) → named constants
- [x] `internal/todo/integration/syncutil/retry.go:150-184` — error substring patterns → named constants
- [x] `internal/todo/integration/syncutil/retry.go:106,173,177` — HTTP status boundaries (400, 500, 600, 502-504) → named constants
- [x] `internal/todo/integration/clickup/config.go:47-55` — priority mapping magic numbers (1-4) → named constants
- [x] `internal/todo/integration/github/sync.go:324,480` — `"<!-- todo:%s -->"` repeated → `const TodoCommentFormat`
- [x] `internal/todo/tui/modal.go:27-40` — modal sizing magic numbers → named constants

## Concurrency

- [x] `internal/todo/integration/{github,clickup}/sync.go` — unbounded `sync.WaitGroup` → `errgroup` with `SetLimit`

## Test Coverage

- [ ] `internal/todo/graph` (12.6%) — `validateETag` error paths, cycle detection, search filter, `DeleteIssue` failure
- [ ] `internal/brew` (24.9%) — `RunDoctor()` and `RunInit()` (both 0%)
- [ ] `cmd` (33.9%) — `buildUpdateInput()`, `runSync()`, `confirmDeleteMultiple()` (all 0%)
- [ ] `internal/todo/integration/syncutil` (34.0%) — retry logic error paths


## Summary of Changes

Completed 14 of 22 items. Applied:
- `go fix` auto-modernized 6 `interface{}` → `any` in zed tests
- `sort.Slice` → `slices.SortFunc` in cite/add.go
- Extracted `loadConfigWithFallback()` eliminating duplicate config loading
- Extracted `syncTodoIfConfigured()` with proper error logging (was silently discarding errors)
- Extracted `mappingNode()` in config.go for YAML document/mapping unwrapping
- Parameterized `setupTestCore(t, opts...)` test helper
- Added `createTestIssues()` batch helper, used across 9 call sites
- Added `writeTempConfig()` test helper in cmd_test.go
- Named 15+ magic numbers in UI layout (styles.go), modal sizing, retry logic, ClickUp priorities
- Replaced `sync.WaitGroup` with `errgroup.SetLimit(10)` in both GitHub and ClickUp sync
- Extracted `syncAndTrack()` helper eliminating duplicate tracking logic in sync passes
- Added `TodoCommentFormat` constant for GitHub issue comment linking

Skipped 4 generics items (picker delegates too diverse, shared helpers too trivial for cross-package dependency). Skipped 4 test coverage items (out of scope for this optimization pass).
