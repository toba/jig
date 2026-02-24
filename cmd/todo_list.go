package cmd

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/graph"
	"github.com/toba/jig/internal/todo/graph/model"
	"github.com/toba/jig/internal/todo/issue"
	"github.com/toba/jig/internal/todo/output"
	"github.com/toba/jig/internal/todo/ui"
	"golang.org/x/term"
)

var (
	listJSON        bool
	listSearch      string
	listStatus      []string
	listNoStatus    []string
	listType        []string
	listNoType      []string
	listPriority    []string
	listNoPriority  []string
	listTag         []string
	listNoTag       []string
	listHasParent   bool
	listNoParent    bool
	listParentID    string
	listHasBlocking bool
	listNoBlocking  bool
	listIsBlocked   bool
	listReady       bool
	listQuiet       bool
	listSort        string
	listFull        bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all issues",
	Long: `Lists all issues in the data directory.

Search Syntax (--search/-S):
  The search flag supports Bleve query string syntax:

  login          Exact term match
  login~         Fuzzy match (1 edit distance)
  log*           Wildcard prefix match
  "user login"   Exact phrase match
  user AND login Both terms required
  user OR login  Either term matches`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := &model.IssueFilter{
			Status:          listStatus,
			ExcludeStatus:   listNoStatus,
			Type:            listType,
			ExcludeType:     listNoType,
			Priority:        listPriority,
			ExcludePriority: listNoPriority,
			Tags:            listTag,
			ExcludeTags:     listNoTag,
		}

		if listSearch != "" {
			filter.Search = &listSearch
		}
		if listHasParent {
			filter.HasParent = &listHasParent
		}
		if listNoParent {
			filter.NoParent = &listNoParent
		}
		if listParentID != "" {
			filter.ParentID = &listParentID
		}
		if listHasBlocking {
			filter.HasBlocking = &listHasBlocking
		}
		if listNoBlocking {
			filter.NoBlocking = &listNoBlocking
		}
		if listReady && listIsBlocked {
			return errors.New("--ready and --is-blocked are mutually exclusive")
		}
		if listIsBlocked {
			filter.IsBlocked = &listIsBlocked
		}
		if listReady {
			isBlocked := false
			filter.IsBlocked = &isBlocked
			filter.ExcludeStatus = append(filter.ExcludeStatus, todoconfig.StatusInProgress, todoconfig.StatusReview, todoconfig.StatusCompleted, todoconfig.StatusScrapped, todoconfig.StatusDraft)
		}

		resolver := &graph.Resolver{Core: todoStore}
		issues, err := resolver.Query().Issues(context.Background(), filter)
		if err != nil {
			return fmt.Errorf("querying issues: %w", err)
		}

		sortIssues(issues, listSort, todoCfg)

		if listJSON {
			if !listFull {
				for _, b := range issues {
					b.Body = ""
				}
			}
			return output.SuccessMultiple(issues)
		}

		if listQuiet {
			for _, b := range issues {
				fmt.Println(b.ID)
			}
			return nil
		}

		// Tree view
		allIssues, err := resolver.Query().Issues(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("querying all issues for tree: %w", err)
		}

		sortFn := func(b []*issue.Issue) {
			sortIssues(b, listSort, todoCfg)
		}

		tree := ui.BuildTree(issues, allIssues, sortFn)

		if len(tree) == 0 {
			fmt.Println(ui.Muted.Render("No issues found. Create one with: jig todo create <title>"))
			return nil
		}

		maxIDWidth := 2
		for _, b := range allIssues {
			if len(b.ID) > maxIDWidth {
				maxIDWidth = len(b.ID)
			}
		}
		maxIDWidth += 2

		hasTags := false
		for _, b := range issues {
			if len(b.Tags) > 0 {
				hasTags = true
				break
			}
		}

		termWidth := 80
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
			termWidth = w
		}

		fmt.Print(ui.RenderTree(tree, todoCfg, maxIDWidth, hasTags, termWidth))
		return nil
	},
}

func sortIssues(issues []*issue.Issue, sortBy string, cfg *todoconfig.Config) {
	statusNames := cfg.StatusNames()
	priorityNames := cfg.PriorityNames()
	typeNames := cfg.TypeNames()

	switch sortBy {
	case "created":
		slices.SortFunc(issues, func(a, b *issue.Issue) int {
			if a.CreatedAt == nil && b.CreatedAt == nil {
				return cmp.Compare(a.ID, b.ID)
			}
			if a.CreatedAt == nil {
				return 1
			}
			if b.CreatedAt == nil {
				return -1
			}
			return b.CreatedAt.Compare(*a.CreatedAt) // newest first
		})
	case "updated":
		slices.SortFunc(issues, func(a, b *issue.Issue) int {
			if a.UpdatedAt == nil && b.UpdatedAt == nil {
				return cmp.Compare(a.ID, b.ID)
			}
			if a.UpdatedAt == nil {
				return 1
			}
			if b.UpdatedAt == nil {
				return -1
			}
			return b.UpdatedAt.Compare(*a.UpdatedAt) // newest first
		})
	case "status":
		statusOrder := make(map[string]int)
		for i, s := range statusNames {
			statusOrder[s] = i
		}
		slices.SortFunc(issues, func(a, b *issue.Issue) int {
			if c := cmp.Compare(statusOrder[a.Status], statusOrder[b.Status]); c != 0 {
				return c
			}
			return cmp.Compare(a.ID, b.ID)
		})
	case "priority":
		priorityOrder := make(map[string]int)
		for i, p := range priorityNames {
			priorityOrder[p] = i
		}
		normalIdx := len(priorityNames)
		for i, p := range priorityNames {
			if p == todoconfig.PriorityNormal {
				normalIdx = i
				break
			}
		}
		getPriority := func(b *issue.Issue) int {
			if b.Priority != "" {
				if order, ok := priorityOrder[b.Priority]; ok {
					return order
				}
			}
			return normalIdx
		}
		slices.SortFunc(issues, func(a, b *issue.Issue) int {
			if c := cmp.Compare(getPriority(a), getPriority(b)); c != 0 {
				return c
			}
			return cmp.Compare(a.ID, b.ID)
		})
	case "due":
		issue.SortByDueDate(issues)
	case "id":
		slices.SortFunc(issues, func(a, b *issue.Issue) int {
			return cmp.Compare(a.ID, b.ID)
		})
	default:
		issue.SortByStatusPriorityAndType(issues, statusNames, priorityNames, typeNames)
	}
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	listCmd.Flags().StringVarP(&listSearch, "search", "S", "", "Full-text search in title and body")
	listCmd.Flags().StringArrayVarP(&listStatus, "status", "s", nil, "Filter by status (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoStatus, "no-status", nil, "Exclude by status (can be repeated)")
	listCmd.Flags().StringArrayVarP(&listType, "type", "t", nil, "Filter by type (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoType, "no-type", nil, "Exclude by type (can be repeated)")
	listCmd.Flags().StringArrayVarP(&listPriority, "priority", "p", nil, "Filter by priority (can be repeated)")
	listCmd.Flags().StringArrayVar(&listNoPriority, "no-priority", nil, "Exclude by priority (can be repeated)")
	listCmd.Flags().StringArrayVar(&listTag, "tag", nil, "Filter by tag (can be repeated, OR logic)")
	listCmd.Flags().StringArrayVar(&listNoTag, "no-tag", nil, "Exclude issues with tag (can be repeated)")
	listCmd.Flags().BoolVar(&listHasParent, "has-parent", false, "Filter issues with a parent")
	listCmd.Flags().BoolVar(&listNoParent, "no-parent", false, "Filter issues without a parent")
	listCmd.Flags().StringVar(&listParentID, "parent", "", "Filter by parent ID")
	listCmd.Flags().BoolVar(&listHasBlocking, "has-blocking", false, "Filter issues that are blocking others")
	listCmd.Flags().BoolVar(&listNoBlocking, "no-blocking", false, "Filter issues that aren't blocking others")
	listCmd.Flags().BoolVar(&listIsBlocked, "is-blocked", false, "Filter issues that are blocked by others")
	listCmd.Flags().BoolVar(&listReady, "ready", false, "Filter issues available to start")
	listCmd.Flags().BoolVarP(&listQuiet, "quiet", "q", false, "Only output IDs (one per line)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "Sort by: created, updated, due, status, priority, id")
	listCmd.Flags().BoolVar(&listFull, "full", false, "Include issue body in JSON output")
	todoCmd.AddCommand(listCmd)
}
