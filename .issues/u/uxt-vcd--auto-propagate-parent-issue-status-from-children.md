---
# uxt-vcd
title: Auto-propagate parent issue status from children
status: completed
type: feature
priority: normal
created_at: 2026-03-10T17:05:25Z
updated_at: 2026-03-10T17:07:32Z
sync:
    github:
        issue_number: "91"
        synced_at: "2026-03-10T17:09:46Z"
---

When child issues change status, parent issues should automatically update to reflect the aggregate state. Adds propagateStatusLocked to core that runs after Update().

## Tasks
- [x] Create propagate.go with findChildrenLocked, computeParentStatus, propagateStatusLocked
- [x] Add propagateStatusLocked call to Update() in core.go
- [x] Create propagate_test.go with comprehensive tests
- [x] Verify all tests pass and lint is clean


## Summary of Changes

Added automatic parent status propagation in `internal/todo/core/propagate.go`. When a child issue is updated via `Update()`, the parent's status is recomputed from all children and updated if a rule matches, then recursion continues up the hierarchy.
