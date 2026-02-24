package nope

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// HookInput is the JSON payload Claude Code sends to PreToolUse hooks via stdin.
type HookInput struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
}

// ExitError is a sentinel error type that carries a process exit code.
type ExitError struct {
	Code int
}

func (e ExitError) Error() string {
	return fmt.Sprintf("exit %d", e.Code)
}

// RunGuard is the core guard logic: read stdin, check rules, return exit code.
// Returns nil for allow (exit 0), or ExitError with the appropriate code.
func RunGuard(version string) error {
	cfg, root, err := FindAndLoadConfig()
	if err != nil {
		// No config found — pass through silently. This allows nope to be
		// installed globally without erroring in repos that haven't run init.
		return nil
	}

	logger := NewDebugLogger(cfg.Debug, root)
	defer logger.Close()

	logger.Log(map[string]any{"event": "start", "version": version, "pid": os.Getpid()})

	rules, err := CompileRules(cfg.Rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nope: %v\n", err)
		return ExitError{Code: 1}
	}

	toolName, input, err := ReadHookInput()
	if err != nil {
		logger.Log(map[string]any{"event": "result", "result": "allow", "reason": "stdin read error", "error": err.Error()})
		return nil
	}
	if input == "" {
		logger.Log(map[string]any{"event": "result", "tool": toolName, "result": "allow", "reason": "empty input"})
		return nil
	}

	logger.Log(map[string]any{"event": "check", "tool": toolName, "input": input})

	if msg := CheckRules(rules, toolName, input, logger); msg != "" {
		logger.Log(map[string]any{"event": "result", "tool": toolName, "result": "block", "message": msg})
		fmt.Fprintf(os.Stderr, "BLOCK: %s\n", msg)
		return ExitError{Code: 2}
	}

	logger.Log(map[string]any{"event": "result", "tool": toolName, "result": "allow"})
	return nil
}

// ReadHookInput reads the JSON payload from stdin that Claude Code sends
// to PreToolUse hooks. Returns the tool name and the tool_input as a JSON
// string (for pattern matching against rules).
func ReadHookInput() (toolName, input string, err error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", "", fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return "Bash", "", nil
	}

	var hi HookInput
	if err := json.Unmarshal(data, &hi); err != nil {
		return "", "", fmt.Errorf("parsing stdin JSON: %w", err)
	}

	if hi.ToolName == "" {
		hi.ToolName = "Bash"
	}

	if len(hi.ToolInput) == 0 || string(hi.ToolInput) == "null" {
		return hi.ToolName, "", nil
	}

	return hi.ToolName, string(hi.ToolInput), nil
}
