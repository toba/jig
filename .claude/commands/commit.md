---
description: Stage all changes and commit with a descriptive message
---

## Active Codebase Expectations

This is an active codebase with multiple agents and people making changes concurrently. Do NOT waste time investigating unexpected git status:
- If a file you edited shows no changes, someone else likely already committed it - move on
- If files you didn't touch appear modified, another agent may have changed them - include or exclude as appropriate
- Focus on what IS changed, not what ISN'T

## Phase 1: Gather

Run `ja commit gather`

### If exit code 2 (gitignore candidates found)

Ask the user whether to:
1. Add the files to .gitignore
2. Proceed with committing them anyway
3. Cancel

### If successful

You'll receive structured output with these sections:
- **STAGED** — files that will be committed
- **DIFF** — full staged diff
- **LATEST_TAG** — current version tag (or "(none)")
- **LOG_SINCE_TAG** — commits since the last tag

Analyze these to determine:
1. A concise commit message (lowercase imperative, focus on "why", include affected issue IDs)
2. Whether a version bump is needed (see below)

## Phase 2: Apply

Run `ja commit apply -m "<message>" [-v <version>] [--push]`

- Always include `-m` with your commit message
- Include `-v <version>` when pushing (see Version Bumps)
- Include `--push` if the user said "push" in their prompt, i.e. $ARGUMENTS contains "push"

## Version Bumps

Every push includes a version bump. Use the LATEST_TAG from gather output and increment:

- **patch**: Bug fixes, docs, refactors, tests — no behavior change
- **minor**: New features, non-breaking additions, breaking changes while pre-1.0
- **major**: Breaking changes (post-1.0 only)

Tag format: `v<major>.<minor>.<patch>` (e.g., v1.2.3)
