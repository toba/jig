package graph

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/toba/jig/internal/todo/graph/model"
	"github.com/toba/jig/internal/todo/issue"
)

func TestApplyFilterNil(t *testing.T) {
	issues := []*issue.Issue{
		{ID: "a"}, {ID: "b"},
	}
	result := ApplyFilter(issues, nil, nil)
	if len(result) != 2 {
		t.Errorf("ApplyFilter(nil) count = %d, want 2", len(result))
	}
}

func TestFilterByTypeAndExcludeType(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	c.Create(&issue.Issue{ID: "t1", Title: "Bug", Status: "todo", Type: "bug"})
	c.Create(&issue.Issue{ID: "t2", Title: "Task", Status: "todo", Type: "task"})
	c.Create(&issue.Issue{ID: "t3", Title: "Feature", Status: "todo", Type: "feature"})

	qr := resolver.Query()

	t.Run("filter by type", func(t *testing.T) {
		filter := &model.IssueFilter{Type: []string{"bug"}}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 1 || got[0].ID != "t1" {
			t.Errorf("Issues(type=bug) = %v, want [t1]", ids(got))
		}
	})

	t.Run("filter by multiple types", func(t *testing.T) {
		filter := &model.IssueFilter{Type: []string{"bug", "feature"}}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 2 {
			t.Errorf("Issues(type=bug,feature) count = %d, want 2", len(got))
		}
	})

	t.Run("exclude type", func(t *testing.T) {
		filter := &model.IssueFilter{ExcludeType: []string{"task"}}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 2 {
			t.Errorf("Issues(excludeType=task) count = %d, want 2", len(got))
		}
		for _, b := range got {
			if b.Type == "task" {
				t.Error("should not include task type")
			}
		}
	})
}

func TestFilterByBlockingID(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	c.Create(&issue.Issue{ID: "target-1", Title: "Target", Status: "todo"})
	c.Create(&issue.Issue{ID: "blocker-a", Title: "Blocker A", Status: "todo", Blocking: []string{"target-1"}})
	c.Create(&issue.Issue{ID: "other-1", Title: "Other", Status: "todo"})

	qr := resolver.Query()

	t.Run("filter by blockingId", func(t *testing.T) {
		targetID := "target-1"
		filter := &model.IssueFilter{BlockingID: &targetID}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 1 || got[0].ID != "blocker-a" {
			t.Errorf("Issues(blockingId=target-1) = %v, want [blocker-a]", ids(got))
		}
	})
}

func TestFilterByNoBlocking(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	c.Create(&issue.Issue{ID: "has-blocking", Title: "Has", Status: "todo", Blocking: []string{"x"}})
	c.Create(&issue.Issue{ID: "no-blocking", Title: "None", Status: "todo"})

	qr := resolver.Query()

	noBlocking := true
	filter := &model.IssueFilter{NoBlocking: &noBlocking}
	got, err := qr.Issues(ctx, filter)
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "no-blocking" {
		t.Errorf("Issues(noBlocking) = %v, want [no-blocking]", ids(got))
	}
}

func TestFilterByHasBlockedBy(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	c.Create(&issue.Issue{ID: "has-bb", Title: "Has", Status: "todo", BlockedBy: []string{"x"}})
	c.Create(&issue.Issue{ID: "no-bb", Title: "None", Status: "todo"})

	qr := resolver.Query()

	t.Run("hasBlockedBy true", func(t *testing.T) {
		hasBB := true
		filter := &model.IssueFilter{HasBlockedBy: &hasBB}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 1 || got[0].ID != "has-bb" {
			t.Errorf("Issues(hasBlockedBy) = %v, want [has-bb]", ids(got))
		}
	})

	t.Run("noBlockedBy true", func(t *testing.T) {
		noBB := true
		filter := &model.IssueFilter{NoBlockedBy: &noBB}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 1 || got[0].ID != "no-bb" {
			t.Errorf("Issues(noBlockedBy) = %v, want [no-bb]", ids(got))
		}
	})

	t.Run("blockedById", func(t *testing.T) {
		bbID := "x"
		filter := &model.IssueFilter{BlockedByID: &bbID}
		got, err := qr.Issues(ctx, filter)
		if err != nil {
			t.Fatalf("Issues() error = %v", err)
		}
		if len(got) != 1 || got[0].ID != "has-bb" {
			t.Errorf("Issues(blockedById=x) = %v, want [has-bb]", ids(got))
		}
	})
}

func TestFilterCombination(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	c.Create(&issue.Issue{ID: "match", Title: "Match", Status: "todo", Type: "bug", Priority: "high", Tags: []string{"urgent"}})
	c.Create(&issue.Issue{ID: "wrong-status", Title: "Wrong", Status: "completed", Type: "bug", Priority: "high"})
	c.Create(&issue.Issue{ID: "wrong-type", Title: "Wrong", Status: "todo", Type: "task", Priority: "high"})
	c.Create(&issue.Issue{ID: "wrong-priority", Title: "Wrong", Status: "todo", Type: "bug", Priority: "low"})

	qr := resolver.Query()

	filter := &model.IssueFilter{
		Status:   []string{"todo"},
		Type:     []string{"bug"},
		Priority: []string{"high"},
		Tags:     []string{"urgent"},
	}
	got, err := qr.Issues(ctx, filter)
	if err != nil {
		t.Fatalf("Issues() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "match" {
		t.Errorf("combined filter = %v, want [match]", ids(got))
	}
}

func TestResolverIssueFieldResolvers(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	b := &issue.Issue{
		ID:        "field-test",
		Title:     "Test",
		Status:    "todo",
		Parent:    "some-parent",
		Blocking:  []string{"target-1"},
		BlockedBy: []string{"blocker-1"},
	}
	c.Create(b)

	br := resolver.Issue()

	t.Run("parentId returns pointer", func(t *testing.T) {
		got, err := br.ParentID(ctx, b)
		if err != nil {
			t.Fatalf("ParentID() error = %v", err)
		}
		if got == nil || *got != "some-parent" {
			t.Errorf("ParentID() = %v, want pointer to \"some-parent\"", got)
		}
	})

	t.Run("parentId returns nil for empty", func(t *testing.T) {
		noParent := &issue.Issue{ID: "no-parent", Title: "Test", Status: "todo"}
		c.Create(noParent)
		got, err := br.ParentID(ctx, noParent)
		if err != nil {
			t.Fatalf("ParentID() error = %v", err)
		}
		if got != nil {
			t.Errorf("ParentID() = %q, want nil", *got)
		}
	})

	t.Run("blockingIds", func(t *testing.T) {
		got, err := br.BlockingIds(ctx, b)
		if err != nil {
			t.Fatalf("BlockingIds() error = %v", err)
		}
		if len(got) != 1 || got[0] != "target-1" {
			t.Errorf("BlockingIds() = %v, want [target-1]", got)
		}
	})

	t.Run("blockedByIds", func(t *testing.T) {
		got, err := br.BlockedByIds(ctx, b)
		if err != nil {
			t.Fatalf("BlockedByIds() error = %v", err)
		}
		if len(got) != 1 || got[0] != "blocker-1" {
			t.Errorf("BlockedByIds() = %v, want [blocker-1]", got)
		}
	})
}

func TestResolverCreateIssueWithDueEmpty(t *testing.T) {
	resolver, _ := setupTestResolver(t)
	ctx := context.Background()

	mr := resolver.Mutation()
	emptyDue := ""
	input := model.CreateIssueInput{
		Title: "No Due",
		Due:   &emptyDue,
	}
	got, err := mr.CreateIssue(ctx, input)
	if err != nil {
		t.Fatalf("CreateIssue() error = %v", err)
	}
	if got.Due != nil {
		t.Errorf("Due = %v, want nil for empty due string", got.Due)
	}
}

func TestResolverUpdateIssueDueInvalid(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	b := &issue.Issue{ID: "due-invalid", Title: "Test", Status: "todo"}
	c.Create(b)

	mr := resolver.Mutation()
	badDue := "not-a-date"
	input := model.UpdateIssueInput{Due: &badDue}
	_, err := mr.UpdateIssue(ctx, "due-invalid", input)
	if err == nil {
		t.Error("UpdateIssue() should fail with invalid due date")
	}
}

func TestResolverValidateETag(t *testing.T) {
	resolver, _ := setupTestResolver(t)

	b := &issue.Issue{ID: "etag-val", Title: "Test", Status: "todo"}

	t.Run("nil ifMatch passes", func(t *testing.T) {
		err := resolver.validateETag(b, nil)
		if err != nil {
			t.Errorf("validateETag(nil) = %v, want nil", err)
		}
	})

	t.Run("empty ifMatch passes", func(t *testing.T) {
		empty := ""
		err := resolver.validateETag(b, &empty)
		if err != nil {
			t.Errorf("validateETag(\"\") = %v, want nil", err)
		}
	})

	t.Run("correct ifMatch passes", func(t *testing.T) {
		etag := b.ETag()
		err := resolver.validateETag(b, &etag)
		if err != nil {
			t.Errorf("validateETag(correct) = %v, want nil", err)
		}
	})

	t.Run("wrong ifMatch fails", func(t *testing.T) {
		wrong := "wrong"
		err := resolver.validateETag(b, &wrong)
		if err == nil {
			t.Error("validateETag(wrong) should fail")
		}
	})
}

func TestResolverValidateAndSetParent(t *testing.T) {
	resolver, c := setupTestResolver(t)

	epic := &issue.Issue{ID: "epic-vp", Title: "Epic", Type: "epic", Status: "todo"}
	task := &issue.Issue{ID: "task-vp", Title: "Task", Type: "task", Status: "todo"}
	c.Create(epic)
	c.Create(task)

	t.Run("empty parent clears", func(t *testing.T) {
		b := &issue.Issue{ID: "test-vp", Parent: "epic-vp", Type: "task"}
		err := resolver.validateAndSetParent(b, "")
		if err != nil {
			t.Fatalf("validateAndSetParent(\"\") error = %v", err)
		}
		if b.Parent != "" {
			t.Errorf("Parent = %q, want empty", b.Parent)
		}
	})

	t.Run("valid parent sets", func(t *testing.T) {
		b := &issue.Issue{ID: "test-vp2", Type: "task"}
		c.Create(b)
		err := resolver.validateAndSetParent(b, "epic-vp")
		if err != nil {
			t.Fatalf("validateAndSetParent(epic-vp) error = %v", err)
		}
		if b.Parent != "epic-vp" {
			t.Errorf("Parent = %q, want epic-vp", b.Parent)
		}
	})
}

func TestResolverValidateAndAddBlocking(t *testing.T) {
	resolver, c := setupTestResolver(t)

	task1 := &issue.Issue{ID: "vab-1", Title: "Task 1", Type: "task", Status: "todo"}
	task2 := &issue.Issue{ID: "vab-2", Title: "Task 2", Type: "task", Status: "todo"}
	c.Create(task1)
	c.Create(task2)

	t.Run("valid blocking", func(t *testing.T) {
		err := resolver.validateAndAddBlocking(task1, []string{"vab-2"})
		if err != nil {
			t.Fatalf("validateAndAddBlocking error = %v", err)
		}
		if len(task1.Blocking) == 0 || task1.Blocking[0] != "vab-2" {
			t.Errorf("Blocking = %v, want [vab-2]", task1.Blocking)
		}
	})

	t.Run("self-blocking fails", func(t *testing.T) {
		b := &issue.Issue{ID: "vab-self", Title: "Self", Type: "task", Status: "todo"}
		c.Create(b)
		err := resolver.validateAndAddBlocking(b, []string{"vab-self"})
		if err == nil {
			t.Error("should fail for self-blocking")
		}
	})

	t.Run("nonexistent target fails", func(t *testing.T) {
		b := &issue.Issue{ID: "vab-ne", Title: "NE", Type: "task", Status: "todo"}
		c.Create(b)
		err := resolver.validateAndAddBlocking(b, []string{"nonexistent"})
		if err == nil {
			t.Error("should fail for nonexistent target")
		}
	})
}

func TestResolverRemoveBlockingRelationships(t *testing.T) {
	resolver, c := setupTestResolver(t)

	task := &issue.Issue{ID: "rem-b", Title: "Task", Type: "task", Status: "todo", Blocking: []string{"t1", "t2"}}
	c.Create(task)
	c.Create(&issue.Issue{ID: "t1", Title: "T1", Status: "todo"})
	c.Create(&issue.Issue{ID: "t2", Title: "T2", Status: "todo"})

	resolver.removeBlockingRelationships(task, []string{"t1"})
	if len(task.Blocking) != 1 || task.Blocking[0] != "t2" {
		t.Errorf("Blocking = %v, want [t2]", task.Blocking)
	}
}

func TestResolverValidateAndAddBlockedBy(t *testing.T) {
	resolver, c := setupTestResolver(t)

	task1 := &issue.Issue{ID: "vabb-1", Title: "Task 1", Type: "task", Status: "todo"}
	task2 := &issue.Issue{ID: "vabb-2", Title: "Task 2", Type: "task", Status: "todo"}
	c.Create(task1)
	c.Create(task2)

	t.Run("valid blockedBy", func(t *testing.T) {
		err := resolver.validateAndAddBlockedBy(task1, []string{"vabb-2"})
		if err != nil {
			t.Fatalf("validateAndAddBlockedBy error = %v", err)
		}
	})

	t.Run("self-blockedBy fails", func(t *testing.T) {
		b := &issue.Issue{ID: "vabb-self", Title: "Self", Type: "task", Status: "todo"}
		c.Create(b)
		err := resolver.validateAndAddBlockedBy(b, []string{"vabb-self"})
		if err == nil {
			t.Error("should fail for self-blockedBy")
		}
	})

	t.Run("nonexistent blocker fails", func(t *testing.T) {
		b := &issue.Issue{ID: "vabb-ne", Title: "NE", Type: "task", Status: "todo"}
		c.Create(b)
		err := resolver.validateAndAddBlockedBy(b, []string{"nonexistent"})
		if err == nil {
			t.Error("should fail for nonexistent blocker")
		}
	})
}

func TestResolverRemoveBlockedByRelationships(t *testing.T) {
	resolver, c := setupTestResolver(t)

	task := &issue.Issue{ID: "rem-bb", Title: "Task", Type: "task", Status: "todo", BlockedBy: []string{"b1", "b2"}}
	c.Create(task)
	c.Create(&issue.Issue{ID: "b1", Title: "B1", Status: "todo"})
	c.Create(&issue.Issue{ID: "b2", Title: "B2", Status: "todo"})

	resolver.removeBlockedByRelationships(task, []string{"b1"})
	if len(task.BlockedBy) != 1 || task.BlockedBy[0] != "b2" {
		t.Errorf("BlockedBy = %v, want [b2]", task.BlockedBy)
	}
}

func TestCreateIssueWithParentValidation(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	epic := &issue.Issue{ID: "epic-cp", Title: "Epic", Type: "epic", Status: "todo"}
	c.Create(epic)

	mr := resolver.Mutation()

	t.Run("valid parent", func(t *testing.T) {
		parent := "epic-cp"
		input := model.CreateIssueInput{
			Title:  "Child Task",
			Parent: &parent,
		}
		got, err := mr.CreateIssue(ctx, input)
		if err != nil {
			t.Fatalf("CreateIssue() error = %v", err)
		}
		if got.Parent != "epic-cp" {
			t.Errorf("Parent = %q, want epic-cp", got.Parent)
		}
	})
}

func TestSetSyncDataWithETag(t *testing.T) {
	resolver, c := setupTestResolverWithRequireIfMatch(t)
	ctx := context.Background()

	b := &issue.Issue{ID: "sync-etag", Title: "Test", Status: "todo"}
	c.Create(b)

	mr := resolver.Mutation()
	etag := b.ETag()
	data := map[string]any{"key": "value"}
	got, err := mr.SetSyncData(ctx, "sync-etag", "test", data, &etag)
	if err != nil {
		t.Fatalf("SetSyncData() error = %v", err)
	}
	if !got.HasSync("test") {
		t.Error("sync data should be set")
	}
}

func TestRemoveSyncDataWithETag(t *testing.T) {
	resolver, c := setupTestResolverWithRequireIfMatch(t)
	ctx := context.Background()

	b := &issue.Issue{
		ID:     "rmsync-etag",
		Title:  "Test",
		Status: "todo",
		Sync:   map[string]map[string]any{"test": {"key": "value"}},
	}
	c.Create(b)

	mr := resolver.Mutation()
	etag := b.ETag()
	got, err := mr.RemoveSyncData(ctx, "rmsync-etag", "test", &etag)
	if err != nil {
		t.Fatalf("RemoveSyncData() error = %v", err)
	}
	if got.HasSync("test") {
		t.Error("sync data should be removed")
	}
}

func TestQueryIssueViaPartialID(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	// NormalizeID in Core supports prefix matching
	c.Create(&issue.Issue{ID: "abc-123", Title: "Test", Status: "todo"})

	qr := resolver.Query()

	// Exact match works
	got, err := qr.Issue(ctx, "abc-123")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if got == nil {
		t.Fatal("Issue(abc-123) returned nil")
	}
}

func TestUpdateIssueAddTagsDuplicate(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	task := &issue.Issue{ID: "dup-tags", Title: "Task", Type: "task", Status: "todo", Tags: []string{"existing"}}
	c.Create(task)

	input := model.UpdateIssueInput{
		AddTags: []string{"existing", "new"},
	}
	got, err := resolver.Mutation().UpdateIssue(ctx, "dup-tags", input)
	if err != nil {
		t.Fatalf("UpdateIssue() error = %v", err)
	}

	// "existing" should not be duplicated
	tagCount := make(map[string]int)
	for _, tag := range got.Tags {
		tagCount[tag]++
	}
	if tagCount["existing"] > 1 {
		t.Errorf("tag 'existing' appears %d times, want 1", tagCount["existing"])
	}
	if tagCount["new"] != 1 {
		t.Errorf("tag 'new' count = %d, want 1", tagCount["new"])
	}
}

func TestDeleteIssueWithParentLink(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	parent := &issue.Issue{ID: "parent-del", Title: "Parent", Type: "epic", Status: "todo"}
	child := &issue.Issue{ID: "child-del", Title: "Child", Type: "task", Status: "todo", Parent: "parent-del"}
	c.Create(parent)
	c.Create(child)

	mr := resolver.Mutation()
	_, err := mr.DeleteIssue(ctx, "parent-del")
	if err != nil {
		t.Fatalf("DeleteIssue() error = %v", err)
	}

	// Verify parent link is cleaned up on the child
	updated, err := c.Get("child-del")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if updated.Parent != "" {
		t.Errorf("child Parent = %q, want empty after parent deletion", updated.Parent)
	}
}

func TestResolverInterfaceMethods(t *testing.T) {
	resolver, _ := setupTestResolver(t)

	// Verify interface methods return non-nil resolvers
	if resolver.Issue() == nil {
		t.Error("Issue() should not be nil")
	}
	if resolver.Mutation() == nil {
		t.Error("Mutation() should not be nil")
	}
	if resolver.Query() == nil {
		t.Error("Query() should not be nil")
	}
}

func TestIsSyncStaleEdgeCases(t *testing.T) {
	t.Run("non-string synced_at returns stale", func(t *testing.T) {
		now := timeNow()
		b := &issue.Issue{
			ID:        "stale-type",
			UpdatedAt: &now,
			Sync: map[string]map[string]any{
				"test": {"synced_at": 12345}, // int, not string
			},
		}
		if !isSyncStale(b, "test") {
			t.Error("non-string synced_at should be treated as stale")
		}
	})

	t.Run("invalid RFC3339 returns stale", func(t *testing.T) {
		now := timeNow()
		b := &issue.Issue{
			ID:        "stale-parse",
			UpdatedAt: &now,
			Sync: map[string]map[string]any{
				"test": {"synced_at": "not-a-date"},
			},
		}
		if !isSyncStale(b, "test") {
			t.Error("unparseable synced_at should be treated as stale")
		}
	})
}

func TestCreateIssueBodyMutualExclusive(t *testing.T) {
	resolver, c := setupTestResolver(t)
	ctx := context.Background()

	b := &issue.Issue{ID: "excl-test", Title: "Test", Status: "todo", Body: "Original", Tags: []string{"a"}}
	c.Create(b)

	mr := resolver.Mutation()

	t.Run("tags and removeTags are mutually exclusive", func(t *testing.T) {
		input := model.UpdateIssueInput{
			Tags:       []string{"b"},
			RemoveTags: []string{"a"},
		}
		_, err := mr.UpdateIssue(ctx, "excl-test", input)
		if err == nil {
			t.Error("should fail with both tags and removeTags")
		}
		if !strings.Contains(err.Error(), "cannot specify both") {
			t.Errorf("error = %v, want 'cannot specify both'", err)
		}
	})
}

// helpers
func ids(issues []*issue.Issue) []string {
	out := make([]string, len(issues))
	for i, b := range issues {
		out[i] = b.ID
	}
	return out
}

func timeNow() time.Time {
	return time.Now().UTC()
}
