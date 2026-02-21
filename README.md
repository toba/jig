# skill

A multi-tool CLI that does two mostly unrelated things under one roof because *why maintain two binaries when you can maintain one slightly confused binary*.

**Upstream monitoring** — track changes in repos you've forked or derived from, classify files by how much you care, and remember what you've already reviewed.

**Nope guard** — a Claude Code `PreToolUse` hook that blocks dangerous tool invocations. When Claude Code wears you down and you go YOLO with `dangerously-skip-permissions`, *nope* keeps a few guardrails in place so your agent doesn't `rm -rf /` your life's work or `git push --force` your main branch into oblivion.

Formerly two projects: `upstream` and [`nogo`](https://github.com/toba/nogo). Now they share a config file (`.toba.yaml`) and a single binary.

## Install

```bash
brew install toba/skill/skill
```

Or build from source:

```bash
go install github.com/toba/skill@latest
```

## Upstream Monitoring

Track what's changed in repos you care about. Read-only by default — `check` looks, `mark` remembers.

```bash
# Add a source to .toba.yaml
skill init

# See what's changed upstream
skill check

# Mark a source as reviewed (updates last_checked_sha)
skill mark owner/repo
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

### Quick Start

```bash
skill nope init
```

This does two things:
1. Adds a `nope:` section to `.toba.yaml` with starter rules
2. Wires up the hook in `.claude/settings.json`

### How It Works

Claude Code pipes a JSON payload to `skill nope` on stdin before each tool call. If a rule matches, the tool is blocked (exit 2). If nothing matches, it's allowed (exit 0).

```bash
# Blocked — exit 2
echo '{"tool_name":"Bash","tool_input":{"command":"git push"}}' | skill nope

# Allowed — exit 0
echo '{"tool_name":"Bash","tool_input":{"command":"ls"}}' | skill nope
```

### Rules

Rules are either regex patterns or built-in checks:

```yaml
nope:
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

### Other Commands

```bash
skill nope doctor    # Validate your config
skill nope help      # Full reference
```

### Migration from nogo

If you were using `nogo`, `skill nope init` will detect existing `nogo` hooks in `.claude/settings.json` and migrate them to `skill nope`. Rules move from `.claude/nope.yaml` to the `nope:` section of `.toba.yaml` — you'll need to move those manually (wrap them under a `nope:` key).

## Configuration

Everything lives in `.toba.yaml`. The upstream and nope sections are independent — you can use one without the other.

```yaml
upstream:
  sources: [...]

nope:
  debug: nope.log    # optional JSONL debug log
  rules: [...]
```

Config reading uses the yaml.v3 Node API for partial read/write, so neither section clobbers the other (or any other sections you might have in the file).

## Requirements

- macOS or Linux (Windows builds exist but are untested)
- `gh` CLI for upstream monitoring (nope guard has no external dependencies)

## License

Apache-2.0
