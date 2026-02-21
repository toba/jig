---
# vdt-69n
title: Add safety level tiers for nope starter rules
status: ready
type: feature
created_at: 2026-02-21T19:50:02Z
updated_at: 2026-02-21T19:50:02Z
---

nope currently has a flat list of starter rules. Add tiered presets so `jig nope init --level <tier>` generates appropriate rule sets:

- **critical**: Only the most destructive operations (rm -rf /, fork bombs, dd to disk)
- **standard** (default): Current starter rules
- **strict**: Adds more restrictive rules (sudo rm, DROP DATABASE, etc.)

Reference: https://github.com/leegonzales/claude-guardrails (critical/high/strict tiers)

- [ ] Define rule tiers
- [ ] Add `--level` flag to `jig nope init`
- [ ] Update StarterConfig to support tier selection
- [ ] Document tier differences
