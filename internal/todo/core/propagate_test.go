package core

import (
	"testing"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/issue"
)

func TestPropagateInProgressBubblesUp(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusReady)
	child := createTestIssue(t, c, "c1", "Child", config.StatusReady)
	child.Parent = parent.ID
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	// Move child to in-progress
	child.Status = config.StatusInProgress
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusInProgress {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusInProgress)
	}
}

func TestPropagateInProgressFromDraft(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusDraft)
	child := createTestIssue(t, c, "c1", "Child", config.StatusDraft)
	child.Parent = parent.ID
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	child.Status = config.StatusInProgress
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusInProgress {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusInProgress)
	}
}

func TestPropagateInProgressSkipsIfParentAlreadyInProgress(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusInProgress)
	child := createTestIssue(t, c, "c1", "Child", config.StatusReady)
	child.Parent = parent.ID
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	child.Status = config.StatusInProgress
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusInProgress {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusInProgress)
	}
}

func TestPropagateInProgressSkipsIfParentInReview(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusReview)
	child := createTestIssue(t, c, "c1", "Child", config.StatusReady)
	child.Parent = parent.ID
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	child.Status = config.StatusInProgress
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusReview {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusReview)
	}
}

func TestPropagateReviewBubblesUp(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusInProgress)
	child1 := createTestIssue(t, c, "c1", "Child1", config.StatusCompleted)
	child1.Parent = parent.ID
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}
	child2 := createTestIssue(t, c, "c2", "Child2", config.StatusInProgress)
	child2.Parent = parent.ID
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	child2.Status = config.StatusReview
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusReview {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusReview)
	}
}

func TestPropagateCompletedBubblesUp(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusInProgress)
	child1 := createTestIssue(t, c, "c1", "Child1", config.StatusCompleted)
	child1.Parent = parent.ID
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}
	child2 := createTestIssue(t, c, "c2", "Child2", config.StatusInProgress)
	child2.Parent = parent.ID
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	child2.Status = config.StatusCompleted
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusCompleted {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusCompleted)
	}
}

func TestPropagateCompletedWithScrappedMix(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusInProgress)
	child1 := createTestIssue(t, c, "c1", "Child1", config.StatusCompleted)
	child1.Parent = parent.ID
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}
	child2 := createTestIssue(t, c, "c2", "Child2", config.StatusScrapped)
	child2.Parent = parent.ID
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusCompleted {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusCompleted)
	}
}

func TestPropagateAllScrapped(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusInProgress)
	child1 := createTestIssue(t, c, "c1", "Child1", config.StatusReady)
	child1.Parent = parent.ID
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}
	child2 := createTestIssue(t, c, "c2", "Child2", config.StatusReady)
	child2.Parent = parent.ID
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	child1.Status = config.StatusScrapped
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}
	child2.Status = config.StatusScrapped
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusScrapped {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusScrapped)
	}
}

func TestPropagateNoParent(t *testing.T) {
	c, _ := setupTestCore(t)
	child := createTestIssue(t, c, "c1", "Child", config.StatusReady)

	child.Status = config.StatusInProgress
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}
	// No crash = pass
}

func TestPropagateNoChildren(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusReady)
	child := createTestIssue(t, c, "c1", "Child", config.StatusReady)
	child.Parent = parent.ID
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	// Update a different parentless issue — parent should be unchanged
	other := createTestIssue(t, c, "o1", "Other", config.StatusReady)
	other.Status = config.StatusInProgress
	if err := c.Update(other, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusReady {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusReady)
	}
}

func TestPropagatePartialNoRuleMatches(t *testing.T) {
	c, _ := setupTestCore(t)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusInProgress)
	// Create both children as ready first, set parents
	child1 := createTestIssue(t, c, "c1", "Child1", config.StatusReady)
	child1.Parent = parent.ID
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}
	child2 := createTestIssue(t, c, "c2", "Child2", config.StatusReady)
	child2.Parent = parent.ID
	if err := c.Update(child2, nil); err != nil {
		t.Fatal(err)
	}

	// Now complete only child1 — child2 is still ready, so no terminal rule matches
	child1.Status = config.StatusCompleted
	if err := c.Update(child1, nil); err != nil {
		t.Fatal(err)
	}

	got, _ := c.Get("p1")
	if got.Status != config.StatusInProgress {
		t.Errorf("parent status = %q, want %q", got.Status, config.StatusInProgress)
	}
}

func TestPropagateRecursiveThreeLevels(t *testing.T) {
	c, _ := setupTestCore(t)
	gp := createTestIssue(t, c, "gp", "Grandparent", config.StatusReady)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusReady)
	parent.Parent = gp.ID
	if err := c.Update(parent, nil); err != nil {
		t.Fatal(err)
	}
	child := createTestIssue(t, c, "c1", "Child", config.StatusReady)
	child.Parent = parent.ID
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	child.Status = config.StatusInProgress
	if err := c.Update(child, nil); err != nil {
		t.Fatal(err)
	}

	gotParent, _ := c.Get("p1")
	if gotParent.Status != config.StatusInProgress {
		t.Errorf("parent status = %q, want %q", gotParent.Status, config.StatusInProgress)
	}
	gotGP, _ := c.Get("gp")
	if gotGP.Status != config.StatusInProgress {
		t.Errorf("grandparent status = %q, want %q", gotGP.Status, config.StatusInProgress)
	}
}

func TestPropagateRecursiveCompleted(t *testing.T) {
	c, _ := setupTestCore(t)
	gp := createTestIssue(t, c, "gp", "Grandparent", config.StatusReady)
	parent := createTestIssue(t, c, "p1", "Parent", config.StatusReady)
	parent.Parent = gp.ID
	if err := c.Update(parent, nil); err != nil {
		t.Fatal(err)
	}
	c1 := createTestIssue(t, c, "c1", "Child1", config.StatusReady)
	c1.Parent = parent.ID
	if err := c.Update(c1, nil); err != nil {
		t.Fatal(err)
	}
	c2 := createTestIssue(t, c, "c2", "Child2", config.StatusReady)
	c2.Parent = parent.ID
	if err := c.Update(c2, nil); err != nil {
		t.Fatal(err)
	}

	// Complete both children
	c1.Status = config.StatusCompleted
	if err := c.Update(c1, nil); err != nil {
		t.Fatal(err)
	}
	c2.Status = config.StatusCompleted
	if err := c.Update(c2, nil); err != nil {
		t.Fatal(err)
	}

	gotParent, _ := c.Get("p1")
	if gotParent.Status != config.StatusCompleted {
		t.Errorf("parent status = %q, want %q", gotParent.Status, config.StatusCompleted)
	}
	gotGP, _ := c.Get("gp")
	if gotGP.Status != config.StatusCompleted {
		t.Errorf("grandparent status = %q, want %q", gotGP.Status, config.StatusCompleted)
	}
}

func TestComputeParentStatusUnit(t *testing.T) {
	parent := &issue.Issue{ID: "p1", Status: config.StatusInProgress}

	tests := []struct {
		name     string
		parent   *issue.Issue
		children []*issue.Issue
		want     string
	}{
		{
			name:     "no children",
			parent:   parent,
			children: nil,
			want:     "",
		},
		{
			name:   "all completed",
			parent: &issue.Issue{ID: "p", Status: config.StatusInProgress},
			children: []*issue.Issue{
				{Status: config.StatusCompleted},
				{Status: config.StatusCompleted},
			},
			want: config.StatusCompleted,
		},
		{
			name:   "all scrapped",
			parent: &issue.Issue{ID: "p", Status: config.StatusInProgress},
			children: []*issue.Issue{
				{Status: config.StatusScrapped},
				{Status: config.StatusScrapped},
			},
			want: config.StatusScrapped,
		},
		{
			name:   "completed + scrapped = completed",
			parent: &issue.Issue{ID: "p", Status: config.StatusInProgress},
			children: []*issue.Issue{
				{Status: config.StatusCompleted},
				{Status: config.StatusScrapped},
			},
			want: config.StatusCompleted,
		},
		{
			name:   "review + completed = review",
			parent: &issue.Issue{ID: "p", Status: config.StatusInProgress},
			children: []*issue.Issue{
				{Status: config.StatusReview},
				{Status: config.StatusCompleted},
			},
			want: config.StatusReview,
		},
		{
			name:   "in-progress with ready parent",
			parent: &issue.Issue{ID: "p", Status: config.StatusReady},
			children: []*issue.Issue{
				{Status: config.StatusInProgress},
				{Status: config.StatusReady},
			},
			want: config.StatusInProgress,
		},
		{
			name:   "in-progress with in-progress parent = no change",
			parent: &issue.Issue{ID: "p", Status: config.StatusInProgress},
			children: []*issue.Issue{
				{Status: config.StatusInProgress},
			},
			want: "",
		},
		{
			name:   "completed + ready = no change",
			parent: &issue.Issue{ID: "p", Status: config.StatusInProgress},
			children: []*issue.Issue{
				{Status: config.StatusCompleted},
				{Status: config.StatusReady},
			},
			want: "",
		},
		{
			name:   "already completed = no change",
			parent: &issue.Issue{ID: "p", Status: config.StatusCompleted},
			children: []*issue.Issue{
				{Status: config.StatusCompleted},
			},
			want: "",
		},
		{
			name:   "all only scrapped no completed",
			parent: &issue.Issue{ID: "p", Status: config.StatusReady},
			children: []*issue.Issue{
				{Status: config.StatusScrapped},
			},
			want: config.StatusScrapped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeParentStatus(tt.parent, tt.children)
			if got != tt.want {
				t.Errorf("computeParentStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
