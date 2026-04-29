package cc

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// HistoryEntry records the last alias used in a given working directory.
type HistoryEntry struct {
	Dir       string    `yaml:"dir"`
	Alias     string    `yaml:"alias"`
	Timestamp time.Time `yaml:"timestamp"`
}

// LoadHistory reads ~/.jig/cc-history.yaml.
func LoadHistory() ([]HistoryEntry, error) {
	path, err := HistoryPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var entries []HistoryEntry
	if err := yaml.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveHistory writes the entries to disk.
func SaveHistory(entries []HistoryEntry) error {
	path, err := HistoryPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// RecordHistory upserts the entry for the given cwd.
func RecordHistory(cwd, alias string) error {
	entries, err := LoadHistory()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	updated := false
	for i := range entries {
		if entries[i].Dir == cwd {
			entries[i].Alias = alias
			entries[i].Timestamp = now
			updated = true
			break
		}
	}
	if !updated {
		entries = append(entries, HistoryEntry{Dir: cwd, Alias: alias, Timestamp: now})
	}
	return SaveHistory(entries)
}

// LastAlias returns the alias most recently used in the given cwd, or "".
func LastAlias(cwd string) string {
	entries, _ := LoadHistory()
	for _, e := range entries {
		if e.Dir == cwd {
			return e.Alias
		}
	}
	return ""
}
