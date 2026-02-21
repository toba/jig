# jig

Multi-tool CLI combining upstream repo monitoring, file-based issue tracking, and Claude Code security guard.

## Build & Test

```bash
go build -o jig .
go test ./...
go vet ./...
```

## Architecture

- `cmd/` — Cobra commands
  - `todo` parent with `init`, `create`, `list`, `show`, `update`, `delete`, `archive`, `roadmap`, `graphql`, `check`, `sync`, `refry`, `tui` subcommands — issue tracking
  - `prime` — output instructions for AI coding agents
  - `tui` — top-level alias for `todo tui`
  - `sync` — top-level alias for `todo sync` (with `check`, `link`, `unlink` subcommands)
  - `upstream` parent with `init`, `check`, `mark` subcommands — upstream monitoring
  - `nope` parent with `init`, `doctor`, `help` subcommands — security guard
  - `brew` parent with `init`, `doctor` subcommands — Homebrew tap management
  - `zed` parent with `init`, `doctor` subcommands — Zed extension management
  - `update`, `version` — top-level utilities
- `internal/config/` — `.jig.yaml` partial read/write via yaml.v3 Node API (upstream section)
- `internal/github/` — GitHub API client wrapping `gh` CLI
- `internal/classify/` — Glob-based file classification (high/medium/low)
- `internal/display/` — Lipgloss-styled terminal output
- `internal/nope/` — PreToolUse guard (reads `nope:` section from `.jig.yaml`)
- `internal/brew/` — Homebrew tap init and doctor logic
- `internal/zed/` — Zed extension init and doctor logic
- `internal/todo/config/` — todo config (reads `todo:` section from `.jig.yaml`, Node API for partial writes)
- `internal/todo/core/` — issue CRUD, archive, link checking, file watcher
- `internal/todo/graph/` — GraphQL schema and resolvers (gqlgen)
- `internal/todo/integration/` — sync integrations (ClickUp, GitHub Issues)
- `internal/todo/issue/` — issue model, frontmatter parsing, sorting
- `internal/todo/output/` — JSON output helpers
- `internal/todo/refry/` — migration from hmans/beans format
- `internal/todo/search/` — Bleve full-text search index
- `internal/todo/tui/` — Bubble Tea interactive TUI
- `internal/todo/ui/` — Lipgloss styles, tree rendering
- `pkg/client/` — GraphQL client library

## Key Design Decisions

- Config uses yaml.v3 Node API for partial read/write to avoid clobbering other sections in `.jig.yaml`
- GitHub calls shell out to `gh` CLI (no API token management needed)
- `check` is strictly read-only; `mark` is the explicit write step
- Uses `doublestar` for `**` glob support since Go's `path.Match` lacks it
- `nope` guard reads rules from `nope:` key in `.jig.yaml` (not a separate file)
- `nope` uses instance-based `DebugLogger` (nil-safe) instead of global state
- Guard mode runs via `RunE` on the parent cobra command; exit codes use `ExitError` sentinel
- Each command group (`upstream`, `nope`, `brew`, `zed`, `todo`) has its own `PersistentPreRunE`; root's is a no-op
- `nope init` writes to `.jig.yaml` and `.claude/settings.json`; migrates old `nogo`/`skill nope`/`ja nope` hooks to `jig nope`
- `todo` config uses yaml.v3 Node API for `Save()` to avoid clobbering other `.jig.yaml` sections
- `todo` stores issues as markdown files with YAML frontmatter in `.issues/`
- `todo` supports GraphQL queries/mutations via embedded gqlgen schema
- `tui` and `sync` have top-level aliases that call `initTodoCore()` in their own PreRunE
