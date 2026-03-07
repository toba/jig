---
# 22t-bt8
title: Add scope field to citation sources
status: completed
type: feature
priority: normal
created_at: 2026-03-07T18:10:00Z
updated_at: 2026-03-07T18:10:55Z
sync:
    github:
        issue_number: "81"
        synced_at: "2026-03-07T18:40:10Z"
---

Add optional freeform `scope` field to citation sources so agents know which local area a citation pertains to.

- [x] Add `Scope` to `Source` struct in config.go
- [x] Include `scope` in `FormatSourceYAML` in add.go
- [x] Include `scope` in `cite review` text output in display.go
- [x] Update config_test.go with scope in test fixture
- [x] Update add_test.go FormatSourceYAML test
- [x] Add scope values to .jig.yaml citations


## Summary of Changes

Added optional `scope` field to citation sources across struct, YAML formatting, display output, tests, and project config.
