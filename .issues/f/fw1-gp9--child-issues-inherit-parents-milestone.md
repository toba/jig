---
# fw1-gp9
title: Child issues inherit parent's milestone
status: completed
type: feature
priority: normal
created_at: 2026-05-27T18:28:00Z
updated_at: 2026-05-27T18:28:00Z
sync:
    github:
        issue_number: "111"
        synced_at: "2026-05-27T18:33:46Z"
---

When an issue is given a parent (on create with --parent, or when reparented via update/TUI parent picker), it now inherits the parent's milestone if the child has no milestone of its own. Explicit milestone choices on the child are never overwritten.

## Summary of Changes

- Added `Resolver.inheritMilestoneFromParent` helper (internal/todo/graph/resolver.go): copies the parent's milestone onto the child only when the child has a parent and no milestone of its own.
- `CreateIssue`: after setting parent, inherits the parent's milestone when none was supplied.
- `UpdateIssue`: after reparenting, inherits the new parent's milestone unless the milestone is being set explicitly in the same update.
- Routes through GraphQL so CLI (`todo create/update`) and TUI parent picker all benefit.
- Added TestMilestoneInheritedFromParent covering inherit-on-create, no-override-of-explicit, and inherit-on-reparent.
