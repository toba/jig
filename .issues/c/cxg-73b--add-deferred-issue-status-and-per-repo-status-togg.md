---
# cxg-73b
title: Add 'deferred' issue status and per-repo status togglability
status: completed
type: feature
priority: normal
created_at: 2026-05-08T14:17:19Z
updated_at: 2026-05-08T14:36:44Z
sync:
    github:
        issue_number: "106"
        synced_at: "2026-05-08T14:46:46Z"
---

## Motivation

Need a status for issues that raised big concerns and require further consideration before they can be worked on. Not scrapped, not ready, not in-progress — parked pending more thought.

Proposed name: `deferred` (open to alternatives like `parked`, `on-hold`, `blocked-on-decision`).

Note: `deferred` is currently used as a *priority* value. Need to decide whether to:
- Rename the priority (e.g., to `someday`) and free up `deferred` for status, or
- Pick a different status name to avoid the collision.

## Symbol

Tentative: hourglass (⏳ or ⌛). Pick something that renders cleanly in the TUI alongside existing status glyphs.

## Tasks

- [x] Kept `deferred`. Priority and status are addressed by separate flags (`-p` vs `-s`) so there is no data-layer collision.
- [x] Picked ⏸ (pause). Hourglass `⏳` was already in use as the due-date indicator, so reusing it would collide visually.
- [x] Added `StatusDeferred` constant and entry in `DefaultStatuses` between `draft` and `completed`.
- [x] Added icon and render cases in `internal/todo/ui/styles.go`; added `orange` to `NamedColors`.
- [x] Automatic via `DefaultStatusNames`.
- [x] `deferred` → `open` in `DefaultStatusMapping`.
- [x] Prime template emits a DEFERRING workflow bullet and lists only the project's enabled statuses; agent is told to consult that list before picking a status.

## Per-repo status togglability

Allow each repo to enable/disable available statuses via `.jig.yaml`. `draft` and `completed` should always be enabled (non-togglable). All others (`ready`, `in-progress`, `review`, `scrapped`, and the new deferred status) should be configurable.

Sketch:

```yaml
todo:
  statuses:
    review: false      # disable code review status for solo projects
    deferred: true
```

### Tasks

- [x] Schema: `todo.statuses: { <name>: <bool> }`. Unmentioned statuses default to enabled; `false` disables.
- [x] Enforce `ready` and `completed` as always-on (`MandatoryStatuses`). These anchor the open/closed lifecycle. `draft` is now optional.
- [x] `cmd/todo create` and `cmd/todo update` reject disabled statuses with a clear error listing what is enabled.
- [ ] (Follow-up) Bubble Tea filter UI and status-cycle pickers do not yet hide disabled statuses.
- [x] Error: `status "review" is disabled in this project (enabled: in-progress, ready, draft, deferred, completed, scrapped)`.
- [x] Covered by prime-template change above.



## Summary of Changes

**`deferred` status**

- Position: between `draft` and `completed`. Color: orange (added to `NamedColors`). Icon: ⏸.
- Hourglass was rejected because `⏳` is already used as the due-date indicator.
- Naming clash with the `deferred` priority is harmless — `-p` and `-s` are distinct flags.
- GitHub sync: `deferred` → `open`.

**Per-repo togglability**

- `Statuses map[string]bool` field on `Config` (`todo.statuses` in `.jig.yaml`).
- `IsStatusEnabled`, `EnabledStatusNames`, `EnabledStatusList` helpers.
- `MandatoryStatuses = {ready, completed}` — these cannot be disabled. Draft is now optional like the rest.
- `cmd/todo create` and `cmd/todo update` reject disabled statuses; existing issues already in a now-disabled status remain visible and editable, only new transitions are blocked.

**Removed implicit github auto-disable of `review`**

- The prime template previously branched on `HasGitHubSync` to advise the agent to skip `review`.
- Replaced with explicit branching on `ReviewEnabled` / `DeferredEnabled`, driven by `todo.statuses`.
- Prime template lists only the project's enabled statuses and tells the agent to consult that list before picking a status (per user request).
- This repo's own `.jig.yaml` now sets `statuses.review: false` explicitly to preserve its prior advice.

**Files touched**

- `internal/todo/config/config.go` — `StatusDeferred`, `MandatoryStatuses`, `Statuses` field, `IsStatusEnabled`, `EnabledStatusNames`, `EnabledStatusList`
- `internal/todo/config/config_test.go` — count bumped to 7, deferred added, new `TestStatusToggle` suite
- `internal/todo/ui/styles.go` — icon, render cases, `orange` in `NamedColors`
- `internal/todo/integration/github/config.go` — `deferred → open`
- `internal/todo/integration/github/config_test.go` — deferred added
- `internal/todo/integration/clickup_adapter_test.go` — deferred in the full-coverage fixture
- `cmd/todo_create.go`, `cmd/todo_update.go` — togglability validation
- `cmd/prime.go` — `ReviewEnabled`/`DeferredEnabled` flags, statuses filtered to enabled
- `cmd/todo_prompt.tmpl` — workflow advice rewritten; explicit note that the agent must consult the listed statuses
- `.jig.yaml` — explicit `statuses.review: false`

**Deferred follow-up**

- Bubble Tea TUI filter UI and status-cycle pickers do not yet hide disabled statuses.
