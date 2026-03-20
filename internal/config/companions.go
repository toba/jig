package config

import (
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

// LoadPackages reads the packages list from a Document.
// Returns nil if the section doesn't exist.
func LoadPackages(doc *Document) []string {
	node := FindKey(doc.Root, "packages")
	if node == nil {
		return nil
	}
	var pkgs []string
	if err := node.Decode(&pkgs); err != nil {
		return nil
	}
	return pkgs
}

// HasPackage returns true if the given package name is in the packages list.
func HasPackage(doc *Document, name string) bool {
	return slices.Contains(LoadPackages(doc), name)
}

// AddPackage adds a package to the packages list if not already present.
func AddPackage(doc *Document, name string) error {
	pkgs := LoadPackages(doc)
	if slices.Contains(pkgs, name) {
		return nil
	}
	pkgs = append(pkgs, name)
	slices.Sort(pkgs)

	var node yaml.Node
	if err := node.Encode(pkgs); err != nil {
		return fmt.Errorf("encoding packages: %w", err)
	}
	node.Style = yaml.FlowStyle

	if ReplaceKey(doc.Root, "packages", &node) {
		// replaced
	} else {
		appendKey(doc.Root, "packages", &node)
	}

	data, err := marshalNode(doc.Root)
	if err != nil {
		return fmt.Errorf("marshaling document: %w", err)
	}
	return os.WriteFile(doc.Path, data, 0o644)
}

// LoadZedExtension reads the zed_extension value from a Document.
// Returns "" if the key doesn't exist.
func LoadZedExtension(doc *Document) string {
	node := FindKey(doc.Root, "zed_extension")
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}

// SaveZedExtension writes the zed_extension value into a Document.
func SaveZedExtension(doc *Document, repo string) error {
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: repo, Tag: "!!str"}

	if ReplaceKey(doc.Root, "zed_extension", node) {
		// replaced
	} else {
		appendKey(doc.Root, "zed_extension", node)
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
