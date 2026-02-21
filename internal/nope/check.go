package nope

import (
	"encoding/json"
	"fmt"
)

// ExtractCommand pulls the "command" field from JSON tool input.
func ExtractCommand(input string) string {
	var obj struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(input), &obj); err != nil {
		return ""
	}
	return obj.Command
}

// syntheticInput builds a JSON input string wrapping cmd in a "command" field.
func syntheticInput(cmd string) string {
	return fmt.Sprintf(`{"command":%q}`, cmd)
}

// CheckRules runs each rule against the input for the given tool.
// For compound commands (containing &&, ||, or ;), each segment is also
// checked independently so that pattern and builtin rules can catch
// dangerous commands hidden after innocuous ones.
// Returns the block message of the first matching rule, or "" if none match.
func CheckRules(rules []CompiledRule, toolName, input string, logger *DebugLogger) string {
	// Extract segments for compound command splitting.
	cmd := ExtractCommand(input)
	segments := SplitSegments(cmd)

	for _, r := range rules {
		if !r.ToolMatch(toolName) {
			logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "tool-skip"})
			continue
		}
		// Always check the original input first (structural rules like
		// chained/pipe need to see the full compound command).
		if r.Check(input) {
			logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "match"})
			return r.Message
		}
		// If there are multiple segments, check each one independently.
		if len(segments) > 1 {
			for _, seg := range segments {
				if r.Check(syntheticInput(seg)) {
					logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "match", "segment": seg})
					return r.Message
				}
			}
		}
		logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "no-match"})
	}
	return ""
}
