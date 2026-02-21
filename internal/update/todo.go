package update

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// migrateTodoConfig detects the legacy todo format in .toba.yaml where
// "issues:" and "sync:" are top-level keys, and restructures them under
// a "todo:" key with "issues:" renamed to its inner fields and "sync:"
// nested inside "todo:".
// Returns (migrated bool, error).
func migrateTodoConfig(tobaPath string) (bool, error) {
	data, err := os.ReadFile(tobaPath) //nolint:gosec // path from caller
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("reading %s: %w", tobaPath, err)
	}

	lines := splitLines(string(data))

	// Already migrated.
	if sectionExists(lines, "todo") {
		return false, nil
	}

	// Need at least the legacy "issues:" key to recognize the pattern.
	if !sectionExists(lines, "issues") {
		return false, nil
	}

	// Parse into a generic map to extract the legacy keys.
	var raw yaml.Node
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return false, fmt.Errorf("parsing %s: %w", tobaPath, err)
	}
	if raw.Kind != yaml.DocumentNode || len(raw.Content) == 0 {
		return false, nil
	}
	mapping := raw.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return false, nil
	}

	// Collect the issues and sync nodes, and everything else.
	var issuesValue, syncValue *yaml.Node
	var otherKeys []*yaml.Node // pairs of key, value
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key := mapping.Content[i]
		val := mapping.Content[i+1]
		switch key.Value {
		case "issues":
			issuesValue = val
		case "sync":
			syncValue = val
		default:
			otherKeys = append(otherKeys, key, val)
		}
	}

	if issuesValue == nil {
		return false, nil
	}

	// Build the todo mapping node from the issues fields + sync.
	todoMapping := &yaml.Node{Kind: yaml.MappingNode}
	if issuesValue.Kind == yaml.MappingNode {
		todoMapping.Content = append(todoMapping.Content, issuesValue.Content...)
	}
	if syncValue != nil {
		todoMapping.Content = append(todoMapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "sync"},
			syncValue,
		)
	}

	// Rebuild the top-level mapping: other keys first, then todo.
	newMapping := &yaml.Node{Kind: yaml.MappingNode}
	newMapping.Content = append(newMapping.Content, otherKeys...)
	newMapping.Content = append(newMapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "todo"},
		todoMapping,
	)

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{newMapping}}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return false, fmt.Errorf("marshaling %s: %w", tobaPath, err)
	}

	// Ensure trailing newline.
	result := string(out)
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	if err := os.WriteFile(tobaPath, []byte(result), 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", tobaPath, err)
	}

	fmt.Fprintf(os.Stderr, "update: restructured issues/sync → todo section in %s\n", tobaPath)
	return true, nil
}
