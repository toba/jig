---
# l8v-n8t
title: 'graphql command: add -f/--file flag to avoid shell escaping issues'
status: completed
type: bug
priority: normal
created_at: 2026-03-09T00:42:57Z
updated_at: 2026-03-09T00:48:13Z
sync:
    github:
        issue_number: "88"
        synced_at: "2026-03-09T00:49:14Z"
---

Agents frequently break when passing GraphQL mutations as shell arguments because backticks in markdown content get interpreted by zsh as command substitution.

Example error:
```
(eval):1: command not found: moveCursor
Error: accepts at most 1 argument (the GraphQL query)
```

The command already supports stdin, but agents often construct the query inline. Adding a `-f`/`--file` flag to read the query from a file completely avoids shell escaping.

## Tasks

- [x] Add `-f`/`--file` flag to `graphql` command
- [x] Write test for file-based query input
- [x] Update help text and examples


## Summary of Changes

Added `-f`/`--file` flag to `jig todo query` command so agents can write GraphQL queries to a temp file and avoid zsh backtick/shell escaping issues. Updated agent prompt template to recommend `-f` for mutations.
