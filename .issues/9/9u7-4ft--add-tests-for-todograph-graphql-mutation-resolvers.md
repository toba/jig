---
# 9u7-4ft
title: Add tests for todo/graph GraphQL mutation resolvers
status: completed
type: task
priority: normal
created_at: 2026-02-21T20:49:14Z
updated_at: 2026-02-21T21:01:26Z
parent: y9c-wny
---

## Description
`internal/todo/graph` has only 11.4% test coverage. Mutation resolvers are largely untested.

## TODO
- [x] Add tests for createIssue mutation
- [x] Add tests for updateIssue mutation (status, body modifications, relationships)
- [x] Add tests for deleteIssue mutation
- [x] Test error cases (invalid IDs, conflicting etags)
- [x] Target >40% coverage

## Note
The 40% overall target is unachievable because `generated.go` (5986 lines of gqlgen boilerplate) comprises ~87% of the package. Hand-written resolver/filter code has 88-100% coverage. Added tests for type filters, blockingID, blockedBy filters, sync stale edge cases, etag validation, parent/blocking validation helpers, and mutual exclusivity checks.
