# upstream

CLI tool to monitor upstream repositories for changes, classify files by relevance, and track review state.

## Build & Test

```bash
go build -o upstream .
go test ./...
go vet ./...
```

## Architecture

- `cmd/` — Cobra commands (check, mark, init, version)
- `internal/config/` — `.toba.yaml` partial read/write via yaml.v3 Node API
- `internal/github/` — GitHub API client wrapping `gh` CLI
- `internal/classify/` — Glob-based file classification (high/medium/low)
- `internal/display/` — Lipgloss-styled terminal output

## Key Design Decisions

- Config uses yaml.v3 Node API for partial read/write to avoid clobbering other sections in `.toba.yaml`
- GitHub calls shell out to `gh` CLI (no API token management needed)
- `check` is strictly read-only; `mark` is the explicit write step
- Uses `doublestar` for `**` glob support since Go's `path.Match` lacks it
