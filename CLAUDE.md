# skill

Multi-tool CLI combining upstream repo monitoring and Claude Code security guard.

## Build & Test

```bash
go build -o skill .
go test ./...
go vet ./...
```

## Architecture

- `cmd/` — Cobra commands
  - `upstream` parent with `init`, `check`, `mark` subcommands — upstream monitoring
  - `nope` parent with `init`, `doctor`, `help` subcommands — security guard
  - `brew` parent with `init`, `doctor` subcommands — Homebrew tap management
  - `zed` parent with `init`, `doctor` subcommands — Zed extension management
  - `update`, `version` — top-level utilities
- `internal/config/` — `.toba.yaml` partial read/write via yaml.v3 Node API (upstream section)
- `internal/github/` — GitHub API client wrapping `gh` CLI
- `internal/classify/` — Glob-based file classification (high/medium/low)
- `internal/display/` — Lipgloss-styled terminal output
- `internal/nope/` — PreToolUse guard (reads `nope:` section from `.toba.yaml`)
- `internal/brew/` — Homebrew tap init and doctor logic
- `internal/zed/` — Zed extension init and doctor logic

## Key Design Decisions

- Config uses yaml.v3 Node API for partial read/write to avoid clobbering other sections in `.toba.yaml`
- GitHub calls shell out to `gh` CLI (no API token management needed)
- `check` is strictly read-only; `mark` is the explicit write step
- Uses `doublestar` for `**` glob support since Go's `path.Match` lacks it
- `nope` guard reads rules from `nope:` key in `.toba.yaml` (not a separate file)
- `nope` uses instance-based `DebugLogger` (nil-safe) instead of global state
- Guard mode runs via `RunE` on the parent cobra command; exit codes use `ExitError` sentinel
- Each command group (`upstream`, `nope`, `brew`, `zed`) has its own `PersistentPreRunE`; root's is a no-op
- `nope init` writes to `.toba.yaml` and `.claude/settings.json`; migrates old `nogo` hooks to `skill nope`
