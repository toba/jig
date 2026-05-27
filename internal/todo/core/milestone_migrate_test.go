package core

import (
	"testing"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/issue"
)

func TestMigrateMilestoneTypeIssues(t *testing.T) {
	core, _ := setupTestCore(t)

	// Legacy milestone-type issue with sync metadata + two children.
	ms := &issue.Issue{
		ID:     "mod-001",
		Slug:   "modernization",
		Title:  "Modernization",
		Status: "completed",
		Type:   config.TypeMilestone,
		Body:   "Modernize everything.",
		Sync:   map[string]map[string]any{"github": {"milestone_number": "7"}},
	}
	if err := core.Create(ms); err != nil {
		t.Fatalf("Create milestone issue: %v", err)
	}
	child1 := &issue.Issue{ID: "chi-001", Title: "Child One", Status: "completed", Type: "task", Parent: "mod-001"}
	child2 := &issue.Issue{ID: "chi-002", Title: "Child Two", Status: "completed", Type: "task", Parent: "mod-001"}
	if err := core.Create(child1); err != nil {
		t.Fatal(err)
	}
	if err := core.Create(child2); err != nil {
		t.Fatal(err)
	}

	// Dry run should report but not change anything.
	dry, err := core.MigrateMilestoneTypeIssues(true)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if len(dry) != 1 {
		t.Fatalf("dry run migrations = %d, want 1", len(dry))
	}
	if len(core.AllMilestones()) != 0 {
		t.Fatalf("dry run created milestones: %d", len(core.AllMilestones()))
	}
	if _, err := core.Get("mod-001"); err != nil {
		t.Fatalf("dry run deleted the issue")
	}

	// Real migration.
	migs, err := core.MigrateMilestoneTypeIssues(false)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if len(migs) != 1 {
		t.Fatalf("migrations = %d, want 1", len(migs))
	}
	newID := migs[0].NewMilestoneID
	if newID == "" {
		t.Fatal("expected new milestone ID")
	}

	// Milestone entity created, carrying sync number.
	m, err := core.GetMilestone(newID)
	if err != nil {
		t.Fatalf("GetMilestone: %v", err)
	}
	if m.Name != "Modernization" {
		t.Errorf("Name = %q", m.Name)
	}
	if m.Sync["github"]["milestone_number"] != "7" {
		t.Errorf("milestone_number not carried over: %v", m.Sync)
	}

	// Old issue deleted.
	if _, err := core.Get("mod-001"); err == nil {
		t.Error("old milestone-type issue should be deleted")
	}

	// Children reassigned: milestone set, parent cleared.
	for _, id := range []string{"chi-001", "chi-002"} {
		ch, err := core.Get(id)
		if err != nil {
			t.Fatalf("Get %s: %v", id, err)
		}
		if ch.Milestone != newID {
			t.Errorf("%s milestone = %q, want %q", id, ch.Milestone, newID)
		}
		if ch.Parent != "" {
			t.Errorf("%s parent = %q, want empty", id, ch.Parent)
		}
	}

	// Idempotent: second run is a no-op.
	again, err := core.MigrateMilestoneTypeIssues(false)
	if err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("second migration should be no-op, got %d", len(again))
	}
}

func TestDeriveShortCollision(t *testing.T) {
	used := map[string]bool{"mod": true}
	got := deriveShort("Modernization", used)
	if got == "mod" {
		t.Errorf("expected collision avoidance, got %q", got)
	}
	if len(got) > 3 {
		t.Errorf("short %q exceeds 3 chars", got)
	}
}
