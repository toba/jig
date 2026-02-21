---
# bc1-ews
title: 'Rename upstream → cite / citations:'
status: completed
type: task
priority: normal
created_at: 2026-02-21T19:02:07Z
updated_at: 2026-02-21T19:06:56Z
---

Rename the upstream subcommand to cite (verb) and the YAML config key to citations: (noun, plural). Update all cmd/, internal/config/, internal/update/, internal/nope/ test fixtures, internal/display/, CLAUDE.md, README.md, and schema.json.

## Summary of Changes\n\n- Renamed `upstream` subcommand to `cite` (cmd/upstream.go → cmd/cite.go)\n- Renamed YAML config key `upstream:` to `citations:`\n- Renamed internal types: upstreamConfig → citationConfig, upstreamSource → citationSource, upstreamPath → citationPath\n- Renamed migrateUpstreamSkill → migrateCiteSkill\n- Added migrateUpstreamKey() migration to rename existing `upstream:` keys to `citations:` in .jig.yaml\n- Updated all test fixtures, schema.json, CLAUDE.md, README.md\n- All tests pass, build succeeds, go vet clean
