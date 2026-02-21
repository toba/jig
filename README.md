# ja

An agent multi-tool for little things.

## Commands

- **`ja`**
   - **`commit`**: stage changes, check for gitignore candidates, signal push intent
   - **`doctor`**: run all doctor checks (brew, zed, nope)
   - **`update`**: migrate legacy config files into `.toba.yaml`
   - **`version`**: print version info
   - **`upstream`**: monitor upstream repositories for changes
      - **`init`**: add starter upstream section to `.toba.yaml`
      - **`check`**: fetch and display changes grouped by relevance
      - **`mark`**: update `last_checked_sha` to current HEAD for a source
   - **`nope`**: Claude Code `PreToolUse` guard (reads JSON from stdin, exits 0 or 2)
      - **`init`**: scaffold nope rules in `.toba.yaml` and hook in `.claude/settings.json`
      - **`doctor`**: validate nope configuration
      - **`help`**: show nope guard reference
   - **`brew`**: Homebrew tap management
      - **`init`**: create tap repo, push initial formula, inject `update-homebrew` CI job
      - **`doctor`**: verify brew tap setup is healthy
   - **`zed`**: Zed extension management
      - **`init`**: create extension repo, push scaffold, inject `sync-extension` CI job
      - **`doctor`**: verify Zed extension setup is healthy

## Install

```bash
brew install toba/ja/ja
```

Or build from source:

```bash
go install github.com/toba/skill@latest
```

## Upstream Monitoring

Track what's changed in repos you care about. Read-only by default — `check` looks, `mark` remembers.

```bash
ja upstream init
ja upstream check
ja upstream mark owner/repo
```

Configure which files matter in `.toba.yaml`:

```yaml
upstream:
  sources:
    - repo: owner/repo
      branch: main
      relationship: derived
      paths:
        high:
          - "src/**/*.go"
        medium:
          - "go.mod"
        low:
          - "README.md"
```

Files are classified as high, medium, or low relevance based on glob patterns. `**` works — we use [doublestar](https://github.com/bmatcuk/doublestar) because Go's `path.Match` stubbornly refuses to support it.

## Nope Guard

A `PreToolUse` hook for Claude Code. Rules live in the `nope:` section of `.toba.yaml` — regex patterns and built-in checks that block tool calls before they execute.

```bash
ja nope init
```

This adds a `nope:` section to `.toba.yaml` with starter rules and wires up the hook in `.claude/settings.json`. Claude Code pipes a JSON payload to `ja nope` on stdin before each tool call. If a rule matches, the tool is blocked (exit 2). If nothing matches, it's allowed (exit 0).

### Structure

The `nope:` section contains a `rules` list and an optional `debug` log path:

```yaml
nope:
  debug: .claude/nope.log   # optional JSONL debug log (omit to disable)
  rules:
    # Regex pattern — matched against the tool_input JSON
    - name: git-push
      pattern: 'git\s+push'
      message: "git push not allowed — only user should push"

    # Built-in check — structural analysis, not just string matching
    - name: pipe-commands
      builtin: pipe
      message: "piped commands not allowed — run commands separately"

    # Scope rules to specific tools (default is Bash only)
    - name: no-write-env
      pattern: '"file_path"\s*:\s*"[^"]*\.env"'
      tools: ["Write", "Edit"]
      message: "writing to .env files not allowed"
```

Rules are either regex patterns or built-in checks. Each rule has a `name`, a `message`, and either a `pattern` (regex) or `builtin` (structural check). The optional `tools` array scopes which tools the rule applies to (defaults to `["Bash"]`; use `["*"]` for all tools).

### Built-in Checks

| Name | What it catches |
|------|----------------|
| `multiline` | Multi-line commands (breaks permission glob matching) |
| `pipe` | Pipe operators outside quotes |
| `chained` | `&&`, `\|\|`, `;` outside quotes |
| `redirect` | `>`, `>>` outside quotes |
| `subshell` | `$()`, backticks outside single quotes |
| `credential-read` | Reading `.env`, `.pem`, `.key`, SSH keys, etc. |
| `network` | `curl`, `wget`, `ssh`, etc. in command position |

Built-ins use proper shell tokenization — they understand quoting, so `grep "foo|bar"` won't trigger the pipe check. Regex patterns get `(?s)` prepended automatically so `.` matches newlines.

### Migration from nogo

If you were using `nogo` or `skill nope`, `ja nope init` will detect existing hooks in `.claude/settings.json` and migrate them to `ja nope`. Rules move from `.claude/nope.yaml` to the `nope:` section of `.toba.yaml` — you'll need to move those manually (wrap them under a `nope:` key).

## Brew Init

One-time setup for Homebrew tap automation. Creates the companion tap repo on GitHub, pushes an initial formula and README, and injects an `update-homebrew` job into the source repo's `release.yml`.

```bash
ja brew init --tap toba/homebrew-todo
```

It auto-detects the source repo, latest release tag, description, and license via `gh`. The formula SHA256 is resolved using the same three-strategy approach (`.sha256` sidecar, `checksums.txt`, direct download). After running, tap updates happen automatically via CI.

```bash
ja brew init --tap toba/homebrew-todo --tag v1.2.3 --repo toba/todo --desc "My tool" --license MIT
```

Use `--dry-run` to preview without creating anything. Use `--json` for machine-readable output.

**After running**, add a `HOMEBREW_TAP_TOKEN` secret to the source repo — a GitHub PAT with Contents write access to the tap repo.

## Zed Init

One-time setup for Zed extension automation. Creates a companion extension repo on GitHub with the full scaffold (extension.toml, Cargo.toml, src/lib.rs, bump-version script and workflow, LICENSE, README), and injects a `sync-extension` job into the source repo's `release.yml`.

```bash
ja zed init --ext toba/gozer --languages "Go Text Template,Go HTML Template"
```

It auto-detects the source repo, latest release tag, and description via `gh`. The `--languages` flag is required — it sets which languages the extension provides LSP support for. After running, extension updates happen automatically via CI.

```bash
ja zed init --ext toba/gozer --languages "CSS" --tag v1.0.0 --repo toba/go-css-lsp --desc "CSS LSP" --lsp-name go-css-lsp
```

Use `--dry-run` to preview all generated files without creating anything. Use `--json` for machine-readable output.

**After running**, add an `EXTENSION_PAT` secret to the source repo — a GitHub PAT with Contents write access to the extension repo. Also run `cargo generate-lockfile` in the extension repo to create the initial `Cargo.lock`.

## Configuration

Everything lives in `.toba.yaml`. Sections are independent — you can use any subset.

```yaml
upstream:
  sources: [...]

nope:
  debug: nope.log    # optional JSONL debug log
  rules: [...]
```

Config reading uses the yaml.v3 Node API for partial read/write, so no section clobbers another.

A [JSON Schema](https://raw.githubusercontent.com/toba/skill/main/schema.json) is available for editor autocomplete and validation. Add this modeline to the top of your `.toba.yaml`:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/toba/skill/main/schema.json
```

## Requirements

- macOS or Linux (Windows builds exist but are untested)
- `gh` CLI for upstream monitoring, brew, and zed commands (nope guard has no external dependencies)

## License

Apache-2.0
