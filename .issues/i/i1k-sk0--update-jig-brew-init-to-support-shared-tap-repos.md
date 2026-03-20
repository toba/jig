---
# i1k-sk0
title: Update jig brew init to support shared tap repos
status: completed
type: feature
priority: normal
created_at: 2026-03-20T19:42:09Z
updated_at: 2026-03-20T19:43:05Z
sync:
    github:
        issue_number: "96"
        synced_at: "2026-03-20T20:29:44Z"
---

\`jig brew init\` currently scaffolds a per-project tap repo (e.g. \`homebrew-musup\`). Now that toba uses a single shared \`homebrew-tap\` repo, the command should support pushing formulae to an existing shared tap instead of creating a new per-project repo.

## Requirements

- [x] \`jig brew init --tap toba/homebrew-tap\` adds/updates a formula in an existing shared tap repo
- [x] Shared tap is the only path; convention defaults to \`owner/homebrew-tap\`
- [x] \`jig brew doctor\` derives tool name from source repo, works with shared taps
- [x] \`brew init\` auto-saves \`companions.brew\` to \`.jig.yaml\`

## Summary of Changes

- Removed per-project tap creation (\`createTapRepo\`, \`pushInitialContent\`, \`generateReadme\`)
- \`brew init\` now clones existing shared tap, adds/updates only the formula file, commits and pushes
- Workflow job generation uses explicit \`Tap\` field for clone URL instead of deriving \`homebrew-<tool>\`
- Convention fallback changed from \`owner/homebrew-<tool>\` to \`owner/homebrew-tap\`
- \`brew doctor\` derives tool name from source repo (not tap repo name)
- \`brew init\` auto-saves \`companions.brew\` to \`.jig.yaml\` after successful setup
- CI workflow commit message changed to \`bump <tool> to\` for shared tap clarity
