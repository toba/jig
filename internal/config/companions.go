package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Companions holds git URLs for optional companion repositories.
type Companions struct {
	Zed  string `yaml:"zed,omitempty"`
	Brew string `yaml:"brew,omitempty"`
}

// LoadCompanions extracts the companions section from a Document.
// Returns nil (no error) if the section doesn't exist.
func LoadCompanions(doc *Document) *Companions {
	node := FindKey(doc.Root, "companions")
	if node == nil {
		return nil
	}
	var c Companions
	if err := node.Decode(&c); err != nil {
		return nil
	}
	return &c
}

// SaveCompanions writes the companions section into an existing Document,
// preserving all other sections.
func SaveCompanions(doc *Document, c *Companions) error {
	var node yaml.Node
	if err := node.Encode(c); err != nil {
		return fmt.Errorf("encoding companions: %w", err)
	}

	if ReplaceKey(doc.Root, "companions", &node) {
		// Key already existed, replaced in-place.
	} else {
		// Append new key to the mapping.
		appendKey(doc.Root, "companions", &node)
	}

	data, err := marshalNode(doc.Root)
	if err != nil {
		return fmt.Errorf("marshaling document: %w", err)
	}
	return os.WriteFile(doc.Path, data, 0o644)
}

// appendKey adds a new key-value pair to the top-level mapping.
func appendKey(root *yaml.Node, key string, value *yaml.Node) {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return
	}
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	root.Content = append(root.Content, keyNode, value)
}
