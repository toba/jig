package cc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetectedDir describes a discovered ~/.claude* directory.
type DetectedDir struct {
	Name      string // alias name derived from dir suffix (".claude" → "main")
	Path      string
	Score     int  // count of "real" entries — higher means more likely source
	HasMarker bool // matched at least one marker file/dir
}

// markers identify a directory as a Claude config home.
var markers = []string{
	".credentials.json",
	".claude.json",
	"CLAUDE.md",
	"agents",
	"skills",
	"commands",
	"projects",
}

// Detect scans $HOME for ~/.claude and ~/.claude-* directories.
func Detect(home string) ([]DetectedDir, error) {
	entries, err := os.ReadDir(home)
	if err != nil {
		return nil, err
	}
	var out []DetectedDir
	for _, e := range entries {
		name := e.Name()
		if name != ".claude" && !strings.HasPrefix(name, ".claude-") {
			continue
		}
		full := filepath.Join(home, name)
		info, err := os.Stat(full)
		if err != nil || !info.IsDir() {
			continue
		}
		d := DetectedDir{
			Name: aliasNameFromDir(name),
			Path: full,
		}
		score, hasMarker := scoreDir(full)
		d.Score = score
		d.HasMarker = hasMarker
		out = append(out, d)
	}
	return out, nil
}

func aliasNameFromDir(dirName string) string {
	if dirName == ".claude" {
		return "main"
	}
	return strings.TrimPrefix(dirName, ".claude-")
}

// scoreDir returns the count of real (non-symlink) entries among markers
// plus a flag indicating whether any markers exist at all.
func scoreDir(dir string) (int, bool) {
	score := 0
	hasMarker := false
	for _, m := range markers {
		p := filepath.Join(dir, m)
		info, err := os.Lstat(p)
		if err != nil {
			continue
		}
		hasMarker = true
		if info.Mode()&os.ModeSymlink == 0 {
			score++
		}
	}
	return score, hasMarker
}

// InitOpts controls Init.
type InitOpts struct {
	Home   string // override for testing; defaults to os.UserHomeDir
	Source string // explicit source path (overrides auto-detection)
	Force  bool   // overwrite existing config
}

// InitResult describes what Init did.
type InitResult struct {
	ConfigPath   string                 `json:"config_path"`
	SharedSource string                 `json:"shared_source"`
	SourceAlias  string                 `json:"source_alias"`
	Aliases      []string               `json:"aliases"`
	Synced       map[string]*SyncReport `json:"synced,omitempty"`
}

// Init scans the user's home, picks a source, and writes a config.
func Init(opts InitOpts) (*InitResult, error) {
	home := opts.Home
	if home == "" {
		h, err := Home()
		if err != nil {
			return nil, err
		}
		home = h
	}

	cfgPath := filepath.Join(home, ".jig", "cc.yaml")
	if !opts.Force {
		if _, err := os.Stat(cfgPath); err == nil {
			return nil, fmt.Errorf("config already exists at %s (use --force to overwrite)", cfgPath)
		}
	}

	dirs, err := Detect(home)
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 && opts.Source == "" {
		return nil, errors.New("no ~/.claude* directories found; pass --source to specify one")
	}

	cfg := &Config{
		Version: 1,
		Private: DefaultPrivate,
		Aliases: map[string]Alias{},
	}

	// Determine source.
	var sourcePath, sourceAlias string
	if opts.Source != "" {
		sourcePath = opts.Source
		sourceAlias = aliasNameFromDir(filepath.Base(sourcePath))
		if sourceAlias == "" {
			sourceAlias = "main"
		}
	} else {
		// Pick the dir with the highest score as source. Ties: prefer ".claude" itself.
		best := -1
		for i, d := range dirs {
			if !d.HasMarker {
				continue
			}
			if best == -1 || d.Score > dirs[best].Score ||
				(d.Score == dirs[best].Score && filepath.Base(d.Path) == ".claude") {
				best = i
			}
		}
		if best == -1 {
			return nil, errors.New("no usable ~/.claude* directory found")
		}
		sourcePath = dirs[best].Path
		sourceAlias = dirs[best].Name
	}
	cfg.SharedSource = sourcePath
	cfg.Aliases[sourceAlias] = Alias{
		CLI:      "claude",
		Path:     sourcePath,
		IsSource: true,
	}

	// Add other detected dirs as non-source aliases (path is moved into ~/.jig/cc/<name>).
	aliasesRoot := filepath.Join(home, ".jig", "cc")
	for _, d := range dirs {
		if d.Path == sourcePath {
			continue
		}
		if !d.HasMarker {
			continue
		}
		newPath := filepath.Join(aliasesRoot, d.Name)
		cfg.Aliases[d.Name] = Alias{
			CLI:  "claude",
			Path: newPath,
		}
	}

	// Validate before writing.
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := cfg.SaveTo(cfgPath); err != nil {
		return nil, err
	}

	// Build alias dirs and copy private files from the original ~/.claude-*
	// dir (if any) before linking.
	res := &InitResult{
		ConfigPath:   cfgPath,
		SharedSource: cfg.SharedSource,
		SourceAlias:  sourceAlias,
		Synced:       map[string]*SyncReport{},
	}
	for name, a := range cfg.Aliases {
		res.Aliases = append(res.Aliases, name)
		if a.IsSource {
			continue
		}
		// Find the original detected dir matching this alias name to copy
		// private files from.
		for _, d := range dirs {
			if d.Name == name && d.Path != a.Path {
				_ = CopyPrivateFiles(d.Path, a.Path, cfg.PrivateList())
				break
			}
		}
		// Ensure .claude.json exists so claude can launch even before login.
		_ = SeedClaudeJSON(cfg, name)
		rep, err := Sync(cfg, name)
		if err != nil {
			return res, err
		}
		res.Synced[name] = rep
	}
	return res, nil
}
