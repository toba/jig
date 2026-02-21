---
# kfr-zhl
title: Add tests for internal/github client wrapper
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:14Z
updated_at: 2026-02-21T20:59:03Z
parent: y9c-wny
---

## Description
`internal/github` has only 11.7% test coverage. The GitHub API client functions are mostly untested.

## TODO
- [ ] Add tests for repo operations (create, view, check existence)
- [ ] Add tests for API content fetching
- [ ] Add tests for error handling (gh CLI not installed, auth failures)
- [ ] Target >40% coverage
