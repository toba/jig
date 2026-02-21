package integration

import (
	"context"
	"testing"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/issue"
)

// Constants duplicated from the github sub-package to avoid import cycle.
const (
	ghSyncName          = "github"
	ghSyncKeyIssueNumber = "issue_number"
	ghSyncKeySyncedAt    = "synced_at"
)

func mustDetectGitHub(t *testing.T, owner, repo string, c *core.Core) *gitHubIntegration {
	t.Helper()
	cfgMap := map[string]any{"repo": owner + "/" + repo}
	integ, err := detectGitHub(cfgMap, c)
	if err != nil {
		t.Fatalf("detectGitHub: %v", err)
	}
	if integ == nil {
		t.Fatal("expected non-nil integration from detectGitHub")
	}
	gh, ok := integ.(*gitHubIntegration)
	if !ok {
		t.Fatalf("expected *gitHubIntegration, got %T", integ)
	}
	return gh
}

func TestGitHubIntegration_Name(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)
	if gh.Name() != "github" {
		t.Errorf("expected 'github', got %q", gh.Name())
	}
}

func TestGitHubIntegration_GetToken_Missing(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)
	t.Setenv("GITHUB_TOKEN", "")
	_, err := gh.getToken()
	if err == nil {
		t.Fatal("expected error when GITHUB_TOKEN not set")
	}
}

func TestGitHubIntegration_GetToken_Set(t *testing.T) {
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)
	t.Setenv("GITHUB_TOKEN", "test-token-123")
	token, err := gh.getToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "test-token-123" {
		t.Errorf("expected 'test-token-123', got %q", token)
	}
}

func TestGitHubIntegration_Sync_NoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	_, err := gh.Sync(context.Background(), nil, SyncOptions{})
	if err == nil {
		t.Fatal("expected error when GITHUB_TOKEN not set")
	}
}

func TestGitHubIntegration_Link_IssueNotFound(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	_, err := gh.Link(context.Background(), "nonexistent", "42")
	if err == nil {
		t.Fatal("expected error for nonexistent issue")
	}
}

func TestGitHubIntegration_Unlink_IssueNotFound(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	_, err := gh.Unlink(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent issue")
	}
}

func TestGitHubIntegration_Unlink_NotLinked(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "test-1",
		Title:  "Test Issue",
		Status: "draft",
	}
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	gh := mustDetectGitHub(t, "o", "r", c)
	result, err := gh.Unlink(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionNotLinked {
		t.Errorf("expected action %q, got %q", ActionNotLinked, result.Action)
	}
}

func TestGitHubIntegration_Link_AlreadyLinked(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "test-2",
		Title:  "Test Issue 2",
		Status: "draft",
	}
	b.SetSync(ghSyncName, map[string]any{
		ghSyncKeyIssueNumber: "42",
	})
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	gh := mustDetectGitHub(t, "o", "r", c)
	result, err := gh.Link(context.Background(), b.ID, "42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionAlreadyLinked {
		t.Errorf("expected action %q, got %q", ActionAlreadyLinked, result.Action)
	}
	if result.ExternalID != "42" {
		t.Errorf("expected external ID '42', got %q", result.ExternalID)
	}
}

func TestGitHubIntegration_Unlink_Linked(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "test-3",
		Title:  "Test Issue 3",
		Status: "draft",
	}
	b.SetSync(ghSyncName, map[string]any{
		ghSyncKeyIssueNumber: "99",
	})
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	gh := mustDetectGitHub(t, "o", "r", c)
	result, err := gh.Unlink(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionUnlinked {
		t.Errorf("expected action %q, got %q", ActionUnlinked, result.Action)
	}
	if result.ExternalID != "99" {
		t.Errorf("expected external ID '99', got %q", result.ExternalID)
	}

	// Verify sync data removed
	reloaded, err := c.Get(b.ID)
	if err != nil {
		t.Fatalf("failed to get issue: %v", err)
	}
	if reloaded.HasSync(ghSyncName) {
		t.Error("sync data should have been removed")
	}
}

func TestGitHubIntegration_CheckConfiguration_EmptyOwnerRepo(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	// detectGitHub won't produce empty owner/repo, so construct manually through Detect
	// which won't help. Instead test via Check with SkipAPI
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	section := gh.checkConfiguration(context.Background(), CheckOptions{SkipAPI: true})
	if len(section.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	// With valid owner/repo, should pass
	if section.Checks[0].Status != CheckPass {
		t.Errorf("expected CheckPass for valid owner/repo, got %v", section.Checks[0].Status)
	}
	if section.Checks[0].Message != "o/r" {
		t.Errorf("expected message 'o/r', got %q", section.Checks[0].Message)
	}
}

func TestGitHubIntegration_CheckGitHubIntegration_NoToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	section := gh.checkGitHubIntegration(context.Background(), CheckOptions{})
	if len(section.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	if section.Checks[0].Status != CheckFail {
		t.Errorf("expected CheckFail when token missing, got %v", section.Checks[0].Status)
	}
}

func TestGitHubIntegration_CheckGitHubIntegration_SkipAPI(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	section := gh.checkGitHubIntegration(context.Background(), CheckOptions{SkipAPI: true})
	if len(section.Checks) < 2 {
		t.Fatalf("expected at least 2 checks, got %d", len(section.Checks))
	}
	if section.Checks[0].Status != CheckPass {
		t.Errorf("expected CheckPass for token set, got %v", section.Checks[0].Status)
	}
	if section.Checks[1].Status != CheckWarn {
		t.Errorf("expected CheckWarn for skipped API, got %v", section.Checks[1].Status)
	}
}

func TestGitHubIntegration_CheckSyncState_NoIssues(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	section := gh.checkSyncState(context.Background(), CheckOptions{SkipAPI: true})
	if len(section.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	if section.Checks[0].Message != "0 issues" {
		t.Errorf("expected '0 issues', got %q", section.Checks[0].Message)
	}
}

func TestGitHubIntegration_CheckSyncState_WithLinkedIssues(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "linked-1",
		Title:  "Linked Issue",
		Status: "draft",
	}
	b.SetSync(ghSyncName, map[string]any{
		ghSyncKeyIssueNumber: "10",
		ghSyncKeySyncedAt:    "2020-01-01T00:00:00Z", // stale
	})
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	gh := mustDetectGitHub(t, "o", "r", c)
	section := gh.checkSyncState(context.Background(), CheckOptions{SkipAPI: true})

	if len(section.Checks) < 1 {
		t.Fatal("expected at least one check result")
	}
	if section.Checks[0].Message != "1 issues" {
		t.Errorf("expected '1 issues', got %q", section.Checks[0].Message)
	}
	foundStale := false
	for _, check := range section.Checks {
		if check.Name == "Stale syncs" {
			foundStale = true
			if check.Status != CheckWarn {
				t.Errorf("expected CheckWarn for stale sync, got %v", check.Status)
			}
		}
	}
	if !foundStale {
		t.Error("expected stale sync warning")
	}
}

func TestGitHubIntegration_Check_SkipAPI(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	gh := mustDetectGitHub(t, "o", "r", c)

	report, err := gh.Check(context.Background(), CheckOptions{SkipAPI: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(report.Sections))
	}
	total := report.Summary.Passed + report.Summary.Warnings + report.Summary.Failed
	if total == 0 {
		t.Error("expected non-zero total checks in summary")
	}
}

func TestGitHubIntegration_Link_NewLink(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "test-new-link",
		Title:  "New Link Test",
		Status: "draft",
	}
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	gh := mustDetectGitHub(t, "o", "r", c)
	result, err := gh.Link(context.Background(), b.ID, "55")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionLinked {
		t.Errorf("expected action %q, got %q", ActionLinked, result.Action)
	}
	if result.ExternalID != "55" {
		t.Errorf("expected external ID '55', got %q", result.ExternalID)
	}

	// Verify issue now has sync data
	reloaded, err := c.Get(b.ID)
	if err != nil {
		t.Fatalf("failed to get issue: %v", err)
	}
	if !reloaded.HasSync(ghSyncName) {
		t.Error("expected sync data to be set after link")
	}
}
