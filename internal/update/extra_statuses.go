package update

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// migrateExtraStatuses populates `todo.extra_statuses` based on the
// configured sync integration, restoring the historical default-status
// set that was implicit before status togglability was introduced.
//
//   - github sync → adds in-progress, draft, scrapped (omits review,
//     since GitHub only tracks open/closed)
//   - clickup sync → adds in-progress, review, draft, scrapped
//   - no sync configured → no-op even if extra_statuses is missing
//   - extra_statuses already present → no-op
//
// `ready` and `completed` are mandatory and not listed in the map.
// `deferred` is a new status and not part of the historical default set,
// so projects must opt into it explicitly.
func migrateExtraStatuses(jigPath string) (bool, error) {
	data, err := os.ReadFile(jigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("reading %s: %w", jigPath, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return false, fmt.Errorf("parsing %s: %w", jigPath, err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return false, nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return false, nil
	}

	todoNode := findMapValue(doc, "todo")
	if todoNode == nil || todoNode.Kind != yaml.MappingNode {
		return false, nil
	}

	// Already migrated.
	if findMapValue(todoNode, "extra_statuses") != nil {
		return false, nil
	}

	syncNode := findMapValue(todoNode, "sync")
	if syncNode == nil || syncNode.Kind != yaml.MappingNode {
		return false, nil
	}

	hasGithub := findMapValue(syncNode, "github") != nil
	hasClickup := findMapValue(syncNode, "clickup") != nil
	if !hasGithub && !hasClickup {
		return false, nil
	}

	// Build the additive list. Order matches DefaultStatuses for stable output.
	var entries []string
	entries = append(entries, "in-progress")
	if hasClickup {
		entries = append(entries, "review")
	}
	entries = append(entries, "draft", "scrapped")

	extraNode := buildBoolMapNode(entries)
	insertMapEntry(todoNode, "extra_statuses", extraNode)

	out, err := yaml.Marshal(&root)
	if err != nil {
		return false, fmt.Errorf("marshaling %s: %w", jigPath, err)
	}
	if err := os.WriteFile(jigPath, out, 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", jigPath, err)
	}
	return true, nil
}

// findMapValue returns the value node for a key in a mapping, or nil.
func findMapValue(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// insertMapEntry appends a new key/value pair to a mapping node.
func insertMapEntry(m *yaml.Node, key string, value *yaml.Node) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		value,
	)
}

// buildBoolMapNode builds a YAML mapping node where each name maps to true.
func buildBoolMapNode(names []string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, name := range names {
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: name},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "true", Tag: "!!bool"},
		)
	}
	return node
}
