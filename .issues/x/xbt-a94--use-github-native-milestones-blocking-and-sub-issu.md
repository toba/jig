---
# xbt-a94
title: Use GitHub native milestones, blocking, and sub-issues APIs
status: completed
type: feature
priority: normal
created_at: 2026-02-22T18:25:36Z
updated_at: 2026-02-23T16:11:24Z
sync:
    github:
        issue_number: "12"
        synced_at: "2026-02-23T17:08:13Z"
---

GitHub now supports milestones, blocking/blocked-by relationships, and sub-issues (parent/child) natively. The sync should use all three native APIs instead of emulating relationships via footer text links.

## Current State

- **Parent/child**: Uses sub-issues API ✓ (already native)
- **Blocking/blocked-by**: Emulated via footer text links (`**Blocks:** #123`, `**Blocked by:** #456`) — not native
- **Milestones**: Not used at all. Milestone-type issues are synced as regular GitHub issues with type "Task"
- **Footer links**: Written for all 4 relationship types (parent, children, blocks, blocked-by) in Pass 3 of sync

## Tasks

- [ ] Use native blocking/blocked-by API instead of footer text links
  - `POST /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by` with `{"issue_id": N}` to add
  - `DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by/{issue_id}` to remove
  - `GET .../dependencies/blocked_by` and `.../dependencies/blocking` to read current state
  - Add `AddBlockedBy`, `RemoveBlockedBy`, `ListBlockedBy`, `ListBlocking` client methods
  - Replace footer-based blocking sync with native API calls in `syncRelationships()`
- [ ] Sync milestone-type issues as GitHub milestones
  - `POST /repos/{owner}/{repo}/milestones` to create milestones (title, description, due_on, state)
  - Assign child issues to the milestone via `milestone` field on create/update issue requests
  - Store milestone number in sync metadata (new `milestone_number` sync key)
  - Map milestone status to GitHub milestone state (open/closed)
  - Add `CreateMilestone`, `UpdateMilestone`, `ListMilestones`, `GetMilestone` client methods
  - Update `CreateIssueRequest` and `UpdateIssueRequest` to include `Milestone *int` field
  - Milestone-type issues should create a GitHub milestone AND a GitHub issue (the issue serves as a container for sub-issues; the milestone provides the native tracking)
- [ ] Remove redundant footer links for relationships that are now native
  - Parent/child footer links are redundant since sub-issues API is used
  - Blocking/blocked-by footer links become redundant once native API is used
  - Remove `syncRelationships()` Pass 3 entirely, or repurpose for edge cases (e.g. cross-repo references)
  - Remove `stripRelationshipLines()` and `relationshipPrefixes`
- [ ] Update type mapping so milestone/epic types are not mapped to "Task"
  - Milestone type should probably not be a regular issue type at all (it's a milestone)
  - Epic could remain as-is or map to a custom type if GitHub adds one
- [ ] Handle sync state for new metadata
  - Track `milestone_number` alongside `issue_number` in sync state
  - Handle milestone number lookups during child issue sync (need milestone number to set on child issues)
- [ ] Update tests for all changes
  - Test native blocking API calls (add, remove, idempotent re-add)
  - Test milestone create/update/assign lifecycle
  - Test that footer links are no longer written
  - Test milestone status mapping (completed → closed, etc.)

## API Reference

### Blocking Dependencies (GA since Aug 2025)
- `POST /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by` — body: `{"issue_id": int}`
- `DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by/{issue_id}`
- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by`
- `GET /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocking`
- Limit: 50 issues per relationship

### Milestones
- `POST /repos/{owner}/{repo}/milestones` — body: `{"title": str, "description": str, "due_on": timestamp, "state": "open"|"closed"}`
- `PATCH /repos/{owner}/{repo}/milestones/{milestone_number}`
- `GET /repos/{owner}/{repo}/milestones`
- Assign via `milestone` field (int, milestone number) on issue create/update

### Sub-Issues (already implemented)
- `POST /repos/{owner}/{repo}/issues/{issue_number}/sub_issues` — body: `{"sub_issue_id": int}`
- `DELETE /repos/{owner}/{repo}/issues/{issue_number}/sub_issue`
