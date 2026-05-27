---
# o3j-7qz
title: Add lightweight milestone field to todos for grouping future work
status: completed
type: feature
priority: normal
created_at: 2026-05-27T16:17:01Z
updated_at: 2026-05-27T17:19:52Z
sync:
    github:
        issue_number: "108"
        synced_at: "2026-05-27T17:21:34Z"
---

Make "milestone" a first-class, file-backed concept (like a GitHub milestone) instead of an issue type. A milestone has a name, a 2–3 char short name (for the TUI grid), an optional description, and a due date. Any issue may optionally be assigned to one milestone.

## Locked design decisions

- **Storage**: milestone files live in `.issues/milestones/` (CLI-created; the issue loader skips that folder). Milestones are NOT issues and NOT an issue type.
- **Reference**: an issue references its milestone by the milestone's **id** (NanoID in filename), consistent with `parent`/`blocking`. Short name is display-only.
- **Legacy type**: migrate `type: milestone` issues into milestone files + reassign children, then retire the `milestone` issue type.
- **GitHub sync**: rework to map the new milestone entity ↔ GitHub milestone.
- **Ordering**: milestone order = due date (then name).

## Requirements

- [x] Milestone entity + `.issues/milestones/` storage; loader skips the subdir
- [x] `issue.Milestone` field (frontmatter, parse/render)
- [x] `jig todo milestone` CLI (create/list/show/update/delete/migrate)
- [x] `--milestone` on `todo create`/`update`; `--milestone`/`--no-milestone`/`--sort milestone` on `list`
- [x] GraphQL: milestone field + filters, plus Milestone type + `milestones`/`milestone` queries + create/update/delete mutations (regen)
- [x] TUI: short-name badge, `m` picker (single+multi), `g m` filter, detail badge, `C` create-interstitial (Issue/Milestone chooser + milestone create form), help
- [x] GitHub sync rework onto the new entity
- [x] Migration retiring the legacy type; run against ../thesis (Modernization → milestone [mod] + 4 children)
- [x] Tests (TDD) and lint clean

Full plan: ~/.claude/plans/work-on-o3j-7qz-dazzling-wind.md


## Summary of Changes

Milestones are now a first-class, file-backed concept (not an issue type).

**Data model & storage** — `issue.Milestone` entity (id/short/name/due/description/sync) stored as markdown in `.issues/milestones/`; the issue loader skips that dir. Issues reference a milestone by ID via a new `milestone:` frontmatter field. Core gains milestone CRUD + `MilestonesSorted`/`MilestoneOrder`.

**CLI** — `jig todo milestone` (create/list/show/update/delete/migrate). `--milestone` on `todo create`/`update`; `--milestone`/`--no-milestone`/`--sort milestone` on `todo list`.

**GraphQL** — `milestone` field on Issue + create/update inputs; `milestone`/`excludeMilestone` filters (gqlgen regenerated); resolver + filter wiring; `SortByMilestone`.

**TUI** — short-name badge column in the list grid; `m` opens a milestone picker to (re)assign single or multi-selected issues (mirrors `s`); `g m` filters the list to a milestone (reuses matched/dimmed tree rendering); milestone badge in the detail header; help entries.

**GitHub sync** — reworked onto the new entity: Pass 0 syncs milestone entities ↔ GitHub milestones, storing `milestone_number` on the milestone file's `sync.github`; issue assignment resolves via `issue.Milestone`.

**Migration / retire type** — `jig todo milestone migrate` (idempotent, `--dry-run`) converts legacy `type: milestone` issues into entities, reassigns their children, carries over the GitHub milestone number, and deletes the old issue. `milestone` removed from `DefaultTypes` so it can't be created anew (constant/hierarchy refs retained for legacy compatibility).

All tests pass; `go vet` and `scripts/lint.sh` clean.

## Deferred Follow-ups

- **TUI `C` create-interstitial + in-TUI milestone creation form**: the `C` key still creates issues only; creating a milestone is via CLI. (User noted agents create via CLI, so friction is acceptable.)
- **Run the migration on `../thesis`**: verified via `--dry-run` (`2cj-tzj` "Modernization" → `[mod]` + 4 children). The actual run modifies the thesis repo and was left for explicit user execution.

## Update: deferrals resolved

The three items originally deferred are now complete: the full GraphQL milestone surface (type/queries/mutations), the TUI `C` create-interstitial with an in-TUI milestone creation form, and the actual migration run on `../thesis` (verified: `2cj-tzj` Modernization → milestone `fqd-rn0` `[mod]`, GitHub milestone #7 carried over, 4 children reassigned, old issue deleted).
