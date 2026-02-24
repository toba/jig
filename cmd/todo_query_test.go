package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/issue"
)

func setupQueryTestCore(t *testing.T) (*core.Core, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, ".issues")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to create test .issues dir: %v", err)
	}

	cfg := todoconfig.Default()
	testCore := core.New(dataDir, cfg)
	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	oldStore := todoStore
	todoStore = testCore

	cleanup := func() {
		todoStore = oldStore
	}

	return testCore, cleanup
}

func createQueryTestIssue(t *testing.T, c *core.Core, id, title, status string) {
	t.Helper()
	b := &issue.Issue{
		ID:     id,
		Slug:   issue.Slugify(title),
		Title:  title,
		Status: status,
	}
	if err := c.Create(b); err != nil {
		t.Fatalf("failed to create test issue: %v", err)
	}
}

func TestExecuteQuery(t *testing.T) {
	testCore, cleanup := setupQueryTestCore(t)
	defer cleanup()

	createQueryTestIssue(t, testCore, "test-1", "First Issue", "todo")
	createQueryTestIssue(t, testCore, "test-2", "Second Issue", "in-progress")
	createQueryTestIssue(t, testCore, "test-3", "Third Issue", "completed")

	t.Run("basic query all issues", func(t *testing.T) {
		query := `{ issues { id title status } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID     string `json:"id"`
				Title  string `json:"title"`
				Status string `json:"status"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 3 {
			t.Errorf("expected 3 issues, got %d", len(data.Issues))
		}
	})

	t.Run("query single issue by id", func(t *testing.T) {
		query := `{ issue(id: "test-1") { id title } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if data.Issue.ID != "test-1" {
			t.Errorf("expected id 'test-1', got %q", data.Issue.ID)
		}
	})

	t.Run("query with filter", func(t *testing.T) {
		query := `{ issues(filter: { status: ["todo"] }) { id } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID string `json:"id"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 1 {
			t.Errorf("expected 1 issue with status 'todo', got %d", len(data.Issues))
		}
	})

	t.Run("query with variables", func(t *testing.T) {
		query := `query GetIssue($id: ID!) { issue(id: $id) { id title } }`
		variables := map[string]any{
			"id": "test-2",
		}
		result, err := executeQuery(query, variables, "GetIssue")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if data.Issue.ID != "test-2" {
			t.Errorf("expected id 'test-2', got %q", data.Issue.ID)
		}
	})

	t.Run("query nonexistent issue returns null", func(t *testing.T) {
		query := `{ issue(id: "nonexistent") { id } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue *struct {
				ID string `json:"id"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if data.Issue != nil {
			t.Errorf("expected null issue, got %+v", data.Issue)
		}
	})

	t.Run("invalid query returns error", func(t *testing.T) {
		query := `{ invalid { field } }`
		_, err := executeQuery(query, nil, "")
		if err == nil {
			t.Fatal("expected error for invalid query, got nil")
		}
		if !strings.Contains(err.Error(), "graphql") {
			t.Errorf("expected error to contain 'graphql', got %q", err.Error())
		}
	})
}

func TestExecuteQueryWithRelationships(t *testing.T) {
	testCore, cleanup := setupQueryTestCore(t)
	defer cleanup()

	parent := &issue.Issue{
		ID:     "parent-1",
		Slug:   "parent-issue",
		Title:  "Parent Issue",
		Status: "todo",
	}
	testCore.Create(parent)

	child := &issue.Issue{
		ID:     "child-1",
		Slug:   "child-issue",
		Title:  "Child Issue",
		Status: "todo",
		Parent: "parent-1",
	}
	testCore.Create(child)

	blocker := &issue.Issue{
		ID:       "blocker-1",
		Slug:     "blocker-issue",
		Title:    "Blocker Issue",
		Status:   "todo",
		Blocking: []string{"child-1"},
	}
	testCore.Create(blocker)

	t.Run("query parent relationship", func(t *testing.T) {
		query := `{ issue(id: "child-1") { id parent { id title } } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue struct {
				ID     string `json:"id"`
				Parent *struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"parent"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if data.Issue.Parent == nil {
			t.Fatal("expected parent to be set")
		}
		if data.Issue.Parent.ID != "parent-1" {
			t.Errorf("expected parent id 'parent-1', got %q", data.Issue.Parent.ID)
		}
	})

	t.Run("query children relationship", func(t *testing.T) {
		query := `{ issue(id: "parent-1") { id children { id title } } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue struct {
				ID       string `json:"id"`
				Children []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"children"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issue.Children) != 1 {
			t.Errorf("expected 1 child, got %d", len(data.Issue.Children))
		}
	})

	t.Run("query blockedBy relationship", func(t *testing.T) {
		query := `{ issue(id: "child-1") { id blockedBy { id title } } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue struct {
				ID        string `json:"id"`
				BlockedBy []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"blockedBy"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issue.BlockedBy) != 1 {
			t.Errorf("expected 1 blocker, got %d", len(data.Issue.BlockedBy))
		}
	})

	t.Run("query blocking relationship", func(t *testing.T) {
		query := `{ issue(id: "blocker-1") { id blocking { id title } } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issue struct {
				ID       string `json:"id"`
				Blocking []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"blocking"`
			} `json:"issue"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issue.Blocking) != 1 {
			t.Errorf("expected 1 blocked issue, got %d", len(data.Issue.Blocking))
		}
	})
}

func TestExecuteQueryWithFilters(t *testing.T) {
	testCore, cleanup := setupQueryTestCore(t)
	defer cleanup()

	testCore.Create(&issue.Issue{ID: "bug-1", Slug: "bug-one", Title: "Bug One", Status: "todo", Type: "bug", Priority: "critical", Tags: []string{"frontend"}})
	testCore.Create(&issue.Issue{ID: "feat-1", Slug: "feature-one", Title: "Feature One", Status: "in-progress", Type: "feature", Priority: "high", Tags: []string{"backend"}})
	testCore.Create(&issue.Issue{ID: "task-1", Slug: "task-one", Title: "Task One", Status: "completed", Type: "task", Priority: "normal", Tags: []string{"frontend", "backend"}})

	t.Run("filter by type", func(t *testing.T) {
		query := `{ issues(filter: { type: ["bug"] }) { id type } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 1 {
			t.Errorf("expected 1 issue with type 'bug', got %d", len(data.Issues))
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		query := `{ issues(filter: { priority: ["critical", "high"] }) { id priority } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID       string `json:"id"`
				Priority string `json:"priority"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 2 {
			t.Errorf("expected 2 issues with priority 'critical' or 'high', got %d", len(data.Issues))
		}
	})

	t.Run("filter by tags", func(t *testing.T) {
		query := `{ issues(filter: { tags: ["frontend"] }) { id tags } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID   string   `json:"id"`
				Tags []string `json:"tags"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 2 {
			t.Errorf("expected 2 issues with tag 'frontend', got %d", len(data.Issues))
		}
	})

	t.Run("exclude by status", func(t *testing.T) {
		query := `{ issues(filter: { excludeStatus: ["completed"] }) { id status } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 2 {
			t.Errorf("expected 2 issues (excluding completed), got %d", len(data.Issues))
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		query := `{ issues(filter: { status: ["todo", "in-progress"], type: ["bug", "feature"] }) { id } }`
		result, err := executeQuery(query, nil, "")
		if err != nil {
			t.Fatalf("executeQuery() error = %v", err)
		}

		var data struct {
			Issues []struct {
				ID string `json:"id"`
			} `json:"issues"`
		}

		if err := json.Unmarshal(result, &data); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if len(data.Issues) != 2 {
			t.Errorf("expected 2 issues matching combined filters, got %d", len(data.Issues))
		}
	})
}

func TestGetGraphQLSchema(t *testing.T) {
	_, cleanup := setupQueryTestCore(t)
	defer cleanup()

	schema := GetGraphQLSchema()

	expectedTypes := []string{
		"type Query",
		"type Issue",
		"input IssueFilter",
	}

	for _, expected := range expectedTypes {
		if !strings.Contains(schema, expected) {
			t.Errorf("schema missing expected type: %s", expected)
		}
	}

	expectedFields := []string{
		"issue(id: ID!)",
		"issues(filter: IssueFilter)",
		"blockedBy",
		"blocking",
		"parent",
		"children",
	}

	for _, expected := range expectedFields {
		if !strings.Contains(schema, expected) {
			t.Errorf("schema missing expected field: %s", expected)
		}
	}
}

func TestReadFromStdin(t *testing.T) {
	t.Run("returns empty when stdin is terminal", func(t *testing.T) {
		result, err := readFromStdin()
		if err != nil {
			t.Fatalf("readFromStdin() error = %v", err)
		}
		if result != "" {
			t.Logf("readFromStdin() returned %q (may vary by test environment)", result)
		}
	})
}
