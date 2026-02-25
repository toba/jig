package changelog

import (
	"testing"
	"time"

	"github.com/toba/jig/internal/todo/issue"
)

func tp(t time.Time) *time.Time { return &t }

func TestGather(t *testing.T) {
	now := time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC)
	since := now.AddDate(0, 0, -7)

	issues := []*issue.Issue{
		{
			ID: "created-in-range", Title: "New feature", Type: "feature", Status: "ready",
			CreatedAt: tp(now.AddDate(0, 0, -3)),
			UpdatedAt: tp(now.AddDate(0, 0, -3)),
		},
		{
			ID: "updated-in-range", Title: "Updated task", Type: "task", Status: "in-progress",
			CreatedAt: tp(now.AddDate(0, 0, -30)),
			UpdatedAt: tp(now.AddDate(0, 0, -1)),
		},
		{
			ID: "completed-in-range", Title: "Fixed bug", Type: "bug", Status: "completed",
			CreatedAt: tp(now.AddDate(0, 0, -20)),
			UpdatedAt: tp(now.AddDate(0, 0, -2)),
		},
		{
			ID: "outside-range", Title: "Old issue", Type: "task", Status: "ready",
			CreatedAt: tp(now.AddDate(0, 0, -60)),
			UpdatedAt: tp(now.AddDate(0, 0, -30)),
		},
		{
			ID: "created-and-completed", Title: "Quick fix", Type: "bug", Status: "completed",
			CreatedAt: tp(now.AddDate(0, 0, -1)),
			UpdatedAt: tp(now.AddDate(0, 0, -1)),
		},
	}

	result := Gather(issues, Options{Since: since, Until: now})

	// completed-in-range and created-and-completed should be in completed
	if len(result.Issues.Completed) != 2 {
		t.Errorf("expected 2 completed, got %d", len(result.Issues.Completed))
	}

	// created-in-range should be in created (not completed since status != completed)
	if len(result.Issues.Created) != 1 {
		t.Errorf("expected 1 created, got %d", len(result.Issues.Created))
	}
	if len(result.Issues.Created) > 0 && result.Issues.Created[0].ID != "created-in-range" {
		t.Errorf("expected created-in-range, got %s", result.Issues.Created[0].ID)
	}

	// updated-in-range should be in updated (created outside range, not completed)
	if len(result.Issues.Updated) != 1 {
		t.Errorf("expected 1 updated, got %d", len(result.Issues.Updated))
	}
	if len(result.Issues.Updated) > 0 && result.Issues.Updated[0].ID != "updated-in-range" {
		t.Errorf("expected updated-in-range, got %s", result.Issues.Updated[0].ID)
	}
}

func TestGather_EmptyInput(t *testing.T) {
	now := time.Now()
	result := Gather(nil, Options{Since: now.AddDate(0, 0, -7), Until: now})

	if len(result.Issues.Created) != 0 {
		t.Errorf("expected 0 created, got %d", len(result.Issues.Created))
	}
	if len(result.Issues.Updated) != 0 {
		t.Errorf("expected 0 updated, got %d", len(result.Issues.Updated))
	}
	if len(result.Issues.Completed) != 0 {
		t.Errorf("expected 0 completed, got %d", len(result.Issues.Completed))
	}
}

func TestGather_NilTimestamps(t *testing.T) {
	now := time.Now()
	issues := []*issue.Issue{
		{ID: "no-dates", Title: "No timestamps", Status: "ready"},
	}
	result := Gather(issues, Options{Since: now.AddDate(0, 0, -7), Until: now})

	total := len(result.Issues.Created) + len(result.Issues.Updated) + len(result.Issues.Completed)
	if total != 0 {
		t.Errorf("expected 0 issues, got %d", total)
	}
}

func TestGather_RangeIsSet(t *testing.T) {
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	result := Gather(nil, Options{Since: since, Until: until})

	if !result.Range.Since.Equal(since) {
		t.Errorf("expected since %v, got %v", since, result.Range.Since)
	}
	if !result.Range.Until.Equal(until) {
		t.Errorf("expected until %v, got %v", until, result.Range.Until)
	}
}
