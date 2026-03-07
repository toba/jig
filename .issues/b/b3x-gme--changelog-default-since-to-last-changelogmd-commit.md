---
# b3x-gme
title: 'changelog: default --since to last CHANGELOG.md commit'
status: completed
type: bug
priority: normal
created_at: 2026-03-07T22:01:00Z
updated_at: 2026-03-07T22:03:38Z
sync:
    github:
        issue_number: "85"
        synced_at: "2026-03-07T22:19:48Z"
---

- [x] Add ChangelogLastModified() to get last git commit date for CHANGELOG.md
- [ ] Use it as default --since when no date flags provided
- [ ] Fall back to 7 days if file untracked or missing
- [ ] Write tests
- [ ] Run lint
