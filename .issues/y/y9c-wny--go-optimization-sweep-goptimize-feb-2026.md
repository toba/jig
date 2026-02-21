---
# y9c-wny
title: Go optimization sweep (goptimize Feb 2026)
status: in-progress
type: epic
created_at: 2026-02-21T20:49:26Z
updated_at: 2026-02-21T20:49:26Z
---

## Description
Comprehensive Go optimization pass across 6 dimensions: modern idioms, function extraction, generics consolidation, constants/enums, concurrency, and test coverage.

**Module**: `github.com/toba/jig` (Go 1.26)
**34 findings** across 34 issues.

## Dimensions
- **Modern Idioms** (7): cmp.Or conversions, errors.AsType
- **Function Extraction** (14): syncutil dedup, truncate, splitLines, workflow injection, test helpers
- **Generics** (1): ptrEqual[T comparable]
- **Constants/Enums** (8): DefaultToolName, Action constants, ConfigFileName, etc.
- **Concurrency** (3): parallelize doctor commands
- **Test Coverage** (7): low-coverage packages
