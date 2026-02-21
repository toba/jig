package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the citations section of .jig.yaml (a list of sources).
type Config []Source

// Source represents a single cited repository to monitor.
type Source struct {
	Repo            string    `yaml:"repo"`
	Branch          string    `yaml:"branch"`
	Notes           string    `yaml:"notes,omitempty"`
	LastCheckedSHA  string    `yaml:"last_checked_sha,omitempty"`
	LastCheckedDate string    `yaml:"last_checked_date,omitempty"`
	Paths           PathDefs  `yaml:"paths"`
}

// PathDefs defines glob patterns grouped by relevance level.
type PathDefs struct {
	High   []string `yaml:"high,omitempty"`
	Medium []string `yaml:"medium,omitempty"`
	Low    []string `yaml:"low,omitempty"`
}

// Document holds the full YAML document tree for partial updates.
type Document struct {
	Path string
	Root *yaml.Node
}

// LoadDocument reads and parses a .jig.yaml file without requiring any
// particular section to exist. Use this when you only need the Document
// (e.g. for LoadCompanions).
func LoadDocument(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &Document{Path: path, Root: &root}, nil
}

// Load reads a .jig.yaml file and extracts only the citations section.
func Load(path string) (*Document, *Config, error) {
	doc, err := LoadDocument(path)
	if err != nil {
		return nil, nil, err
	}

	citationsNode := FindKey(doc.Root, "citations")
	if citationsNode == nil {
		return nil, nil, fmt.Errorf("no 'citations' section found in %s", path)
	}

	var cfg Config
	if err := citationsNode.Decode(&cfg); err != nil {
		return nil, nil, fmt.Errorf("decoding citations section: %w", err)
	}

	// Default branch to "main" if not set.
	for i := range cfg {
		if cfg[i].Branch == "" {
			cfg[i].Branch = "main"
		}
	}

	return doc, &cfg, nil
}

// Save writes the updated config back to the document, preserving other sections.
func Save(doc *Document, cfg *Config) error {
	// Encode the updated config into a new YAML node.
	var newCitations yaml.Node
	if err := newCitations.Encode(cfg); err != nil {
		return fmt.Errorf("encoding citations config: %w", err)
	}

	// Find and replace the citations value node in the document tree.
	if !ReplaceKey(doc.Root, "citations", &newCitations) {
		return fmt.Errorf("could not find 'citations' key in document to update")
	}

	data, err := marshalNode(doc.Root)
	if err != nil {
		return fmt.Errorf("marshaling document: %w", err)
	}

	if err := os.WriteFile(doc.Path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", doc.Path, err)
	}

	return nil
}

// FindSource returns a pointer to the source matching the given repo name.
// It matches against the full "owner/name" or just "name".
func FindSource(cfg *Config, name string) *Source {
	for i := range *cfg {
		if (*cfg)[i].Repo == name {
			return &(*cfg)[i]
		}
	}
	// Try matching just the repo name part (after /).
	for i := range *cfg {
		repo := (*cfg)[i].Repo
		for j := len(repo) - 1; j >= 0; j-- {
			if repo[j] == '/' {
				if repo[j+1:] == name {
					return &(*cfg)[i]
				}
				break
			}
		}
	}
	return nil
}

// MarkSource updates the last_checked_sha and last_checked_date for a source.
func MarkSource(src *Source, sha string) {
	src.LastCheckedSHA = sha
	src.LastCheckedDate = time.Now().UTC().Format(time.RFC3339)
}

// FindKey finds the value node for a given key in a YAML mapping node.
// When root is a DocumentNode, it descends into the first content node.
func FindKey(root *yaml.Node, key string) *yaml.Node {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == key {
			return root.Content[i+1]
		}
	}
	return nil
}

// ReplaceKey replaces the value node for a given top-level key.
func ReplaceKey(root *yaml.Node, key string, value *yaml.Node) bool {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == key {
			root.Content[i+1] = value
			return true
		}
	}
	return false
}

// AppendSource adds a new source to the citations section of the document.
// If no citations section exists, one is created.
func AppendSource(doc *Document, src Source) error {
	root := doc.Root
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node at document root")
	}

	citationsNode := FindKey(doc.Root, "citations")
	if citationsNode == nil {
		// Create citations key and empty sequence.
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "citations", Tag: "!!str"}
		citationsNode = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		root.Content = append(root.Content, keyNode, citationsNode)
	}

	// Encode the new source as a YAML node and append it.
	var srcNode yaml.Node
	if err := srcNode.Encode(&src); err != nil {
		return fmt.Errorf("encoding source: %w", err)
	}
	setFlowPaths(&srcNode)
	citationsNode.Content = append(citationsNode.Content, &srcNode)

	data, err := marshalNode(doc.Root)
	if err != nil {
		return fmt.Errorf("marshaling document: %w", err)
	}

	return os.WriteFile(doc.Path, data, 0o644)
}

// setFlowPaths finds the "paths" mapping inside a source node and sets
// its sequence children (high, medium, low) to flow style (e.g. [a, b]).
func setFlowPaths(node *yaml.Node) {
	if node.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == "paths" {
			paths := node.Content[i+1]
			if paths.Kind != yaml.MappingNode {
				return
			}
			for j := 0; j < len(paths.Content)-1; j += 2 {
				if paths.Content[j+1].Kind == yaml.SequenceNode {
					paths.Content[j+1].Style = yaml.FlowStyle
				}
			}
			return
		}
	}
}

// marshalNode marshals a yaml.Node back to bytes.
func marshalNode(node *yaml.Node) ([]byte, error) {
	return yaml.Marshal(node)
}
