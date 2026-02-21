package integration

import (
	"testing"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
)

func TestDetect_NilConfig(t *testing.T) {
	integ, err := Detect(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ != nil {
		t.Fatalf("expected nil integration, got %v", integ)
	}
}

func TestDetect_EmptyConfig(t *testing.T) {
	integ, err := Detect(map[string]map[string]any{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ != nil {
		t.Fatalf("expected nil integration, got %v", integ)
	}
}

func TestDetect_ClickUp(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)

	syncCfg := map[string]map[string]any{
		"clickup": {
			"list_id": "12345",
		},
	}
	integ, err := Detect(syncCfg, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ == nil {
		t.Fatal("expected non-nil integration")
	}
	if integ.Name() != "clickup" {
		t.Errorf("expected name 'clickup', got %q", integ.Name())
	}
}

func TestDetect_GitHub(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)

	syncCfg := map[string]map[string]any{
		"github": {
			"repo": "owner/repo",
		},
	}
	integ, err := Detect(syncCfg, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ == nil {
		t.Fatal("expected non-nil integration")
	}
	if integ.Name() != "github" {
		t.Errorf("expected name 'github', got %q", integ.Name())
	}
}

func TestDetect_ClickUpPriority(t *testing.T) {
	// When both clickup and github are configured, clickup wins (checked first)
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)

	syncCfg := map[string]map[string]any{
		"clickup": {"list_id": "12345"},
		"github":  {"repo": "owner/repo"},
	}
	integ, err := Detect(syncCfg, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ == nil {
		t.Fatal("expected non-nil integration")
	}
	if integ.Name() != "clickup" {
		t.Errorf("expected clickup to take priority, got %q", integ.Name())
	}
}

func TestDetect_InvalidClickUpConfig(t *testing.T) {
	// clickup present but no list_id falls through to github
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)

	syncCfg := map[string]map[string]any{
		"clickup": {"not_list_id": "12345"},
		"github":  {"repo": "owner/repo"},
	}
	integ, err := Detect(syncCfg, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ == nil {
		t.Fatal("expected non-nil integration")
	}
	if integ.Name() != "github" {
		t.Errorf("expected github fallback, got %q", integ.Name())
	}
}

func TestDetect_InvalidGitHubRepo(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)

	syncCfg := map[string]map[string]any{
		"github": {"repo": "invalid-no-slash"},
	}
	_, err := Detect(syncCfg, c)
	if err == nil {
		t.Fatal("expected error for invalid repo format")
	}
}

func TestDetect_GitHubNoRepo(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)

	syncCfg := map[string]map[string]any{
		"github": {"not_repo": "value"},
	}
	integ, err := Detect(syncCfg, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if integ != nil {
		t.Fatal("expected nil integration when github has no repo")
	}
}
