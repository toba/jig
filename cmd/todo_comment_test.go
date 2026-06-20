package cmd

import (
	"strings"
	"testing"

	"github.com/toba/jig/internal/todo/issue"
)

func TestCommentIssueAppends(t *testing.T) {
	testCore, cleanup := setupQueryTestCore(t)
	defer cleanup()

	if err := testCore.Create(&issue.Issue{
		ID:     "cmt-1",
		Slug:   issue.Slugify("Has a body"),
		Title:  "Has a body",
		Status: "in-progress",
		Body:   "Original body.",
	}); err != nil {
		t.Fatalf("seeding issue: %v", err)
	}

	b, err := commentIssue("cmt-1", "## Summary\n\nDid the thing.")
	if err != nil {
		t.Fatalf("commentIssue() error = %v", err)
	}

	if !strings.Contains(b.Body, "Original body.") {
		t.Errorf("comment clobbered existing body: %q", b.Body)
	}
	if !strings.Contains(b.Body, "## Summary") || !strings.Contains(b.Body, "Did the thing.") {
		t.Errorf("comment did not append text: %q", b.Body)
	}
	// Appended content should be separated from the original by a blank line.
	if !strings.Contains(b.Body, "Original body.\n\n## Summary") {
		t.Errorf("comment not separated from existing body by blank line: %q", b.Body)
	}
}

func TestCommentIssueEmptyTextErrors(t *testing.T) {
	testCore, cleanup := setupQueryTestCore(t)
	defer cleanup()

	createQueryTestIssue(t, testCore, "cmt-2", "Empty target", "in-progress")

	if _, err := commentIssue("cmt-2", "   "); err == nil {
		t.Error("expected error for empty comment text, got nil")
	}
}

func TestCommentIssueNotFound(t *testing.T) {
	_, cleanup := setupQueryTestCore(t)
	defer cleanup()

	if _, err := commentIssue("does-not-exist", "hello"); err == nil {
		t.Error("expected error for missing issue, got nil")
	}
}
