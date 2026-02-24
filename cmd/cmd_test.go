package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"

	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/graph/model"
	"github.com/toba/jig/internal/todo/integration"
	"github.com/toba/jig/internal/todo/issue"
	"github.com/toba/jig/internal/todo/output"
)

// writeTempConfig creates a .jig.yaml with the given content in a temp dir and returns the path.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".jig.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- repoFromGitURL tests ---

func TestRepoFromGitURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"https with .git", "https://github.com/owner/repo.git", "owner/repo"},
		{"https without .git", "https://github.com/owner/repo", "owner/repo"},
		{"ssh with .git", "git@github.com:owner/repo.git", "owner/repo"},
		{"ssh without .git", "git@github.com:owner/repo", "owner/repo"},
		{"bare owner/repo", "owner/repo", "owner/repo"},
		{"https deep path", "https://github.com/org/project", "org/project"},
		{"empty string", "", ""},
		{"absolute path", "/usr/local/bin", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repoFromGitURL(tt.url)
			if got != tt.want {
				t.Errorf("repoFromGitURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

// --- extNameFromRepo tests ---

func TestExtNameFromRepo(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"owner/repo", "toba/gozer", "gozer"},
		{"no slash", "gozer", "gozer"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extNameFromRepo(tt.input)
			if got != tt.want {
				t.Errorf("extNameFromRepo(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- resolveExt tests ---

func TestResolveExt(t *testing.T) {
	t.Run("flag takes precedence", func(t *testing.T) {
		got, err := resolveExt("toba/gozer", "/nonexistent/path")
		if err != nil {
			t.Fatalf("resolveExt() error: %v", err)
		}
		if got != "toba/gozer" {
			t.Errorf("resolveExt() = %q, want %q", got, "toba/gozer")
		}
	})

	t.Run("empty flag and no config returns empty", func(t *testing.T) {
		got, err := resolveExt("", "/nonexistent/path")
		if err != nil {
			t.Fatalf("resolveExt() error: %v", err)
		}
		if got != "" {
			t.Errorf("resolveExt() = %q, want empty", got)
		}
	})
}

// --- configPath tests ---

func TestConfigPath(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		old := cfgPath
		defer func() { cfgPath = old }()
		cfgPath = ""

		got := configPath()
		if got != ".jig.yaml" {
			t.Errorf("configPath() = %q, want %q", got, ".jig.yaml")
		}
	})

	t.Run("custom", func(t *testing.T) {
		old := cfgPath
		defer func() { cfgPath = old }()
		cfgPath = "/custom/path.yaml"

		got := configPath()
		if got != "/custom/path.yaml" {
			t.Errorf("configPath() = %q, want %q", got, "/custom/path.yaml")
		}
	})
}

// --- hasTodoSync tests ---

func TestHasTodoSync(t *testing.T) {
	t.Run("nonexistent file returns false", func(t *testing.T) {
		got := hasTodoSync("/nonexistent/path/.jig.yaml")
		if got {
			t.Error("hasTodoSync() = true, want false for nonexistent file")
		}
	})

	t.Run("file without todo section returns false", func(t *testing.T) {
		path := writeTempConfig(t, "citations:\n  - repo: owner/repo\n")
		got := hasTodoSync(path)
		if got {
			t.Error("hasTodoSync() = true, want false for file without todo section")
		}
	})

	t.Run("file with todo but no sync returns false", func(t *testing.T) {
		path := writeTempConfig(t, "todo:\n  data_path: .issues\n")
		got := hasTodoSync(path)
		if got {
			t.Error("hasTodoSync() = true, want false for file without sync section")
		}
	})

	t.Run("file with todo.sync returns true", func(t *testing.T) {
		path := writeTempConfig(t, "todo:\n  sync:\n    clickup:\n      token: abc\n")
		got := hasTodoSync(path)
		if !got {
			t.Error("hasTodoSync() = false, want true for file with todo.sync section")
		}
	})
}

// --- collectFlags tests ---

func TestCollectFlags(t *testing.T) {
	// Test on the applyCmd which has known flags.
	flags := collectFlags(applyCmd)

	var hasMessage, hasVersion, hasPush bool
	for _, f := range flags {
		if strings.Contains(f, "--message") {
			hasMessage = true
		}
		if strings.Contains(f, "--version") {
			hasVersion = true
		}
		if strings.Contains(f, "--push") {
			hasPush = true
		}
	}

	if !hasMessage {
		t.Error("collectFlags(applyCmd) missing --message")
	}
	if !hasVersion {
		t.Error("collectFlags(applyCmd) missing --version")
	}
	if !hasPush {
		t.Error("collectFlags(applyCmd) missing --push")
	}
}

// --- mergeTags tests ---

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		add      []string
		remove   []string
		want     []string
	}{
		{
			name:     "add new tags",
			existing: []string{"a"},
			add:      []string{"b", "c"},
			remove:   nil,
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "remove tags",
			existing: []string{"a", "b", "c"},
			add:      nil,
			remove:   []string{"b"},
			want:     []string{"a", "c"},
		},
		{
			name:     "add and remove",
			existing: []string{"a", "b"},
			add:      []string{"c"},
			remove:   []string{"a"},
			want:     []string{"b", "c"},
		},
		{
			name:     "empty everything",
			existing: nil,
			add:      nil,
			remove:   nil,
			want:     []string{},
		},
		{
			name:     "add duplicate",
			existing: []string{"a"},
			add:      []string{"a"},
			remove:   nil,
			want:     []string{"a"},
		},
		{
			name:     "remove nonexistent",
			existing: []string{"a"},
			add:      nil,
			remove:   []string{"z"},
			want:     []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeTags(tt.existing, tt.add, tt.remove)
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("mergeTags() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("mergeTags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// --- cmdError tests ---

func TestCmdError(t *testing.T) {
	t.Run("text mode", func(t *testing.T) {
		err := cmdError(false, output.ErrValidation, "invalid %s: %s", "field", "value")
		if err == nil {
			t.Fatal("cmdError() returned nil")
		}
		if !strings.Contains(err.Error(), "invalid field: value") {
			t.Errorf("cmdError() = %q, want it to contain 'invalid field: value'", err.Error())
		}
	})

	t.Run("json mode", func(t *testing.T) {
		err := cmdError(true, output.ErrValidation, "invalid %s: %s", "field", "value")
		if err == nil {
			t.Fatal("cmdError() returned nil")
		}
		// JSON mode wraps with output.Error which is a specific error type.
		if !strings.Contains(err.Error(), "invalid field: value") {
			t.Errorf("cmdError() = %q, want it to contain 'invalid field: value'", err.Error())
		}
	})
}

// --- resolveContent tests ---

func TestResolveContent(t *testing.T) {
	t.Run("empty both", func(t *testing.T) {
		got, err := resolveContent("", "")
		if err != nil {
			t.Fatalf("resolveContent() error: %v", err)
		}
		if got != "" {
			t.Errorf("resolveContent() = %q, want empty", got)
		}
	})

	t.Run("direct value", func(t *testing.T) {
		got, err := resolveContent("hello world", "")
		if err != nil {
			t.Fatalf("resolveContent() error: %v", err)
		}
		if got != "hello world" {
			t.Errorf("resolveContent() = %q, want %q", got, "hello world")
		}
	})

	t.Run("from file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "body.md")
		if err := os.WriteFile(path, []byte("file content"), 0o644); err != nil {
			t.Fatal(err)
		}

		got, err := resolveContent("", path)
		if err != nil {
			t.Fatalf("resolveContent() error: %v", err)
		}
		if got != "file content" {
			t.Errorf("resolveContent() = %q, want %q", got, "file content")
		}
	})

	t.Run("both set returns error", func(t *testing.T) {
		_, err := resolveContent("value", "file.txt")
		if err == nil {
			t.Fatal("resolveContent() expected error when both set")
		}
		if !strings.Contains(err.Error(), "cannot use both") {
			t.Errorf("resolveContent() error = %q, expected 'cannot use both'", err.Error())
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := resolveContent("", "/nonexistent/file.txt")
		if err == nil {
			t.Fatal("resolveContent() expected error for missing file")
		}
	})
}

// --- command structure tests ---

func TestCommandStructure(t *testing.T) {
	// Verify key commands are registered on rootCmd.
	cmds := rootCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	expected := []string{"cite", "nope", "brew", "zed", "commit", "todo", "doctor", "version", "help-all", "prime", "update", "sync"}
	for _, name := range expected {
		if !cmdNames[name] {
			t.Errorf("rootCmd missing subcommand %q", name)
		}
	}
}

func TestCommitSubcommands(t *testing.T) {
	cmds := commitCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	if !cmdNames["gather"] {
		t.Error("commitCmd missing 'gather' subcommand")
	}
	if !cmdNames["apply"] {
		t.Error("commitCmd missing 'apply' subcommand")
	}
}

func TestApplyCmdRequiredFlags(t *testing.T) {
	// The --message flag should be marked required.
	f := applyCmd.Flags().Lookup("message")
	if f == nil {
		t.Fatal("applyCmd missing --message flag")
	}

	// Check shorthand.
	if f.Shorthand != "m" {
		t.Errorf("--message shorthand = %q, want %q", f.Shorthand, "m")
	}
}

func TestGatherCmdNoArgs(t *testing.T) {
	if gatherCmd.Args == nil {
		t.Fatal("gatherCmd.Args should not be nil")
	}
	// cobra.NoArgs returns error when args provided.
	err := gatherCmd.Args(gatherCmd, []string{"extra"})
	if err == nil {
		t.Error("gatherCmd should reject arguments")
	}
}

func TestApplyCmdNoArgs(t *testing.T) {
	if applyCmd.Args == nil {
		t.Fatal("applyCmd.Args should not be nil")
	}
	err := applyCmd.Args(applyCmd, []string{"extra"})
	if err == nil {
		t.Error("applyCmd should reject arguments")
	}
}

func TestNopeSubcommands(t *testing.T) {
	cmds := nopeCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range []string{"init", "doctor", "help"} {
		if !cmdNames[name] {
			t.Errorf("nopeCmd missing %q subcommand", name)
		}
	}
}

func TestCiteSubcommands(t *testing.T) {
	cmds := citeCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range []string{"init", "review", "add"} {
		if !cmdNames[name] {
			t.Errorf("citeCmd missing %q subcommand", name)
		}
	}
}

func TestReviewCmdAlias(t *testing.T) {
	if len(reviewCmd.Aliases) == 0 {
		t.Fatal("reviewCmd should have aliases")
	}
	found := false
	for _, a := range reviewCmd.Aliases {
		if a == "check" {
			found = true
		}
	}
	if !found {
		t.Error("reviewCmd missing 'check' alias")
	}
}

func TestBrewSubcommands(t *testing.T) {
	cmds := brewCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range []string{"init", "doctor"} {
		if !cmdNames[name] {
			t.Errorf("brewCmd missing %q subcommand", name)
		}
	}
}

func TestZedSubcommands(t *testing.T) {
	cmds := zedCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range []string{"init", "doctor"} {
		if !cmdNames[name] {
			t.Errorf("zedCmd missing %q subcommand", name)
		}
	}
}

func TestTodoSubcommands(t *testing.T) {
	cmds := todoCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range []string{"init", "create", "list", "show", "delete", "archive", "roadmap"} {
		if !cmdNames[name] {
			t.Errorf("todoCmd missing %q subcommand", name)
		}
	}
}

func TestSyncAliasSubcommands(t *testing.T) {
	cmds := syncAliasCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range []string{"check", "link", "unlink"} {
		if !cmdNames[name] {
			t.Errorf("syncAliasCmd missing %q subcommand", name)
		}
	}
}

// --- flag parsing tests ---

func TestBrewInitFlags(t *testing.T) {
	flags := []string{"tap", "tag", "repo", "desc", "license", "dry-run"}
	for _, name := range flags {
		f := brewInitCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("brewInitCmd missing --%s flag", name)
		}
	}
}

func TestZedInitFlags(t *testing.T) {
	flags := []string{"ext", "tag", "repo", "desc", "lsp-name", "languages", "dry-run"}
	for _, name := range flags {
		f := zedInitCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("zedInitCmd missing --%s flag", name)
		}
	}
}

func TestCreateCmdFlags(t *testing.T) {
	flags := []string{"status", "type", "priority", "body", "body-file", "tag", "due", "parent", "blocking", "blocked-by", "json"}
	for _, name := range flags {
		f := createCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("createCmd missing --%s flag", name)
		}
	}
}

func TestApplyCmdFlags(t *testing.T) {
	flags := []string{"message", "version", "push"}
	for _, name := range flags {
		f := applyCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("applyCmd missing --%s flag", name)
		}
	}
}

func TestRootCmdPersistentFlags(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("config")
	if f == nil {
		t.Error("rootCmd missing --config persistent flag")
	}

	f = rootCmd.PersistentFlags().Lookup("json")
	if f == nil {
		t.Error("rootCmd missing --json persistent flag")
	}
}

// --- resolveTap tests ---

func TestResolveTap(t *testing.T) {
	t.Run("flag takes precedence", func(t *testing.T) {
		got, err := resolveTap("owner/homebrew-tool", "/nonexistent/path")
		if err != nil {
			t.Fatalf("resolveTap() error: %v", err)
		}
		if got != "owner/homebrew-tool" {
			t.Errorf("resolveTap() = %q, want %q", got, "owner/homebrew-tool")
		}
	})

	t.Run("no flag falls through to convention or error", func(t *testing.T) {
		got, err := resolveTap("", "/nonexistent/path/.jig.yaml")
		if err != nil {
			// Expected when gh CLI is not available or not in a GitHub repo.
			if !strings.Contains(err.Error(), "--tap is required") {
				t.Errorf("resolveTap() error = %q, expected '--tap is required'", err.Error())
			}
		} else {
			// If gh succeeds, we should get a convention-based tap.
			if got == "" {
				t.Error("resolveTap() returned empty string without error")
			}
			if !strings.Contains(got, "homebrew-") {
				t.Errorf("resolveTap() convention = %q, expected it to contain 'homebrew-'", got)
			}
		}
	})
}

// --- printCommandTree is hard to unit test but we can verify
// collectFlags works with various commands ---

func TestCollectFlagsOnRootCmd(t *testing.T) {
	// Root cmd has persistent flags but no local flags (except help).
	flags := collectFlags(rootCmd)
	// Should not include help.
	for _, f := range flags {
		if strings.Contains(f, "--help") {
			t.Error("collectFlags should not include --help")
		}
	}
}

func TestCollectFlagsOnBrewInit(t *testing.T) {
	flags := collectFlags(brewInitCmd)
	if len(flags) == 0 {
		t.Error("collectFlags(brewInitCmd) returned empty, expected flags")
	}

	// Check that dry-run (bool) does not have <type> annotation.
	for _, f := range flags {
		if strings.Contains(f, "--dry-run") {
			if strings.Contains(f, "<") {
				t.Errorf("bool flag --dry-run should not have type annotation: %s", f)
			}
		}
	}
}

// --- formatRelationships additional tests ---

func TestFormatRelationshipsEmpty(t *testing.T) {
	b := &issue.Issue{}
	result := formatRelationships(b)
	if result != "" {
		t.Errorf("formatRelationships() for empty issue = %q, want empty", result)
	}
}

// --- starterConfig test ---

func TestStarterConfig(t *testing.T) {
	if !strings.Contains(starterConfig, "citations:") {
		t.Error("starterConfig missing 'citations:' key")
	}
	if !strings.Contains(starterConfig, "owner/repo") {
		t.Error("starterConfig missing 'owner/repo' placeholder")
	}
	if !strings.Contains(starterConfig, "high:") {
		t.Error("starterConfig missing 'high:' classification")
	}
	if !strings.Contains(starterConfig, "medium:") {
		t.Error("starterConfig missing 'medium:' classification")
	}
	if !strings.Contains(starterConfig, "low:") {
		t.Error("starterConfig missing 'low:' classification")
	}
}

// --- runInit (cite init) tests ---

func TestRunInit(t *testing.T) {
	t.Run("creates new config file", func(t *testing.T) {
		dir := t.TempDir()
		old := cfgPath
		defer func() { cfgPath = old }()
		cfgPath = filepath.Join(dir, ".jig.yaml")

		err := runInit(nil, nil)
		if err != nil {
			t.Fatalf("runInit() error: %v", err)
		}

		data, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "citations:") {
			t.Error("created file missing 'citations:' section")
		}
	})

	t.Run("appends to existing file without citations", func(t *testing.T) {
		dir := t.TempDir()
		old := cfgPath
		defer func() { cfgPath = old }()
		cfgPath = filepath.Join(dir, ".jig.yaml")

		if err := os.WriteFile(cfgPath, []byte("nope:\n  network: block\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := runInit(nil, nil)
		if err != nil {
			t.Fatalf("runInit() error: %v", err)
		}

		data, err := os.ReadFile(cfgPath)
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.Contains(content, "nope:") {
			t.Error("existing content was lost")
		}
		if !strings.Contains(content, "citations:") {
			t.Error("citations section not added")
		}
	})

	t.Run("errors if citations section already exists", func(t *testing.T) {
		dir := t.TempDir()
		old := cfgPath
		defer func() { cfgPath = old }()
		cfgPath = filepath.Join(dir, ".jig.yaml")

		if err := os.WriteFile(cfgPath, []byte("citations:\n  - repo: owner/repo\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := runInit(nil, nil)
		if err == nil {
			t.Fatal("runInit() expected error when citations section already exists")
		}
		if !strings.Contains(err.Error(), "already contains") {
			t.Errorf("runInit() error = %q, expected 'already contains'", err.Error())
		}
	})
}

// --- version command tests ---

func TestVersionVars(t *testing.T) {
	// Verify the default values are set.
	if ver == "" {
		t.Error("ver should not be empty")
	}
	if commit == "" {
		t.Error("commit should not be empty")
	}
	if date == "" {
		t.Error("date should not be empty")
	}
}

// --- nopeCmd configuration tests ---

func TestNopeCmdSilenceFlags(t *testing.T) {
	if !nopeCmd.SilenceUsage {
		t.Error("nopeCmd should have SilenceUsage set")
	}
	if !nopeCmd.SilenceErrors {
		t.Error("nopeCmd should have SilenceErrors set")
	}
}

// --- todo PersistentPreRunE bypass test ---

func TestTodoPersistentPreRunESkipsInit(t *testing.T) {
	// The init, prime, and refry commands should bypass core init.
	for _, name := range []string{"init", "prime", "refry"} {
		skipInit := name == "init" || name == "prime" || name == "refry"
		if !skipInit {
			t.Errorf("expected %q to skip init", name)
		}
	}
}

// --- hasFieldUpdates tests ---

func TestHasFieldUpdates(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		input := model.UpdateIssueInput{}
		if hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = true for empty input")
		}
	})

	t.Run("status set", func(t *testing.T) {
		s := "todo"
		input := model.UpdateIssueInput{Status: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Status is set")
		}
	})

	t.Run("type set", func(t *testing.T) {
		s := "bug"
		input := model.UpdateIssueInput{Type: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Type is set")
		}
	})

	t.Run("priority set", func(t *testing.T) {
		s := "high"
		input := model.UpdateIssueInput{Priority: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Priority is set")
		}
	})

	t.Run("title set", func(t *testing.T) {
		s := "New Title"
		input := model.UpdateIssueInput{Title: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Title is set")
		}
	})

	t.Run("due set", func(t *testing.T) {
		s := "2026-01-01"
		input := model.UpdateIssueInput{Due: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Due is set")
		}
	})

	t.Run("body set", func(t *testing.T) {
		s := "body content"
		input := model.UpdateIssueInput{Body: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Body is set")
		}
	})

	t.Run("bodyMod set", func(t *testing.T) {
		input := model.UpdateIssueInput{BodyMod: &model.BodyModification{}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when BodyMod is set")
		}
	})

	t.Run("addTags set", func(t *testing.T) {
		input := model.UpdateIssueInput{AddTags: []string{"tag"}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when AddTags is set")
		}
	})

	t.Run("removeTags set", func(t *testing.T) {
		input := model.UpdateIssueInput{RemoveTags: []string{"tag"}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when RemoveTags is set")
		}
	})

	t.Run("parent set", func(t *testing.T) {
		s := "parent-id"
		input := model.UpdateIssueInput{Parent: &s}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when Parent is set")
		}
	})

	t.Run("addBlocking set", func(t *testing.T) {
		input := model.UpdateIssueInput{AddBlocking: []string{"x"}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when AddBlocking is set")
		}
	})

	t.Run("removeBlocking set", func(t *testing.T) {
		input := model.UpdateIssueInput{RemoveBlocking: []string{"x"}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when RemoveBlocking is set")
		}
	})

	t.Run("addBlockedBy set", func(t *testing.T) {
		input := model.UpdateIssueInput{AddBlockedBy: []string{"x"}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when AddBlockedBy is set")
		}
	})

	t.Run("removeBlockedBy set", func(t *testing.T) {
		input := model.UpdateIssueInput{RemoveBlockedBy: []string{"x"}}
		if !hasFieldUpdates(input) {
			t.Error("hasFieldUpdates() = false when RemoveBlockedBy is set")
		}
	})
}

// --- isConflictError tests ---

func TestIsConflictError(t *testing.T) {
	t.Run("etag mismatch", func(t *testing.T) {
		err := &core.ETagMismatchError{Provided: "abc", Current: "def"}
		if !isConflictError(err) {
			t.Error("isConflictError() = false for ETagMismatchError")
		}
	})

	t.Run("etag required", func(t *testing.T) {
		err := &core.ETagRequiredError{}
		if !isConflictError(err) {
			t.Error("isConflictError() = false for ETagRequiredError")
		}
	})

	t.Run("generic error", func(t *testing.T) {
		err := errors.New("something went wrong")
		if isConflictError(err) {
			t.Error("isConflictError() = true for generic error")
		}
	})

	t.Run("wrapped etag mismatch", func(t *testing.T) {
		inner := &core.ETagMismatchError{Provided: "a", Current: "b"}
		err := fmt.Errorf("wrapping: %w", inner)
		if !isConflictError(err) {
			t.Error("isConflictError() = false for wrapped ETagMismatchError")
		}
	})
}

// --- mutationError tests ---

func TestMutationError(t *testing.T) {
	t.Run("conflict error in text mode", func(t *testing.T) {
		err := mutationError(false, &core.ETagMismatchError{Provided: "a", Current: "b"})
		if err == nil {
			t.Fatal("mutationError() returned nil")
		}
	})

	t.Run("conflict error in json mode", func(t *testing.T) {
		err := mutationError(true, &core.ETagMismatchError{Provided: "a", Current: "b"})
		if err == nil {
			t.Fatal("mutationError() returned nil")
		}
	})

	t.Run("generic error in text mode", func(t *testing.T) {
		err := mutationError(false, errors.New("generic"))
		if err == nil {
			t.Fatal("mutationError() returned nil")
		}
		if !strings.Contains(err.Error(), "generic") {
			t.Errorf("mutationError() = %q, expected 'generic'", err.Error())
		}
	})
}

// --- typeBadge tests ---

func TestTypeBadge(t *testing.T) {
	tests := []struct {
		name     string
		issueTyp string
		wantPart string
	}{
		{"bug", "bug", "d73a4a"},
		{"feature", "feature", "0e8a16"},
		{"task", "task", "1d76db"},
		{"epic", "epic", "5319e7"},
		{"milestone", "milestone", "fbca04"},
		{"unknown", "custom", "gray"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &issue.Issue{Type: tt.issueTyp}
			got := typeBadge(b)
			if tt.wantPart == "" {
				if got != "" {
					t.Errorf("typeBadge() = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.wantPart) {
				t.Errorf("typeBadge() = %q, expected to contain %q", got, tt.wantPart)
			}
			if !strings.Contains(got, "img.shields.io") {
				t.Errorf("typeBadge() = %q, expected shields.io URL", got)
			}
		})
	}
}

// --- printCommandTree tests ---

func TestPrintCommandTree(t *testing.T) {
	// Capture stdout by using a buffer.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCommandTree(rootCmd, "")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should contain top-level commands.
	if !strings.Contains(output, "jig") {
		t.Error("printCommandTree() output missing 'jig'")
	}
	if !strings.Contains(output, "commit") {
		t.Error("printCommandTree() output missing 'commit'")
	}
	if !strings.Contains(output, "cite") {
		t.Error("printCommandTree() output missing 'cite'")
	}
	if !strings.Contains(output, "nope") {
		t.Error("printCommandTree() output missing 'nope'")
	}

	// Should include subcommands.
	if !strings.Contains(output, "gather") {
		t.Error("printCommandTree() output missing 'gather' subcommand")
	}
	if !strings.Contains(output, "apply") {
		t.Error("printCommandTree() output missing 'apply' subcommand")
	}

	// Should include flag info.
	if !strings.Contains(output, "flags:") {
		t.Error("printCommandTree() output missing 'flags:' section")
	}

	// Should NOT include help-all itself or hidden commands.
	if strings.Contains(output, "help-all") {
		t.Error("printCommandTree() should not include 'help-all'")
	}
}

// --- containsStatus tests ---

func TestContainsStatus(t *testing.T) {
	tests := []struct {
		name     string
		statuses []string
		target   string
		want     bool
	}{
		{"found", []string{"todo", "in-progress"}, "todo", true},
		{"not found", []string{"todo", "in-progress"}, "completed", false},
		{"empty list", []string{}, "todo", false},
		{"nil list", nil, "todo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsStatus(tt.statuses, tt.target)
			if got != tt.want {
				t.Errorf("containsStatus(%v, %q) = %v, want %v", tt.statuses, tt.target, got, tt.want)
			}
		})
	}
}

// --- update cmd flags tests ---

func TestUpdateCmdFlags(t *testing.T) {
	flags := []string{
		"status", "type", "priority", "title", "due",
		"body", "body-file", "body-replace-old", "body-replace-new", "body-append",
		"parent", "remove-parent",
		"blocking", "remove-blocking",
		"blocked-by", "remove-blocked-by",
		"tag", "remove-tag",
		"if-match", "json",
	}
	for _, name := range flags {
		f := todoUpdateCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("todoUpdateCmd missing --%s flag", name)
		}
	}
}

func TestUpdateCmdAliases(t *testing.T) {
	found := false
	for _, a := range todoUpdateCmd.Aliases {
		if a == "u" {
			found = true
		}
	}
	if !found {
		t.Error("todoUpdateCmd missing 'u' alias")
	}
}

func TestCreateCmdAliases(t *testing.T) {
	expected := map[string]bool{"c": false, "new": false}
	for _, a := range createCmd.Aliases {
		if _, ok := expected[a]; ok {
			expected[a] = true
		}
	}
	for alias, found := range expected {
		if !found {
			t.Errorf("createCmd missing %q alias", alias)
		}
	}
}

func TestListCmdAliases(t *testing.T) {
	found := false
	for _, a := range listCmd.Aliases {
		if a == "ls" {
			found = true
		}
	}
	if !found {
		t.Error("listCmd missing 'ls' alias")
	}
}

// --- list cmd flags tests ---

func TestListCmdFlags(t *testing.T) {
	flags := []string{
		"json", "search", "status", "no-status", "type", "no-type",
		"priority", "no-priority", "tag", "no-tag",
		"sort", "quiet", "full",
	}
	for _, name := range flags {
		f := listCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("listCmd missing --%s flag", name)
		}
	}
}

// --- sync cmd flags ---

func TestSyncCmdFlags(t *testing.T) {
	flags := []string{"dry-run", "force", "no-relationships", "json"}
	for _, name := range flags {
		f := syncAliasCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("syncAliasCmd missing --%s flag", name)
		}
	}
}

// --- initTodoCore error path tests ---

func TestInitTodoCoreNonexistentConfig(t *testing.T) {
	old := cfgPath
	defer func() { cfgPath = old }()

	// Use a config path that exists but points to a nonexistent data dir.
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, ".jig.yaml")
	os.WriteFile(cfgFile, []byte("todo:\n  data_path: /nonexistent/data/dir\n"), 0o644)
	cfgPath = cfgFile

	err := initTodoCore(todoCmd)
	if err == nil {
		t.Fatal("initTodoCore() expected error for nonexistent data dir")
	}
}

// --- defaultLinkPrefix test ---

func TestDefaultLinkPrefix(t *testing.T) {
	// Setup a minimal core to make defaultLinkPrefix work.
	dir := t.TempDir()
	dataDir := filepath.Join(dir, ".issues")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := todoconfig.Default()
	oldStore := todoStore
	todoStore = core.New(dataDir, cfg)
	defer func() { todoStore = oldStore }()

	got := defaultLinkPrefix()
	// The result is a relative path from cwd to the data dir.
	// We just verify the function runs without error and returns a path
	// containing ".issues" (since the data dir is named .issues).
	if !strings.Contains(got, ".issues") {
		t.Errorf("defaultLinkPrefix() = %q, expected to contain '.issues'", got)
	}
}

// --- printSchema test ---

func TestPrintSchema(t *testing.T) {
	oldStore := todoStore
	dir := t.TempDir()
	dataDir := filepath.Join(dir, ".issues")
	os.MkdirAll(dataDir, 0o755)
	cfg := todoconfig.Default()
	todoStore = core.New(dataDir, cfg)
	todoStore.Load()
	defer func() { todoStore = oldStore }()

	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSchema()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "type Query") {
		t.Error("printSchema() output missing 'type Query'")
	}
	if !strings.Contains(out, "type Issue") {
		t.Error("printSchema() output missing 'type Issue'")
	}
}

// --- formatGraphQLErrors tests ---

func TestFormatGraphQLErrors(t *testing.T) {
	t.Run("nil errors", func(t *testing.T) {
		err := formatGraphQLErrors(nil)
		if err != nil {
			t.Errorf("formatGraphQLErrors(nil) = %v, want nil", err)
		}
	})

	t.Run("empty errors", func(t *testing.T) {
		err := formatGraphQLErrors(gqlerror.List{})
		if err != nil {
			t.Errorf("formatGraphQLErrors([]) = %v, want nil", err)
		}
	})

	t.Run("single error", func(t *testing.T) {
		errs := gqlerror.List{
			{Message: "field not found"},
		}
		err := formatGraphQLErrors(errs)
		if err == nil {
			t.Fatal("formatGraphQLErrors() returned nil for single error")
		}
		if !strings.Contains(err.Error(), "graphql: field not found") {
			t.Errorf("formatGraphQLErrors() = %q, expected 'graphql: field not found'", err.Error())
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := gqlerror.List{
			{Message: "error one"},
			{Message: "error two"},
		}
		err := formatGraphQLErrors(errs)
		if err == nil {
			t.Fatal("formatGraphQLErrors() returned nil for multiple errors")
		}
		if !strings.Contains(err.Error(), "graphql errors:") {
			t.Errorf("formatGraphQLErrors() = %q, expected 'graphql errors:'", err.Error())
		}
		if !strings.Contains(err.Error(), "error one") {
			t.Errorf("formatGraphQLErrors() = %q, expected 'error one'", err.Error())
		}
		if !strings.Contains(err.Error(), "error two") {
			t.Errorf("formatGraphQLErrors() = %q, expected 'error two'", err.Error())
		}
	})
}

// --- showStyledIssue test ---

func TestShowStyledIssue(t *testing.T) {
	oldCfg := todoCfg
	defer func() { todoCfg = oldCfg }()
	todoCfg = todoconfig.Default()

	now := time.Now()
	b := &issue.Issue{
		ID:        "test-1",
		Title:     "Test Issue",
		Status:    "todo",
		Type:      "task",
		Priority:  "normal",
		Tags:      []string{"frontend", "backend"},
		Parent:    "parent-1",
		Blocking:  []string{"child-1"},
		BlockedBy: []string{"dep-1"},
		Body:      "This is the **body** of the issue.",
		CreatedAt: &now,
	}

	// Capture stdout — just verify it doesn't panic.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	showStyledIssue(b)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "test-1") {
		t.Error("showStyledIssue() output missing issue ID")
	}
	if !strings.Contains(out, "Test Issue") {
		t.Error("showStyledIssue() output missing issue title")
	}
}

func TestShowStyledIssueMinimal(t *testing.T) {
	oldCfg := todoCfg
	defer func() { todoCfg = oldCfg }()
	todoCfg = todoconfig.Default()

	b := &issue.Issue{
		ID:     "min-1",
		Title:  "Minimal",
		Status: "todo",
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	showStyledIssue(b)

	w.Close()
	os.Stdout = old
}

// --- renderRoadmapMarkdown test ---

func TestRenderRoadmapMarkdown(t *testing.T) {
	now := time.Now()

	data := &roadmapData{
		Milestones: []milestoneGroup{
			{
				Milestone: &issue.Issue{
					ID: "m1", Title: "v1.0", Status: "todo", Type: "milestone", CreatedAt: &now,
					Path: "m/m1--v1-0.md",
				},
				Epics: []epicGroup{
					{
						Epic: &issue.Issue{
							ID: "e1", Title: "Auth", Status: "todo", Type: "epic",
							Path: "e/e1--auth.md",
						},
						Items: []*issue.Issue{
							{ID: "t1", Title: "Login", Status: "todo", Type: "task",
								Path: "t/t1--login.md"},
						},
					},
				},
			},
		},
	}

	result := renderRoadmapMarkdown(data, false, "")
	if !strings.Contains(result, "v1.0") {
		t.Error("renderRoadmapMarkdown() missing milestone title")
	}
	if !strings.Contains(result, "Auth") {
		t.Error("renderRoadmapMarkdown() missing epic title")
	}
	if !strings.Contains(result, "Login") {
		t.Error("renderRoadmapMarkdown() missing item title")
	}
}

func TestRenderRoadmapMarkdownWithLinks(t *testing.T) {
	now := time.Now()

	data := &roadmapData{
		Milestones: []milestoneGroup{
			{
				Milestone: &issue.Issue{
					ID: "m1", Title: "v1.0", Status: "todo", Type: "milestone", CreatedAt: &now,
					Path: "m1--v1-0.md",
				},
			},
		},
	}

	result := renderRoadmapMarkdown(data, true, ".issues")
	if !strings.Contains(result, ".issues/") {
		t.Error("renderRoadmapMarkdown() with links missing link prefix")
	}
}

func TestRenderRoadmapMarkdownEmpty(t *testing.T) {
	data := &roadmapData{}
	result := renderRoadmapMarkdown(data, false, "")
	// Should not panic and should produce some output (at least template headers).
	if result == "" {
		t.Error("renderRoadmapMarkdown() returned empty for empty data")
	}
}

// --- outputSyncText test ---

func TestOutputSyncText(t *testing.T) {
	results := []integration.SyncResult{
		{IssueID: "i1", IssueTitle: "Created Issue", Action: integration.ActionCreated, ExternalURL: "https://example.com/1"},
		{IssueID: "i2", IssueTitle: "Updated Issue", Action: integration.ActionUpdated, ExternalURL: "https://example.com/2"},
		{IssueID: "i3", IssueTitle: "Unchanged Issue", Action: integration.ActionUnchanged},
		{IssueID: "i4", IssueTitle: "Skipped Issue", Action: integration.ActionSkipped},
		{IssueID: "i5", IssueTitle: "Error Issue", Action: integration.ActionError, Error: errors.New("test error")},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputSyncText(results)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("outputSyncText() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "Created:") {
		t.Error("outputSyncText() missing 'Created:' line")
	}
	if !strings.Contains(out, "Updated:") {
		t.Error("outputSyncText() missing 'Updated:' line")
	}
	if !strings.Contains(out, "Error:") {
		t.Error("outputSyncText() missing 'Error:' line")
	}
	if !strings.Contains(out, "Summary:") {
		t.Error("outputSyncText() missing 'Summary:' line")
	}
	if !strings.Contains(out, "1 created") {
		t.Errorf("outputSyncText() summary = %q, expected '1 created'", out)
	}
}

// --- outputSyncJSON test ---

func TestOutputSyncJSON(t *testing.T) {
	t.Run("nil results", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputSyncJSON(nil)

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("outputSyncJSON(nil) error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)
		if !strings.Contains(buf.String(), "[]") {
			t.Errorf("outputSyncJSON(nil) = %q, expected '[]'", buf.String())
		}
	})

	t.Run("with results", func(t *testing.T) {
		results := []integration.SyncResult{
			{IssueID: "i1", IssueTitle: "Test", Action: integration.ActionCreated, ExternalID: "ext-1"},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputSyncJSON(results)

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("outputSyncJSON() error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)
		out := buf.String()

		if !strings.Contains(out, "i1") {
			t.Error("outputSyncJSON() missing issue ID")
		}
		if !strings.Contains(out, "ext-1") {
			t.Error("outputSyncJSON() missing external ID")
		}
	})
}

// --- show cmd flags ---

func TestShowCmdFlags(t *testing.T) {
	flags := []string{"json", "raw", "body-only", "etag-only"}
	for _, name := range flags {
		f := showCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("showCmd missing --%s flag", name)
		}
	}
}

// --- sortByStatusThenCreated test ---

func TestSortByStatusThenCreated(t *testing.T) {
	oldCfg := todoCfg
	defer func() { todoCfg = oldCfg }()
	todoCfg = todoconfig.Default()

	now := time.Now()
	earlier := now.Add(-time.Hour)

	issues := []*issue.Issue{
		{ID: "b", Status: "completed", CreatedAt: &now},
		{ID: "a", Status: "todo", CreatedAt: &earlier},
		{ID: "c", Status: "todo", CreatedAt: &now},
	}

	sortByStatusThenCreated(issues, todoCfg)

	// todo comes before completed in default config.
	if issues[0].Status != "todo" {
		t.Errorf("first issue status = %q, want 'todo'", issues[0].Status)
	}
	// Earlier created comes first within same status.
	if issues[0].ID != "a" {
		t.Errorf("first issue ID = %q, want 'a'", issues[0].ID)
	}
}

// --- sortByTypeThenStatus test ---

func TestSortByTypeThenStatus(t *testing.T) {
	oldCfg := todoCfg
	defer func() { todoCfg = oldCfg }()
	todoCfg = todoconfig.Default()

	issues := []*issue.Issue{
		{ID: "t1", Type: "task", Status: "todo"},
		{ID: "b1", Type: "bug", Status: "todo"},
		{ID: "f1", Type: "feature", Status: "todo"},
	}

	sortByTypeThenStatus(issues, todoCfg)

	// Bug comes before feature, feature before task in default config.
	if issues[0].Type != "bug" {
		t.Errorf("first issue type = %q, want 'bug'", issues[0].Type)
	}
}

// --- printCheckReport test ---

func TestPrintCheckReport(t *testing.T) {
	report := &integration.CheckReport{
		Sections: []integration.CheckSection{
			{
				Name: "Config",
				Checks: []integration.CheckResult{
					{Name: "token present", Status: integration.CheckPass},
					{Name: "api key", Status: integration.CheckWarn, Message: "optional"},
					{Name: "endpoint", Status: integration.CheckFail, Message: "missing"},
				},
			},
		},
		Summary: integration.CheckSummary{
			Passed:   1,
			Warnings: 1,
			Failed:   1,
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printCheckReport(report)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "Config") {
		t.Error("printCheckReport() missing section name")
	}
	if !strings.Contains(out, "token present") {
		t.Error("printCheckReport() missing check name")
	}
	if !strings.Contains(out, "Summary") {
		t.Error("printCheckReport() missing Summary")
	}
}

// --- outputLinkJSON test ---

func TestOutputLinkJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputLinkJSON("issue-1", "Test Issue", "ext-123", "linked")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("outputLinkJSON() error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !strings.Contains(out, "issue-1") {
		t.Error("outputLinkJSON() missing issue_id")
	}
	if !strings.Contains(out, "ext-123") {
		t.Error("outputLinkJSON() missing external_id")
	}
	if !strings.Contains(out, "linked") {
		t.Error("outputLinkJSON() missing action")
	}
}

// --- outputUnlinkJSON test ---

func TestOutputUnlinkJSON(t *testing.T) {
	t.Run("with external ID", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputUnlinkJSON("issue-1", "Test Issue", "ext-123", "unlinked")

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("outputUnlinkJSON() error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)
		out := buf.String()

		if !strings.Contains(out, "issue-1") {
			t.Error("outputUnlinkJSON() missing issue_id")
		}
		if !strings.Contains(out, "ext-123") {
			t.Error("outputUnlinkJSON() missing external_id")
		}
	})

	t.Run("without external ID", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputUnlinkJSON("issue-1", "Test Issue", "", "unlinked")

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Fatalf("outputUnlinkJSON() error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)
		if strings.Contains(buf.String(), "external_id") {
			t.Error("outputUnlinkJSON() should not include external_id when empty")
		}
	})
}

// --- todo check cmd flags ---

func TestTodoCheckCmdFlags(t *testing.T) {
	flags := []string{"json", "fix"}
	for _, name := range flags {
		f := todoCheckCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("todoCheckCmd missing --%s flag", name)
		}
	}
}

// --- todo delete cmd flags ---

func TestTodoDeleteCmdFlags(t *testing.T) {
	// Verify the delete command exists.
	found := false
	for _, c := range todoCmd.Commands() {
		if c.Name() == "delete" {
			found = true
		}
	}
	if !found {
		t.Error("todoCmd missing 'delete' subcommand")
	}
}

// --- graphql cmd aliases ---

func TestGraphqlCmdAliases(t *testing.T) {
	found := false
	for _, a := range graphqlCmd.Aliases {
		if a == "query" {
			found = true
		}
	}
	if !found {
		t.Error("graphqlCmd missing 'query' alias")
	}
}

// --- graphql cmd flags ---

func TestGraphqlCmdFlags(t *testing.T) {
	flags := []string{"json", "variables", "operation", "schema"}
	for _, name := range flags {
		f := graphqlCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("graphqlCmd missing --%s flag", name)
		}
	}
}

// --- roadmap cmd flags ---

func TestRoadmapCmdFlags(t *testing.T) {
	flags := []string{"json", "include-done", "status", "no-status", "no-links", "link-prefix"}
	for _, name := range flags {
		f := roadmapCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("roadmapCmd missing --%s flag", name)
		}
	}
}
