---
# bw7-1ks
title: Rename cite check → cite review; add cite add command
status: in-progress
type: feature
priority: normal
created_at: 2026-02-21T19:21:01Z
updated_at: 2026-02-21T19:21:17Z
---

Two improvements to the cite subcommand:

- [ ] Rename `jig cite check` to `jig cite review` — better verb for the action of reviewing upstream changes
- [ ] Add `jig cite add <git-url>` command that accepts a git URL, inspects the repo (via `gh` or `git`), and suggests the right citation config entry (repo, branch, paths) for the user to confirm and append to `.jig.yaml`
