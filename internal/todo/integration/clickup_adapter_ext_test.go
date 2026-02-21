package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/integration/clickup"
	"github.com/toba/jig/internal/todo/issue"
)

func TestClickUpIntegration_Name(t *testing.T) {
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, nil)
	if cu.Name() != "clickup" {
		t.Errorf("expected 'clickup', got %q", cu.Name())
	}
}

func TestClickUpIntegration_GetToken_Missing(t *testing.T) {
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, nil)
	t.Setenv("CLICKUP_TOKEN", "")
	_, err := cu.getToken()
	if err == nil {
		t.Fatal("expected error when CLICKUP_TOKEN not set")
	}
}

func TestClickUpIntegration_GetToken_Set(t *testing.T) {
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, nil)
	t.Setenv("CLICKUP_TOKEN", "pk_test_token")
	token, err := cu.getToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "pk_test_token" {
		t.Errorf("expected 'pk_test_token', got %q", token)
	}
}

func TestClickUpIntegration_Sync_NoToken(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)

	_, err := cu.Sync(context.Background(), nil, SyncOptions{})
	if err == nil {
		t.Fatal("expected error when CLICKUP_TOKEN not set")
	}
}

func TestClickUpIntegration_Link_IssueNotFound(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)

	_, err := cu.Link(context.Background(), "nonexistent", "abc123")
	if err == nil {
		t.Fatal("expected error for nonexistent issue")
	}
}

func TestClickUpIntegration_Unlink_IssueNotFound(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)

	_, err := cu.Unlink(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent issue")
	}
}

func TestClickUpIntegration_Unlink_NotLinked(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "cu-test-1",
		Title:  "Test Issue",
		Status: "draft",
	}
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)
	result, err := cu.Unlink(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionNotLinked {
		t.Errorf("expected action %q, got %q", ActionNotLinked, result.Action)
	}
}

func TestClickUpIntegration_Link_AlreadyLinked(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "cu-test-2",
		Title:  "Test Issue 2",
		Status: "draft",
	}
	b.SetSync(clickup.SyncName, map[string]any{
		clickup.SyncKeyTaskID: "task_abc",
	})
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)
	result, err := cu.Link(context.Background(), b.ID, "task_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionAlreadyLinked {
		t.Errorf("expected action %q, got %q", ActionAlreadyLinked, result.Action)
	}
	if result.ExternalID != "task_abc" {
		t.Errorf("expected external ID 'task_abc', got %q", result.ExternalID)
	}
}

func TestClickUpIntegration_Unlink_Linked(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "cu-test-3",
		Title:  "Test Issue 3",
		Status: "draft",
	}
	b.SetSync(clickup.SyncName, map[string]any{
		clickup.SyncKeyTaskID: "task_xyz",
	})
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)
	result, err := cu.Unlink(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Action != ActionUnlinked {
		t.Errorf("expected action %q, got %q", ActionUnlinked, result.Action)
	}
	if result.ExternalID != "task_xyz" {
		t.Errorf("expected external ID 'task_xyz', got %q", result.ExternalID)
	}

	// Verify sync data removed
	reloaded, err := c.Get(b.ID)
	if err != nil {
		t.Fatalf("failed to get issue: %v", err)
	}
	if reloaded.HasSync(clickup.SyncName) {
		t.Error("sync data should have been removed")
	}
}

func TestClickUpIntegration_CheckConfiguration_EmptyListID(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{
		ListID:          "",
		StatusMapping:   clickup.DefaultStatusMapping,
		PriorityMapping: clickup.DefaultPriorityMapping,
	}, c)
	section := cu.checkConfiguration(context.Background(), CheckOptions{SkipAPI: true})

	if len(section.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	foundListCheck := false
	for _, check := range section.Checks {
		if check.Name == "List ID configured" {
			foundListCheck = true
			if check.Status != CheckFail {
				t.Errorf("expected CheckFail for empty list_id, got %v", check.Status)
			}
		}
	}
	if !foundListCheck {
		t.Error("expected 'List ID configured' check")
	}
}

func TestClickUpIntegration_CheckConfiguration_ValidListID(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{
		ListID:          "12345",
		StatusMapping:   clickup.DefaultStatusMapping,
		PriorityMapping: clickup.DefaultPriorityMapping,
	}, c)
	section := cu.checkConfiguration(context.Background(), CheckOptions{SkipAPI: true})

	foundListCheck := false
	for _, check := range section.Checks {
		if check.Name == "List ID configured" {
			foundListCheck = true
			if check.Status != CheckPass {
				t.Errorf("expected CheckPass for valid list_id, got %v", check.Status)
			}
			if check.Message != "12345" {
				t.Errorf("expected message '12345', got %q", check.Message)
			}
		}
	}
	if !foundListCheck {
		t.Error("expected 'List ID configured' check")
	}
}

func TestClickUpIntegration_CheckClickUpIntegration_NoToken(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, nil)
	section := cu.checkClickUpIntegration(context.Background(), CheckOptions{})

	if len(section.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	if section.Checks[0].Status != CheckFail {
		t.Errorf("expected CheckFail when token missing, got %v", section.Checks[0].Status)
	}
}

func TestClickUpIntegration_CheckClickUpIntegration_SkipAPI(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "pk_test_token")
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, nil)
	section := cu.checkClickUpIntegration(context.Background(), CheckOptions{SkipAPI: true})

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

func TestClickUpIntegration_CheckSyncState_NoIssues(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)

	section := cu.checkSyncState(context.Background(), CheckOptions{SkipAPI: true})
	if len(section.Checks) == 0 {
		t.Fatal("expected at least one check result")
	}
	if section.Checks[0].Message != "0 issues" {
		t.Errorf("expected '0 issues', got %q", section.Checks[0].Message)
	}
}

func TestClickUpIntegration_CheckSyncState_WithLinkedIssues(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	dir := t.TempDir()
	c := core.New(dir, cfg)

	b := &issue.Issue{
		ID:     "cu-linked-1",
		Title:  "Linked ClickUp Issue",
		Status: "draft",
	}
	b.SetSync(clickup.SyncName, map[string]any{
		clickup.SyncKeyTaskID:   "task_stale",
		clickup.SyncKeySyncedAt: "2020-01-01T00:00:00Z", // stale
	})
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create issue: %v", err)
	}

	cu := newClickUpIntegration(&clickup.Config{ListID: "123"}, c)
	section := cu.checkSyncState(context.Background(), CheckOptions{SkipAPI: true})

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

func TestClickUpIntegration_Check_SkipAPI(t *testing.T) {
	t.Setenv("CLICKUP_TOKEN", "")
	cfg := config.Default()
	c := core.New(t.TempDir(), cfg)
	cu := newClickUpIntegration(&clickup.Config{
		ListID:          "123",
		StatusMapping:   clickup.DefaultStatusMapping,
		PriorityMapping: clickup.DefaultPriorityMapping,
	}, c)

	report, err := cu.Check(context.Background(), CheckOptions{SkipAPI: true})
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

func TestConvertClickUpResult(t *testing.T) {
	cuResult := clickup.SyncResult{
		IssueID:    "issue-1",
		IssueTitle: "Test Issue",
		TaskID:     "task_abc",
		TaskURL:    "https://app.clickup.com/t/task_abc",
		Action:     "created",
		Error:      nil,
	}
	result := convertClickUpResult(cuResult)
	if result.IssueID != "issue-1" {
		t.Errorf("expected issue ID 'issue-1', got %q", result.IssueID)
	}
	if result.IssueTitle != "Test Issue" {
		t.Errorf("expected title 'Test Issue', got %q", result.IssueTitle)
	}
	if result.ExternalID != "task_abc" {
		t.Errorf("expected external ID 'task_abc', got %q", result.ExternalID)
	}
	if result.ExternalURL != "https://app.clickup.com/t/task_abc" {
		t.Errorf("expected URL, got %q", result.ExternalURL)
	}
	if result.Action != "created" {
		t.Errorf("expected action 'created', got %q", result.Action)
	}
}

func TestValidStatusList(t *testing.T) {
	cfg := config.Default()
	result := validStatusList(cfg)
	if result == "" {
		t.Error("expected non-empty status list")
	}
	// Should contain at least "draft" and "completed"
	if !strings.Contains(result, "draft") {
		t.Errorf("expected status list to contain 'draft', got %q", result)
	}
	if !strings.Contains(result, "completed") {
		t.Errorf("expected status list to contain 'completed', got %q", result)
	}
}
