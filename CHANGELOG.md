# Changelog

## Week of Feb 23 – Mar 1, 2026

### ✨ Features

- Add `changelog` command for gathering recent issues and commits by time range ([#69](https://github.com/toba/jig/issues/69))
- Add Scoop bucket companion support; init, doctor, and CI workflow generation for Windows distribution
- Add `review` status for code-complete issues awaiting evaluation ([#65](https://github.com/toba/jig/issues/65))
- Color due date hourglass by urgency; red ≤24h, orange ≤3d, yellow ≤7d, green beyond
- Add status and priority sort options with newest-created tiebreaker

### 🐞 Fixes

- Fix TUI filter breaking tree hierarchy; preserve ancestor chain when filtering ([#64](https://github.com/toba/jig/issues/64))
- Fix TUI layout; use TypeAbbrev for dimmed rows to prevent lipgloss word-wrap ([#68](https://github.com/toba/jig/issues/68))
- Fix `jig commit` leaving dirty files after sync metadata updates ([#70](https://github.com/toba/jig/issues/70))
- Fix release workflow; add GoReleaser replace mode, parallelize scoop job

### 🗜️ Tweaks

- Fix all golangci-lint issues; add config, fix syntax errors, suppress test false positives ([#67](https://github.com/toba/jig/issues/67))
- Modernize Go idioms; extract helpers, add constants, bound sync concurrency ([#66](https://github.com/toba/jig/issues/66))
- Shorten TUI type column to two-letter abbreviations

## Week of Feb 16 – Feb 22, 2026

### ✨ Features

- Add data exfiltration detection; sensitive file uploads over network ([#14](https://github.com/toba/jig/issues/14))
- Add environment self-defense built-in check ([#27](https://github.com/toba/jig/issues/27))
- Add inline secret detection built-in ([#19](https://github.com/toba/jig/issues/19))
- Detect `$var` in command position as evasion ([#16](https://github.com/toba/jig/issues/16))
- Fail closed on malformed stdin ([#11](https://github.com/toba/jig/issues/11))
- Strip wrapper commands (`sudo`, `timeout`, `env`, etc.) in CheckNetwork ([#9](https://github.com/toba/jig/issues/9))
- Split compound commands into segments for independent rule checking ([#26](https://github.com/toba/jig/issues/26))
- Rename `cite check` → `cite review`; add `cite add` command ([#10](https://github.com/toba/jig/issues/10))
- Add `cite doctor` subcommand to verify license attribution ([#29](https://github.com/toba/jig/issues/29))
- Brew doctor; detect project language and adjust diagnostics ([#3](https://github.com/toba/jig/issues/3))
- Enhance GitHub sync to fully preserve issue relationships; parent/child via sub-issues API, footer links for blocks/blocked-by ([#20](https://github.com/toba/jig/issues/20))
- Sync milestones and blocking natively; replace footer links with GitHub milestones API and dependencies API ([#12](https://github.com/toba/jig/issues/12))
- Add tag registry; import GitHub labels as project tags with relaxed validation
- Add file-based issue tracking (todo) with Go idiom modernization
- Add top-level `jig init` command to run all sub-inits in sequence
- Upload local images during sync
- TUI auto-refresh when issues change on disk
- Add sync footer note to externally created issues

### 🐞 Fixes

- Fix brew doctor false positive on workflow asset reference check ([#30](https://github.com/toba/jig/issues/30))
- Fix commit push; push tags in version order, allow push-only without staged changes ([#17](https://github.com/toba/jig/issues/17), [#28](https://github.com/toba/jig/issues/28))
- Fix sub-issue sync; pass GitHub issue ID instead of number to sub-issues API
- Fix GitHub sync 422 on milestone clear; serialize as null instead of 0
- Fix TUI detail view selection resets on file watcher refresh
- Fix sync to update parent/subtask relationships on existing ClickUp tasks
- Fix ClickUp sync not setting parent on tasks whose parent isn't in the sync batch

### 🗜️ Tweaks

- Rename project; skill/ja → jig ([#5](https://github.com/toba/jig/issues/5))
- Rename config file; .toba.yaml → .jig.yaml ([#18](https://github.com/toba/jig/issues/18))
- Rename upstream → cite with flattened config ([#24](https://github.com/toba/jig/issues/24))
- Call todo sync in-process instead of shelling out to subprocess ([#15](https://github.com/toba/jig/issues/15))
- Go optimization sweep; extract shared utilities, add constants, parallelize doctor, expand test coverage across ~30 sub-tasks ([#7](https://github.com/toba/jig/issues/7))
- Skip brew/zed doctor gracefully when companions not configured
- Add all jig tools to `prime` command output
- Simplify GraphQL schema for agentic use
- Apply all goptimize findings

## Week of Feb 9 – Feb 15, 2026

### ✨ Features

- Implement `migrate` subcommand for importing from beans format
- Integrate ClickUp sync into todo
- Integrate GitHub sync into todo
- Import ClickUp config during migration
- Remove configurable prefix; adopt fixed xxx-xxx ID format with hash subfolders
- Add due field and due-date sorting
- Support OS default app for markdown editing in TUI
- Provide a public Go client package for external tools

### 🐞 Fixes

- Fix GitHub sync; use native types, remove label abuse
- Fix concurrent map write crash in GitHub sync

### 🗜️ Tweaks

- Optimize codebase; update to Go 1.26 and apply goptimize analysis
- Update prime command prompt template with all fork features
- Show status icon and label in status picker
- Display blockedBy relationships in show output
- Cherry-pick upstream atomic relationship updates

## Week of Feb 2 – Feb 8, 2026

### ✨ Features

- Add deep search; `//` to filter by title and body
- Add sort picker to TUI
- Replace fuzzy filter with substring filter in TUI
- Add external integration metadata to issues
- Support editor config field from config file

### 🐞 Fixes

- Fix deep search pointer invalidation bug

### 🗜️ Tweaks

- Configure GoReleaser for toba/todo fork
- Add build number to help modal and version output