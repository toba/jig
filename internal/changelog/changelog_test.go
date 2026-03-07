package changelog

import (
	"strings"
	"testing"
	"time"

	"github.com/toba/jig/internal/todo/issue"
)

func TestGather(t *testing.T) {
	now := time.Date(2026, 2, 25, 12, 0, 0, 0, time.UTC)
	since := now.AddDate(0, 0, -7)

	issues := []*issue.Issue{
		{
			ID: "created-in-range", Title: "New feature", Type: "feature", Status: "ready",
			CreatedAt: new(now.AddDate(0, 0, -3)),
			UpdatedAt: new(now.AddDate(0, 0, -3)),
		},
		{
			ID: "updated-in-range", Title: "Updated task", Type: "task", Status: "in-progress",
			CreatedAt: new(now.AddDate(0, 0, -30)),
			UpdatedAt: new(now.AddDate(0, 0, -1)),
		},
		{
			ID: "completed-in-range", Title: "Fixed bug", Type: "bug", Status: "completed",
			CreatedAt: new(now.AddDate(0, 0, -20)),
			UpdatedAt: new(now.AddDate(0, 0, -2)),
		},
		{
			ID: "outside-range", Title: "Old issue", Type: "task", Status: "ready",
			CreatedAt: new(now.AddDate(0, 0, -60)),
			UpdatedAt: new(now.AddDate(0, 0, -30)),
		},
		{
			ID: "created-and-completed", Title: "Quick fix", Type: "bug", Status: "completed",
			CreatedAt: new(now.AddDate(0, 0, -1)),
			UpdatedAt: new(now.AddDate(0, 0, -1)),
		},
		{
			ID: "review-in-range", Title: "Pending review", Type: "feature", Status: "review",
			CreatedAt: new(now.AddDate(0, 0, -10)),
			UpdatedAt: new(now.AddDate(0, 0, -2)),
		},
	}

	result := Gather(issues, Options{Since: since, Until: now})

	// completed-in-range, created-and-completed, and review-in-range should be in completed
	if len(result.Issues.Completed) != 3 {
		t.Errorf("expected 3 completed, got %d", len(result.Issues.Completed))
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

func TestGather_SingleCommitRange(t *testing.T) {
	// When CommitTimeRange returns since == until (single commit), it now
	// extends until by 1 second. Verify that Gather finds issues at that timestamp.
	commitTime := time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC)
	since := commitTime
	until := commitTime.Add(time.Second) // simulates the fix in CommitTimeRange

	completed := &issue.Issue{
		ID: "done", Title: "Fixed it", Type: "bug", Status: "completed",
		CreatedAt: new(commitTime.AddDate(0, 0, -5)),
		UpdatedAt: new(commitTime),
	}
	result := Gather([]*issue.Issue{completed}, Options{Since: since, Until: until})

	if len(result.Issues.Completed) != 1 {
		t.Errorf("expected 1 completed issue, got %d", len(result.Issues.Completed))
	}
}

func TestFormatMarkdown_GroupsByType(t *testing.T) {
	r := &Result{
		GitHub: "https://github.com/owner/repo",
		Range: TimeRange{
			Since: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), // Sunday
			Until: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{ID: "feat-1", Title: "Add widget support", Type: "feature"},
				{ID: "bug-1", Title: "Fix crash on startup", Type: "bug"},
				{ID: "task-1", Title: "Update dependencies", Type: "task"},
				{ID: "feat-2", Title: "New dashboard", Type: "feature"},
			},
		},
	}

	md := FormatMarkdown(r, MarkdownOptions{Mode: "weekly"}, "")

	if !strings.Contains(md, "### ✨ Features") {
		t.Error("missing Features heading")
	}
	if !strings.Contains(md, "### 🐞 Fixes") {
		t.Error("missing Fixes heading")
	}
	if !strings.Contains(md, "### 🗜️ Tweaks") {
		t.Error("missing Tweaks heading")
	}
	if !strings.Contains(md, "- Add widget support") {
		t.Error("missing feature entry")
	}
	if !strings.Contains(md, "- Fix crash on startup") {
		t.Error("missing bug entry")
	}
	if !strings.Contains(md, "- Update dependencies") {
		t.Error("missing task entry")
	}
	// Features should appear before Fixes.
	featIdx := strings.Index(md, "Features")
	fixIdx := strings.Index(md, "Fixes")
	tweakIdx := strings.Index(md, "Tweaks")
	if featIdx > fixIdx || fixIdx > tweakIdx {
		t.Errorf("wrong section order: features=%d fixes=%d tweaks=%d", featIdx, fixIdx, tweakIdx)
	}
}

func TestFormatMarkdown_GitHubLinks(t *testing.T) {
	r := &Result{
		GitHub: "https://github.com/owner/repo",
		Range: TimeRange{
			Since: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{
					ID: "with-gh", Title: "Linked issue", Type: "bug",
					Sync: map[string]map[string]any{
						"github": {"issue_number": "42"},
					},
				},
				{ID: "no-gh", Title: "Unlinked issue", Type: "bug"},
			},
		},
	}

	md := FormatMarkdown(r, MarkdownOptions{Mode: "since"}, "")

	if !strings.Contains(md, "([#42](https://github.com/owner/repo/issues/42))") {
		t.Error("missing GitHub issue link")
	}
	if strings.Contains(md, "Unlinked issue (") {
		t.Error("unlinked issue should not have a reference")
	}
}

func TestFormatMarkdown_ExcludesExisting(t *testing.T) {
	r := &Result{
		Range: TimeRange{
			Since: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{ID: "already-there", Title: "Old entry", Type: "feature"},
				{ID: "brand-new", Title: "New entry", Type: "feature"},
			},
		},
	}

	existing := "## Week of Feb 23\n\n- Old entry (already-there)\n"
	md := FormatMarkdown(r, MarkdownOptions{Mode: "weekly"}, existing)

	if strings.Contains(md, "Old entry") {
		t.Error("should have excluded already-present issue")
	}
	if !strings.Contains(md, "New entry") {
		t.Error("should have included new issue")
	}
}

func TestFormatMarkdown_EmptyWhenAllExcluded(t *testing.T) {
	r := &Result{
		Range: TimeRange{
			Since: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{ID: "abc-123", Title: "Done", Type: "task"},
			},
		},
	}

	md := FormatMarkdown(r, MarkdownOptions{Mode: "weekly"}, "some content abc-123 more")
	if md != "" {
		t.Errorf("expected empty string, got %q", md)
	}
}

func TestFormatMarkdown_OmitsEmptySections(t *testing.T) {
	r := &Result{
		Range: TimeRange{
			Since: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{ID: "bug-only", Title: "Fix something", Type: "bug"},
			},
		},
	}

	md := FormatMarkdown(r, MarkdownOptions{Mode: "weekly"}, "")

	if strings.Contains(md, "Features") {
		t.Error("should omit empty Features section")
	}
	if strings.Contains(md, "Tweaks") {
		t.Error("should omit empty Tweaks section")
	}
	if !strings.Contains(md, "Fixes") {
		t.Error("should include non-empty Fixes section")
	}
}

func TestFormatMarkdown_SectionHeaders(t *testing.T) {
	r := &Result{
		Range: TimeRange{
			Since: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), // Sunday
			Until: time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{ID: "x", Title: "Something", Type: "task"},
			},
		},
	}

	tests := []struct {
		mode     string
		contains string
	}{
		{"weekly", "## Week of Mar 1 – Mar 7, 2026"},
		{"since", "## Since Mar 1, 2026"},
		{"daily", "## Mar 1, 2026"},
	}
	for _, tt := range tests {
		md := FormatMarkdown(r, MarkdownOptions{Mode: tt.mode}, "")
		if !strings.Contains(md, tt.contains) {
			t.Errorf("mode=%s: expected %q in output, got:\n%s", tt.mode, tt.contains, md)
		}
	}
}

func TestFormatMarkdown_AppendModeNoHeader(t *testing.T) {
	r := &Result{
		Range: TimeRange{
			Since: time.Date(2026, 3, 7, 10, 0, 0, 0, time.UTC),
			Until: time.Date(2026, 3, 7, 10, 0, 1, 0, time.UTC),
		},
		Issues: Issues{
			Completed: []*issue.Issue{
				{ID: "x", Title: "Quick fix", Type: "bug"},
			},
		},
	}

	md := FormatMarkdown(r, MarkdownOptions{Mode: "append"}, "")
	if strings.HasPrefix(md, "## ") {
		t.Error("append mode should not start with section header")
	}
	if !strings.Contains(md, "- Quick fix") {
		t.Error("should contain the entry")
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
