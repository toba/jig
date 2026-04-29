package cc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultPrivate is the canonical list of per-alias private files/dirs.
// Items in this list are real files in each alias dir (not symlinks).
var DefaultPrivate = []string{
	".credentials.json",
	".claude.json",
	"policy-limits.json",
	"mcp-needs-auth-cache.json",
	"remote-settings.json",
	"settings.local.json",
	"stats-cache.json",
	"statsig",
	"telemetry",
}

// Alias is one Claude profile entry.
type Alias struct {
	CLI      string `yaml:"cli"`
	Path     string `yaml:"path"`
	IsSource bool   `yaml:"is_source,omitempty"`
}

// Config is the on-disk shape of ~/.jig/cc.yaml.
type Config struct {
	Version      int              `yaml:"version"`
	SharedSource string           `yaml:"shared_source"`
	Aliases      map[string]Alias `yaml:"aliases"`
	Private      []string         `yaml:"private"`
}

// Load reads ~/.jig/cc.yaml. Returns os.ErrNotExist if absent.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom reads a Config from the given path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", path, err)
	}
	return &c, nil
}

// Save writes the config to ~/.jig/cc.yaml, creating parent dirs.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	return c.SaveTo(path)
}

// SaveTo writes the config to the given path.
func (c *Config) SaveTo(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Validate checks structural invariants.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("version must be 1, got %d", c.Version)
	}
	if strings.TrimSpace(c.SharedSource) == "" {
		return errors.New("shared_source is required")
	}
	if len(c.Aliases) == 0 {
		return errors.New("at least one alias is required")
	}
	sourceCount := 0
	for name, a := range c.Aliases {
		if name == "" {
			return errors.New("alias name cannot be empty")
		}
		if strings.TrimSpace(a.CLI) == "" {
			return fmt.Errorf("alias %q: cli is required", name)
		}
		if strings.TrimSpace(a.Path) == "" {
			return fmt.Errorf("alias %q: path is required", name)
		}
		if a.IsSource {
			sourceCount++
		}
	}
	if sourceCount != 1 {
		return fmt.Errorf("exactly one alias must have is_source: true, got %d", sourceCount)
	}
	return nil
}

// SourceAlias returns the alias marked is_source.
func (c *Config) SourceAlias() (string, Alias, bool) {
	for n, a := range c.Aliases {
		if a.IsSource {
			return n, a, true
		}
	}
	return "", Alias{}, false
}

// PrivateList returns the configured private list, falling back to defaults
// when empty.
func (c *Config) PrivateList() []string {
	if len(c.Private) == 0 {
		return DefaultPrivate
	}
	return c.Private
}

// Names returns alias names sorted.
func (c *Config) Names() []string {
	out := make([]string, 0, len(c.Aliases))
	for n := range c.Aliases {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Resolve picks an alias by exact name or unique prefix.
func (c *Config) Resolve(query string) (string, Alias, error) {
	if a, ok := c.Aliases[query]; ok {
		return query, a, nil
	}
	var matches []string
	for n := range c.Aliases {
		if strings.HasPrefix(n, query) {
			matches = append(matches, n)
		}
	}
	switch len(matches) {
	case 0:
		return "", Alias{}, fmt.Errorf("no alias matches %q", query)
	case 1:
		return matches[0], c.Aliases[matches[0]], nil
	default:
		sort.Strings(matches)
		return "", Alias{}, fmt.Errorf("ambiguous alias %q: matches %s", query, strings.Join(matches, ", "))
	}
}
