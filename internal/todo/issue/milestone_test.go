package issue

import (
	"strings"
	"testing"
)

func TestMilestoneRoundTrip(t *testing.T) {
	m := &Milestone{
		ID:          "ab1-xyz",
		Short:       "v1",
		Name:        "Version 1.0",
		Description: "First public release.",
	}
	due, err := ParseDueDate("2026-07-01")
	if err != nil {
		t.Fatalf("ParseDueDate: %v", err)
	}
	m.Due = due

	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	got, err := ParseMilestone(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("ParseMilestone: %v", err)
	}

	if got.Short != "v1" {
		t.Errorf("Short = %q, want v1", got.Short)
	}
	if got.Name != "Version 1.0" {
		t.Errorf("Name = %q, want Version 1.0", got.Name)
	}
	if got.Due == nil || got.Due.String() != "2026-07-01" {
		t.Errorf("Due = %v, want 2026-07-01", got.Due)
	}
	if got.Description != "First public release." {
		t.Errorf("Description = %q", got.Description)
	}
}

func TestMilestoneSyncRoundTrip(t *testing.T) {
	m := &Milestone{ID: "ab1-xyz", Short: "v1", Name: "V1"}
	m.SetSync("github", map[string]any{"milestone_number": "7"})

	out, err := m.Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	got, err := ParseMilestone(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("ParseMilestone: %v", err)
	}
	if !got.HasSync("github") {
		t.Fatalf("expected github sync data")
	}
	if got.Sync["github"]["milestone_number"] != "7" {
		t.Errorf("milestone_number = %v, want 7", got.Sync["github"]["milestone_number"])
	}
}

func TestBuildMilestonePath(t *testing.T) {
	got := BuildMilestonePath("ab1-xyz", "v1")
	want := "milestones/ab1-xyz--v1.md"
	if got != want {
		t.Errorf("BuildMilestonePath = %q, want %q", got, want)
	}
	// No slug.
	if got := BuildMilestonePath("ab1-xyz", ""); got != "milestones/ab1-xyz.md" {
		t.Errorf("BuildMilestonePath (no slug) = %q", got)
	}
}

func TestValidateShort(t *testing.T) {
	valid := []string{"v1", "v2", "1", "abc", "m1"}
	for _, s := range valid {
		if err := ValidateShort(s); err != nil {
			t.Errorf("ValidateShort(%q) unexpected error: %v", s, err)
		}
	}
	invalid := []string{"", "  ", "v1.0", "abcd", "a b"}
	for _, s := range invalid {
		if err := ValidateShort(s); err == nil {
			t.Errorf("ValidateShort(%q) expected error", s)
		}
	}
}

func TestIssueMilestoneFieldRoundTrip(t *testing.T) {
	in := `---
title: Has milestone
status: ready
milestone: ab1-xyz
---

Body.`
	b, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if b.Milestone != "ab1-xyz" {
		t.Fatalf("Milestone = %q, want ab1-xyz", b.Milestone)
	}

	out, err := b.Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(string(out), "milestone: ab1-xyz") {
		t.Errorf("rendered output missing milestone field:\n%s", out)
	}

	// Empty milestone must be omitted from output.
	b2 := &Issue{Title: "No milestone", Status: "ready"}
	out2, err := b2.Render()
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(string(out2), "milestone:") {
		t.Errorf("empty milestone should be omitted:\n%s", out2)
	}
}
