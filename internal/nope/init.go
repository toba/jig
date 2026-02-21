package nope

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// StarterConfig is the default nope rules written to .toba.yaml.
var StarterConfig = `nope:
  rules:
    - name: multiline-commands
      builtin: multiline
      message: >-
        multiline bash commands break permission matching
        (glob * can't match newlines). Reformat as a single
        line and use the description parameter for comments.

    - name: git-push
      pattern: 'git\s+push'
      message: "git push not allowed — only user should push"

    - name: git-checkout-switch
      pattern: 'git\s+(checkout|switch)\s+'
      message: "git checkout/switch not allowed — branch changes require user approval"

    - name: destructive-rm
      pattern: 'rm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+|--force\s+).*(/|~|\$HOME)'
      message: "destructive rm on broad paths not allowed"

    - name: curl-bearer
      pattern: 'curl.*Authorization.*Bearer'
      message: "direct API calls with bearer tokens not allowed"

    - name: pipe-commands
      builtin: pipe
      message: >-
        piped commands not allowed — run commands separately
        so each can be individually reviewed.

    - name: chained-commands
      builtin: chained
      message: >-
        chained commands (&&, ||, ;) not allowed — run commands
        separately so each can be individually reviewed.

    - name: redirect-output
      builtin: redirect
      message: >-
        output redirection (>, >>) not allowed — use tee or
        write to file directly.

    - name: subshell-expansion
      builtin: subshell
      message: >-
        subshell expansion ($(), backticks) not allowed — run
        commands separately.

    - name: credential-read
      builtin: credential-read
      message: >-
        reading credential or secret files (.env, .pem, .key,
        ssh keys, etc.) not allowed.

    - name: network-access
      builtin: network
      message: >-
        direct network tool usage (curl, wget, ssh, etc.) not
        allowed — only user should make network requests.

    # Non-Bash tool rules
    - name: no-write-env
      pattern: '"file_path"\s*:\s*"[^"]*\.env"'
      tools: ["Write", "Edit"]
      message: "writing to .env files not allowed"

    - name: no-write-credentials
      pattern: '"file_path"\s*:\s*"[^"]*(\.pem|\.key|\.p12|credentials\.json)"'
      tools: ["Write", "Edit"]
      message: "writing to credential files not allowed"
`

const hookCommand = "jig nope"

var hookEntry = map[string]any{
	"matcher": ".*",
	"hooks": []any{
		map[string]any{
			"type":    "command",
			"command": hookCommand,
		},
	},
}

// RunInit scaffolds the nope section in .toba.yaml and the hook in .claude/settings.json.
func RunInit() int {
	tobaPath := ".toba.yaml"
	claudeDir := ".claude"
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Write nope section to .toba.yaml
	if data, err := os.ReadFile(tobaPath); err == nil {
		// File exists — check if it already has a nope section.
		content := string(data)
		if hasNopeSection(content) {
			fmt.Fprintf(os.Stderr, "nope init: %s already contains a 'nope' section, skipping\n", tobaPath)
		} else {
			// Append nope section.
			if len(content) > 0 && content[len(content)-1] != '\n' {
				content += "\n"
			}
			content += "\n" + StarterConfig
			if err := os.WriteFile(tobaPath, []byte(content), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "nope init: write %s: %v\n", tobaPath, err)
				return 1
			}
			fmt.Fprintf(os.Stderr, "nope init: added nope section to %s\n", tobaPath)
		}
	} else {
		// Create new file.
		if err := os.WriteFile(tobaPath, []byte(StarterConfig), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "nope init: write %s: %v\n", tobaPath, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "nope init: created %s\n", tobaPath)
	}

	// Write or merge settings.json
	if err := os.MkdirAll(claudeDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "nope init: create %s: %v\n", claudeDir, err)
		return 1
	}

	if err := mergeSettings(settingsPath); err != nil {
		fmt.Fprintf(os.Stderr, "nope init: %v\n", err)
		return 1
	}

	return 0
}

func hasNopeSection(content string) bool {
	return slices.Contains(splitLines(content), "nope:")
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func mergeSettings(path string) error {
	var settings map[string]any

	data, err := os.ReadFile(path) //nolint:gosec // settings path is constructed internally
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read %s: %w", path, err)
		}
		// File doesn't exist — create fresh
		settings = map[string]any{
			"hooks": map[string]any{
				"PreToolUse": []any{hookEntry},
			},
		}
		return writeSettings(path, settings, true)
	}

	// File exists — parse and merge
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	if hasNopeHook(settings) {
		if migrateLegacyCommand(settings) {
			fmt.Fprintf(os.Stderr, "nope init: migrated nogo hook command to %q\n", hookCommand)
			return writeSettings(path, settings, false)
		}
		if migrateNogoMatcher(settings) {
			fmt.Fprintf(os.Stderr, "nope init: migrated nogo hook matcher to \".*\"\n")
			return writeSettings(path, settings, false)
		}
		fmt.Fprintf(os.Stderr, "nope init: %s already has nope hook, skipping\n", path)
		return nil
	}

	// Ensure hooks.PreToolUse exists and append our entry
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}

	preToolUse, _ := hooks["PreToolUse"].([]any)
	hooks["PreToolUse"] = append(preToolUse, hookEntry)

	return writeSettings(path, settings, false)
}

func writeSettings(path string, settings map[string]any, created bool) error {
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if created {
		fmt.Fprintf(os.Stderr, "nope init: created %s\n", path)
	} else {
		fmt.Fprintf(os.Stderr, "nope init: updated %s with nope hook\n", path)
	}
	return nil
}

// migrateNogoMatcher updates existing nope hook entries that have a "Bash"
// matcher to ".*" so nope runs for all tools. Returns true if any change was made.
func migrateNogoMatcher(settings map[string]any) bool {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return false
	}
	preToolUse, _ := hooks["PreToolUse"].([]any)
	changed := false
	for _, entry := range preToolUse {
		m, _ := entry.(map[string]any)
		if m == nil {
			continue
		}
		if m["matcher"] != "Bash" {
			continue
		}
		innerHooks, _ := m["hooks"].([]any)
		for _, h := range innerHooks {
			hm, _ := h.(map[string]any)
			if hm != nil && isNopeCommand(hm["command"]) {
				m["matcher"] = ".*"
				changed = true
				break
			}
		}
	}
	return changed
}

// migrateLegacyCommand updates existing hook entries that use "nogo",
// "skill nope", or "ja nope" command to use "jig nope" instead. Returns true if any change was made.
func migrateLegacyCommand(settings map[string]any) bool {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return false
	}
	preToolUse, _ := hooks["PreToolUse"].([]any)
	changed := false
	for _, entry := range preToolUse {
		m, _ := entry.(map[string]any)
		if m == nil {
			continue
		}
		innerHooks, _ := m["hooks"].([]any)
		for _, h := range innerHooks {
			hm, _ := h.(map[string]any)
			if hm == nil {
				continue
			}
			cmd, _ := hm["command"].(string)
			if cmd == "nogo" || cmd == "skill nope" || cmd == "ja nope" {
				hm["command"] = hookCommand
				changed = true
			}
		}
	}
	return changed
}

// hasNopeHook checks whether settings already contains a nope/nogo hook in PreToolUse.
func hasNopeHook(settings map[string]any) bool {
	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		return false
	}
	preToolUse, _ := hooks["PreToolUse"].([]any)
	for _, entry := range preToolUse {
		m, _ := entry.(map[string]any)
		if m == nil {
			continue
		}
		innerHooks, _ := m["hooks"].([]any)
		for _, h := range innerHooks {
			hm, _ := h.(map[string]any)
			if hm != nil && isNopeCommand(hm["command"]) {
				return true
			}
		}
	}
	return false
}

// isNopeCommand returns true if the command is "jig nope", "ja nope", "skill nope", or "nogo".
func isNopeCommand(cmd any) bool {
	s, ok := cmd.(string)
	if !ok {
		return false
	}
	return s == hookCommand || s == "skill nope" || s == "ja nope" || s == "nogo"
}
