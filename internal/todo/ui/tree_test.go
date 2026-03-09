package ui

import (
	"slices"
	"testing"

	"github.com/toba/jig/internal/todo/issue"
)

func TestBuildTree(t *testing.T) {
	// Create test issues with parent relationships:
	// milestone1
	//   └── epic1
	//       └── task1
	// task2 (orphan)

	milestone1 := &issue.Issue{ID: "m1", Title: "Milestone 1", Type: "milestone"}
	epic1 := &issue.Issue{ID: "e1", Title: "Epic 1", Type: "epic", Parent: "m1"}
	task1 := &issue.Issue{ID: "t1", Title: "Task 1", Type: "task", Parent: "e1"}
	task2 := &issue.Issue{ID: "t2", Title: "Task 2", Type: "task"} // orphan

	allIssues := []*issue.Issue{milestone1, epic1, task1, task2}

	// Identity sort function (no sorting)
	noSort := func(b []*issue.Issue) {}

	t.Run("all issues matched", func(t *testing.T) {
		tree := BuildTree(allIssues, allIssues, noSort)

		// Should have 2 root nodes: milestone1 and task2
		if len(tree) != 2 {
			t.Errorf("expected 2 root nodes, got %d", len(tree))
		}

		// Find milestone node
		var milestoneNode *TreeNode
		for _, n := range tree {
			if n.Issue.ID == "m1" {
				milestoneNode = n
				break
			}
		}
		if milestoneNode == nil {
			t.Fatal("milestone node not found")
		}
		if !milestoneNode.Matched {
			t.Error("milestone should be marked as matched")
		}

		// Milestone should have epic as child
		if len(milestoneNode.Children) != 1 {
			t.Errorf("milestone should have 1 child, got %d", len(milestoneNode.Children))
		}
		epicNode := milestoneNode.Children[0]
		if epicNode.Issue.ID != "e1" {
			t.Errorf("expected epic child, got %s", epicNode.Issue.ID)
		}

		// Epic should have task as child
		if len(epicNode.Children) != 1 {
			t.Errorf("epic should have 1 child, got %d", len(epicNode.Children))
		}
		taskNode := epicNode.Children[0]
		if taskNode.Issue.ID != "t1" {
			t.Errorf("expected task child, got %s", taskNode.Issue.ID)
		}
	})

	t.Run("filter leaf only - ancestors included", func(t *testing.T) {
		// Only task1 matched, but ancestors should be included
		matchedIssues := []*issue.Issue{task1}
		tree := BuildTree(matchedIssues, allIssues, noSort)

		// Should have 1 root: milestone (as ancestor)
		if len(tree) != 1 {
			t.Errorf("expected 1 root node, got %d", len(tree))
		}

		milestoneNode := tree[0]
		if milestoneNode.Issue.ID != "m1" {
			t.Errorf("expected milestone as root, got %s", milestoneNode.Issue.ID)
		}
		if milestoneNode.Matched {
			t.Error("milestone should NOT be marked as matched (it's an ancestor)")
		}

		// Should have epic as child (also ancestor)
		if len(milestoneNode.Children) != 1 {
			t.Fatalf("milestone should have 1 child, got %d", len(milestoneNode.Children))
		}
		epicNode := milestoneNode.Children[0]
		if epicNode.Matched {
			t.Error("epic should NOT be marked as matched (it's an ancestor)")
		}

		// Task should be matched
		if len(epicNode.Children) != 1 {
			t.Fatalf("epic should have 1 child, got %d", len(epicNode.Children))
		}
		taskNode := epicNode.Children[0]
		if !taskNode.Matched {
			t.Error("task should be marked as matched")
		}
	})

	t.Run("filter middle - ancestors included", func(t *testing.T) {
		// Only epic1 matched
		matchedIssues := []*issue.Issue{epic1}
		tree := BuildTree(matchedIssues, allIssues, noSort)

		// Should have 1 root: milestone (ancestor)
		if len(tree) != 1 {
			t.Errorf("expected 1 root node, got %d", len(tree))
		}

		milestoneNode := tree[0]
		if milestoneNode.Matched {
			t.Error("milestone should NOT be marked as matched")
		}

		epicNode := milestoneNode.Children[0]
		if !epicNode.Matched {
			t.Error("epic should be marked as matched")
		}

		// Epic should have no children (task1 was not matched)
		if len(epicNode.Children) != 0 {
			t.Errorf("epic should have 0 children (task not matched), got %d", len(epicNode.Children))
		}
	})

	t.Run("orphan issue", func(t *testing.T) {
		matchedIssues := []*issue.Issue{task2}
		tree := BuildTree(matchedIssues, allIssues, noSort)

		if len(tree) != 1 {
			t.Errorf("expected 1 root node, got %d", len(tree))
		}
		if tree[0].Issue.ID != "t2" {
			t.Errorf("expected task2 as root, got %s", tree[0].Issue.ID)
		}
		if !tree[0].Matched {
			t.Error("task2 should be marked as matched")
		}
	})

	t.Run("broken parent link", func(t *testing.T) {
		// Issue with parent that doesn't exist
		brokenIssue := &issue.Issue{ID: "broken", Title: "Broken", Parent: "nonexistent"}
		matchedIssues := []*issue.Issue{brokenIssue}
		allIssuesWithBroken := append(slices.Clone(allIssues), brokenIssue)

		tree := BuildTree(matchedIssues, allIssuesWithBroken, noSort)

		// Should be treated as root (parent not found)
		if len(tree) != 1 {
			t.Errorf("expected 1 root node, got %d", len(tree))
		}
		if tree[0].Issue.ID != "broken" {
			t.Errorf("expected broken issue as root, got %s", tree[0].Issue.ID)
		}
	})
}

func TestLeafCounts(t *testing.T) {
	// Tree structure:
	// milestone1
	//   ├── epic1
	//   │   ├── task1 (leaf)
	//   │   └── task2 (leaf)
	//   └── task3 (leaf)
	// task4 (leaf, orphan root)

	task1 := &TreeNode{Issue: &issue.Issue{ID: "t1"}}
	task2 := &TreeNode{Issue: &issue.Issue{ID: "t2"}}
	task3 := &TreeNode{Issue: &issue.Issue{ID: "t3"}}
	epic1 := &TreeNode{Issue: &issue.Issue{ID: "e1"}, Children: []*TreeNode{task1, task2}}
	milestone1 := &TreeNode{Issue: &issue.Issue{ID: "m1"}, Children: []*TreeNode{epic1, task3}}
	task4 := &TreeNode{Issue: &issue.Issue{ID: "t4"}}

	nodes := []*TreeNode{milestone1, task4}
	counts := LeafCounts(nodes)

	t.Run("root with children has correct leaf count", func(t *testing.T) {
		if counts["m1"] != 3 {
			t.Errorf("expected 3 leaves for m1, got %d", counts["m1"])
		}
	})

	t.Run("leaf root not in counts", func(t *testing.T) {
		if _, ok := counts["t4"]; ok {
			t.Error("leaf root t4 should not be in counts")
		}
	})

	t.Run("intermediate node not in counts (only top-level)", func(t *testing.T) {
		// LeafCounts only considers top-level nodes passed to it
		if _, ok := counts["e1"]; ok {
			t.Error("epic e1 should not be in counts (not a top-level node)")
		}
	})
}

func TestFlattenTreeRootID(t *testing.T) {
	// Tree:
	// root1
	//   └── child1
	//       └── grandchild1
	// root2 (leaf)

	grandchild1 := &TreeNode{Issue: &issue.Issue{ID: "gc1"}, Matched: true}
	child1 := &TreeNode{Issue: &issue.Issue{ID: "c1"}, Matched: true, Children: []*TreeNode{grandchild1}}
	root1 := &TreeNode{Issue: &issue.Issue{ID: "r1"}, Matched: true, Children: []*TreeNode{child1}}
	root2 := &TreeNode{Issue: &issue.Issue{ID: "r2"}, Matched: true}

	items := FlattenTree([]*TreeNode{root1, root2})

	if len(items) != 4 {
		t.Fatalf("expected 4 flat items, got %d", len(items))
	}

	// root1 (depth 0) → RootID = "r1"
	if items[0].RootID != "r1" {
		t.Errorf("root1: expected RootID 'r1', got %q", items[0].RootID)
	}
	// child1 (depth 1) → RootID = "r1"
	if items[1].RootID != "r1" {
		t.Errorf("child1: expected RootID 'r1', got %q", items[1].RootID)
	}
	// grandchild1 (depth 2) → RootID = "r1"
	if items[2].RootID != "r1" {
		t.Errorf("grandchild1: expected RootID 'r1', got %q", items[2].RootID)
	}
	// root2 (depth 0) → RootID = "r2"
	if items[3].RootID != "r2" {
		t.Errorf("root2: expected RootID 'r2', got %q", items[3].RootID)
	}
}

func TestTreeNodeToJSON(t *testing.T) {
	b := &issue.Issue{
		ID:       "test-id",
		Slug:     "test-slug",
		Path:     "test.md",
		Title:    "Test Title",
		Status:   "todo",
		Type:     "task",
		Priority: "high",
		Tags:     []string{"tag1", "tag2"},
		Body:     "Test body content",
	}

	node := &TreeNode{
		Issue:   b,
		Matched: true,
		Children: []*TreeNode{
			{
				Issue:   &issue.Issue{ID: "child-id", Title: "Child"},
				Matched: false,
			},
		},
	}

	t.Run("without full body", func(t *testing.T) {
		json := node.ToJSON(false)
		if json.ID != "test-id" {
			t.Errorf("expected id 'test-id', got %s", json.ID)
		}
		if json.Body != "" {
			t.Error("body should be empty when includeFull is false")
		}
		if !json.Matched {
			t.Error("matched should be true")
		}
		if len(json.Children) != 1 {
			t.Errorf("expected 1 child, got %d", len(json.Children))
		}
	})

	t.Run("with full body", func(t *testing.T) {
		json := node.ToJSON(true)
		if json.Body != "Test body content" {
			t.Errorf("expected body content, got %s", json.Body)
		}
	})
}
