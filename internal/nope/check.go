package nope

import (
	"encoding/json"
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

// CheckRules runs each rule against the input for the given tool.
// Returns the block message of the first matching rule, or "" if none match.
func CheckRules(rules []CompiledRule, toolName, input string, logger *DebugLogger) string {
	for _, r := range rules {
		if !r.ToolMatch(toolName) {
			logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "tool-skip"})
			continue
		}
		if r.Check(input) {
			logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "match"})
			return r.Message
		}
		logger.Log(map[string]any{"event": "rule", "rule": r.Name, "result": "no-match"})
	}
	return ""
}
