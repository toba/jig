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
  - `check`, `mark`, `init` — upstream monitoring
  - `nope`, `nope init`, `nope doctor`, `nope help` — security guard
  - `version`
- `internal/config/` — `.toba.yaml` partial read/write via yaml.v3 Node API (upstream section)
- `internal/github/` — GitHub API client wrapping `gh` CLI
- `internal/classify/` — Glob-based file classification (high/medium/low)
- `internal/display/` — Lipgloss-styled terminal output
- `internal/nope/` — PreToolUse guard (reads `nope:` section from `.toba.yaml`)

## Key Design Decisions

- Config uses yaml.v3 Node API for partial read/write to avoid clobbering other sections in `.toba.yaml`
- GitHub calls shell out to `gh` CLI (no API token management needed)
- `check` is strictly read-only; `mark` is the explicit write step
- Uses `doublestar` for `**` glob support since Go's `path.Match` lacks it
- `nope` guard reads rules from `nope:` key in `.toba.yaml` (not a separate file)
- `nope` uses instance-based `DebugLogger` (nil-safe) instead of global state
- Guard mode runs via `RunE` on the parent cobra command; exit codes use `ExitError` sentinel
- `PersistentPreRunE` skips upstream config loading for `nope` and its subcommands
- `nope init` writes to `.toba.yaml` and `.claude/settings.json`; migrates old `nogo` hooks to `skill nope`
