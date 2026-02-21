package issue

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"github.com/toba/jig/internal/todo/config"
)

// CompareByStatusPriorityAndType returns true if a should sort before b,
// using status order, then priority, then type, then title.
// Unrecognized statuses, priorities, and types are sorted last within their category.
// Issues without priority are treated as "normal" priority for sorting purposes.
func CompareByStatusPriorityAndType(a, b *Issue, statusNames, priorityNames, typeNames []string) bool {
	statusOrder := make(map[string]int)
	for i, s := range statusNames {
		statusOrder[s] = i
	}
	priorityOrder := make(map[string]int)
	for i, p := range priorityNames {
		priorityOrder[p] = i
	}
	typeOrder := make(map[string]int)
	for i, t := range typeNames {
		typeOrder[t] = i
	}

	// Find the index of "normal" priority for issues without priority set
	normalPriorityOrder := len(priorityNames) // default to last if "normal" not found
	for i, p := range priorityNames {
		if p == config.PriorityNormal {
			normalPriorityOrder = i
			break
		}
	}

	// Helper to get order with unrecognized values sorted last
	getStatusOrder := func(status string) int {
		if order, ok := statusOrder[status]; ok {
			return order
		}
		return len(statusNames)
	}
	getPriorityOrder := func(priority string) int {
		if priority == "" {
			return normalPriorityOrder
		}
		if order, ok := priorityOrder[priority]; ok {
			return order
		}
		return len(priorityNames)
	}
	getTypeOrder := func(typ string) int {
		if order, ok := typeOrder[typ]; ok {
			return order
		}
		return len(typeNames)
	}

	// Primary: status order
	oi, oj := getStatusOrder(a.Status), getStatusOrder(b.Status)
	if oi != oj {
		return oi < oj
	}
	// Secondary: priority order
	pi, pj := getPriorityOrder(a.Priority), getPriorityOrder(b.Priority)
	if pi != pj {
		return pi < pj
	}
	// Tertiary: type order
	ti, tj := getTypeOrder(a.Type), getTypeOrder(b.Type)
	if ti != tj {
		return ti < tj
	}
	// Quaternary: title (case-insensitive) for stable, user-friendly ordering
	return strings.ToLower(a.Title) < strings.ToLower(b.Title)
}

// SortByStatusPriorityAndType sorts issues by status order, then priority, then type, then title.
// This is the default sorting used by both CLI and TUI.
func SortByStatusPriorityAndType(issues []*Issue, statusNames, priorityNames, typeNames []string) {
	slices.SortFunc(issues, func(a, b *Issue) int {
		if CompareByStatusPriorityAndType(a, b, statusNames, priorityNames, typeNames) {
			return -1
		}
		if CompareByStatusPriorityAndType(b, a, statusNames, priorityNames, typeNames) {
			return 1
		}
		return 0
	})
}

// ComputeEffectiveDates builds a map of issue ID to effective date for sorting.
// The effective date for an issue is the maximum of its own date and all descendants' dates.
// field must be "created_at" or "updated_at".
func ComputeEffectiveDates(allIssues []*Issue, field string) map[string]time.Time {
	// Build parent→children index
	children := map[string][]string{}
	issueByID := map[string]*Issue{}
	for _, b := range allIssues {
		issueByID[b.ID] = b
		if b.Parent != "" {
			children[b.Parent] = append(children[b.Parent], b.ID)
		}
	}

	// Recursive: effective date = max(own date, max of children's effective dates)
	cache := map[string]time.Time{}
	var compute func(id string) time.Time
	compute = func(id string) time.Time {
		if t, ok := cache[id]; ok {
			return t
		}
		b := issueByID[id]
		var best time.Time
		if b != nil {
			switch field {
			case FieldCreatedAt:
				if b.CreatedAt != nil {
					best = *b.CreatedAt
				}
			case FieldUpdatedAt:
				if b.UpdatedAt != nil {
					best = *b.UpdatedAt
				}
			}
		}
		for _, childID := range children[id] {
			if ct := compute(childID); ct.After(best) {
				best = ct
			}
		}
		cache[id] = best
		return best
	}

	for _, b := range allIssues {
		compute(b.ID)
	}
	return cache
}

// SortByEffectiveDate sorts issues by effective date, newest first.
// Issues without dates sort last. Ties are broken by title for stability.
func SortByEffectiveDate(issues []*Issue, effectiveDates map[string]time.Time) {
	slices.SortFunc(issues, func(a, b *Issue) int {
		da := effectiveDates[a.ID]
		db := effectiveDates[b.ID]
		if da.IsZero() && db.IsZero() {
			return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
		}
		if da.IsZero() {
			return 1 // no date sorts last
		}
		if db.IsZero() {
			return -1
		}
		if !da.Equal(db) {
			return db.Compare(da) // newest first
		}
		return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})
}

// SortByDueDate sorts issues by due date, soonest first.
// Issues without a due date sort last. Ties are broken by title for stability.
func SortByDueDate(issues []*Issue) {
	slices.SortFunc(issues, func(a, b *Issue) int {
		da := a.Due
		db := b.Due
		if da == nil && db == nil {
			return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
		}
		if da == nil {
			return 1 // no due date sorts last
		}
		if db == nil {
			return -1
		}
		if !da.Time.Equal(db.Time) {
			return da.Time.Compare(db.Time) // soonest first
		}
		return cmp.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})
}
