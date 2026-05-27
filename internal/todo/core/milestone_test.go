package core

import (
	"testing"

	"github.com/toba/jig/internal/todo/issue"
)

func TestMilestoneCRUD(t *testing.T) {
	core, _ := setupTestCore(t)

	m := &issue.Milestone{Short: "v1", Name: "Version 1.0", Description: "First release"}
	if err := core.CreateMilestone(m); err != nil {
		t.Fatalf("CreateMilestone: %v", err)
	}
	if m.ID == "" {
		t.Fatal("expected generated ID")
	}

	got, err := core.GetMilestone(m.ID)
	if err != nil {
		t.Fatalf("GetMilestone: %v", err)
	}
	if got.Name != "Version 1.0" {
		t.Errorf("Name = %q", got.Name)
	}

	got.Short = "v2"
	if err := core.UpdateMilestone(got); err != nil {
		t.Fatalf("UpdateMilestone: %v", err)
	}

	all := core.AllMilestones()
	if len(all) != 1 {
		t.Fatalf("AllMilestones len = %d, want 1", len(all))
	}

	if !core.MilestoneExists(m.ID) {
		t.Error("MilestoneExists should be true")
	}

	if err := core.DeleteMilestone(m.ID); err != nil {
		t.Fatalf("DeleteMilestone: %v", err)
	}
	if core.MilestoneExists(m.ID) {
		t.Error("MilestoneExists should be false after delete")
	}
}

// TestMilestoneNotLoadedAsIssue verifies milestone files in .issues/milestones/
// never appear in the issue set.
func TestMilestoneNotLoadedAsIssue(t *testing.T) {
	core, _ := setupTestCore(t)

	m := &issue.Milestone{Short: "v1", Name: "V1"}
	if err := core.CreateMilestone(m); err != nil {
		t.Fatalf("CreateMilestone: %v", err)
	}
	createTestIssue(t, core, "aaa-bbb", "A real issue", "ready")

	// Reload from disk and confirm the milestone is not in the issue set.
	if err := core.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, b := range core.All() {
		if b.ID == m.ID {
			t.Fatalf("milestone %s leaked into issue set", m.ID)
		}
	}
	if len(core.All()) != 1 {
		t.Errorf("issue count = %d, want 1", len(core.All()))
	}
	if len(core.AllMilestones()) != 1 {
		t.Errorf("milestone count = %d, want 1", len(core.AllMilestones()))
	}
}

func TestMilestonePersistsAcrossReload(t *testing.T) {
	core, dataDir := setupTestCore(t)
	m := &issue.Milestone{Short: "v1", Name: "Persisted"}
	if err := core.CreateMilestone(m); err != nil {
		t.Fatalf("CreateMilestone: %v", err)
	}

	core2 := New(dataDir, core.Config())
	core2.SetWarnWriter(nil)
	if err := core2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	got, err := core2.GetMilestone(m.ID)
	if err != nil {
		t.Fatalf("GetMilestone after reload: %v", err)
	}
	if got.Name != "Persisted" {
		t.Errorf("Name = %q after reload", got.Name)
	}
}
