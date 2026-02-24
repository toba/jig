package cmd

import (
	"cmp"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/graph"
	"github.com/toba/jig/internal/todo/issue"
)

//go:embed todo_roadmap.tmpl
var roadmapTemplateContent string

var (
	roadmapJSON        bool
	roadmapIncludeDone bool
	roadmapStatus      []string
	roadmapNoStatus    []string
	roadmapNoLinks     bool
	roadmapLinkPrefix  string
)

type roadmapData struct {
	Milestones  []milestoneGroup  `json:"milestones"`
	Unscheduled *unscheduledGroup `json:"unscheduled,omitempty"`
}

type unscheduledGroup struct {
	Epics []epicGroup    `json:"epics,omitempty"`
	Other []*issue.Issue `json:"other,omitempty"`
}

type milestoneGroup struct {
	Milestone *issue.Issue   `json:"milestone"`
	Epics     []epicGroup    `json:"epics,omitempty"`
	Other     []*issue.Issue `json:"other,omitempty"`
}

type epicGroup struct {
	Epic  *issue.Issue   `json:"epic"`
	Items []*issue.Issue `json:"items,omitempty"`
}

var roadmapCmd = &cobra.Command{
	Use:   "roadmap",
	Short: "Generate a Markdown roadmap from milestones and epics",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolver := &graph.Resolver{Core: todoStore}
		allIssues, err := resolver.Query().Issues(context.Background(), nil)
		if err != nil {
			return fmt.Errorf("querying issues: %w", err)
		}

		data := buildRoadmap(allIssues, roadmapIncludeDone, roadmapStatus, roadmapNoStatus)

		if roadmapJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		}

		links := !roadmapNoLinks
		linkPrefix := roadmapLinkPrefix
		if links && linkPrefix == "" {
			linkPrefix = defaultLinkPrefix()
		}
		md := renderRoadmapMarkdown(data, links, linkPrefix)
		fmt.Print(md)
		return nil
	},
}

func buildRoadmap(allIssues []*issue.Issue, includeDone bool, statusFilter, noStatusFilter []string) *roadmapData {
	byID := make(map[string]*issue.Issue)
	for _, b := range allIssues {
		byID[b.ID] = b
	}

	children := make(map[string][]*issue.Issue)
	for _, b := range allIssues {
		if b.Parent != "" {
			children[b.Parent] = append(children[b.Parent], b)
		}
	}

	var milestones []*issue.Issue
	for _, b := range allIssues {
		if b.Type != todoconfig.TypeMilestone {
			continue
		}
		if len(statusFilter) > 0 && !containsStatus(statusFilter, b.Status) {
			continue
		}
		if len(noStatusFilter) > 0 && containsStatus(noStatusFilter, b.Status) {
			continue
		}
		milestones = append(milestones, b)
	}

	sortByStatusThenCreated(milestones, todoCfg)

	var milestoneGroups []milestoneGroup
	for _, m := range milestones {
		group := buildMilestoneGroup(m, children, includeDone)
		if len(group.Epics) > 0 || len(group.Other) > 0 {
			milestoneGroups = append(milestoneGroups, group)
		}
	}

	underMilestone := make(map[string]bool)
	for _, m := range milestones {
		underMilestone[m.ID] = true
		for _, child := range children[m.ID] {
			underMilestone[child.ID] = true
			if child.Type == todoconfig.TypeEpic {
				for _, epicChild := range children[child.ID] {
					underMilestone[epicChild.ID] = true
				}
			}
		}
	}

	var unscheduledEpics []epicGroup
	for _, b := range allIssues {
		if b.Type != todoconfig.TypeEpic {
			continue
		}
		if underMilestone[b.ID] {
			continue
		}
		epicItems := filterChildren(children[b.ID], includeDone)
		if len(epicItems) > 0 {
			sortByTypeThenStatus(epicItems, todoCfg)
			unscheduledEpics = append(unscheduledEpics, epicGroup{Epic: b, Items: epicItems})
		}
	}

	slices.SortFunc(unscheduledEpics, func(a, b epicGroup) int {
		return cmp.Compare(a.Epic.Title, b.Epic.Title)
	})

	var orphanItems []*issue.Issue
	for _, b := range allIssues {
		if b.Type == todoconfig.TypeMilestone || b.Type == todoconfig.TypeEpic {
			continue
		}
		if underMilestone[b.ID] {
			continue
		}
		if b.Parent != "" {
			continue
		}
		if !includeDone && todoCfg.IsArchiveStatus(b.Status) {
			continue
		}
		orphanItems = append(orphanItems, b)
	}

	sortByTypeThenStatus(orphanItems, todoCfg)

	var unscheduled *unscheduledGroup
	if len(unscheduledEpics) > 0 || len(orphanItems) > 0 {
		unscheduled = &unscheduledGroup{
			Epics: unscheduledEpics,
			Other: orphanItems,
		}
	}

	return &roadmapData{
		Milestones:  milestoneGroups,
		Unscheduled: unscheduled,
	}
}

func buildMilestoneGroup(m *issue.Issue, children map[string][]*issue.Issue, includeDone bool) milestoneGroup {
	group := milestoneGroup{Milestone: m}

	directChildren := children[m.ID]

	var epics []*issue.Issue
	for _, child := range directChildren {
		if child.Type == todoconfig.TypeEpic {
			epics = append(epics, child)
		}
	}

	for _, epic := range epics {
		epicItems := filterChildren(children[epic.ID], includeDone)
		if len(epicItems) > 0 {
			sortByTypeThenStatus(epicItems, todoCfg)
			group.Epics = append(group.Epics, epicGroup{Epic: epic, Items: epicItems})
		}
	}

	var other []*issue.Issue
	for _, child := range directChildren {
		if child.Type == todoconfig.TypeEpic {
			continue
		}
		if includeDone || !todoCfg.IsArchiveStatus(child.Status) {
			other = append(other, child)
		}
	}

	slices.SortFunc(group.Epics, func(a, b epicGroup) int {
		return cmp.Compare(a.Epic.Title, b.Epic.Title)
	})

	sortByTypeThenStatus(other, todoCfg)
	group.Other = other

	return group
}

func filterChildren(children []*issue.Issue, includeDone bool) []*issue.Issue {
	if includeDone {
		result := make([]*issue.Issue, len(children))
		copy(result, children)
		return result
	}

	var filtered []*issue.Issue
	for _, b := range children {
		if !todoCfg.IsArchiveStatus(b.Status) {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func containsStatus(statuses []string, status string) bool {
	return slices.Contains(statuses, status)
}

func sortByStatusThenCreated(issues []*issue.Issue, cfg interface{ StatusNames() []string }) {
	statusOrder := make(map[string]int)
	for i, s := range cfg.StatusNames() {
		statusOrder[s] = i
	}

	slices.SortFunc(issues, func(a, b *issue.Issue) int {
		if c := cmp.Compare(statusOrder[a.Status], statusOrder[b.Status]); c != 0 {
			return c
		}
		if a.CreatedAt != nil && b.CreatedAt != nil {
			return a.CreatedAt.Compare(*b.CreatedAt)
		}
		return cmp.Compare(a.ID, b.ID)
	})
}

func sortByTypeThenStatus(issues []*issue.Issue, cfg interface {
	StatusNames() []string
	TypeNames() []string
}) {
	statusOrder := make(map[string]int)
	for i, s := range cfg.StatusNames() {
		statusOrder[s] = i
	}
	typeOrder := make(map[string]int)
	for i, t := range cfg.TypeNames() {
		typeOrder[t] = i
	}

	slices.SortFunc(issues, func(a, b *issue.Issue) int {
		if c := cmp.Compare(typeOrder[a.Type], typeOrder[b.Type]); c != 0 {
			return c
		}
		if c := cmp.Compare(statusOrder[a.Status], statusOrder[b.Status]); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
}

func renderRoadmapMarkdown(data *roadmapData, links bool, linkPrefix string) string {
	tmpl := template.Must(
		template.New("roadmap").Funcs(template.FuncMap{
			"firstParagraph": firstParagraph,
			"typeBadge":      typeBadge,
			"beanRef": func(b *issue.Issue) string {
				return renderIssueRef(b, links, linkPrefix)
			},
		}).Parse(roadmapTemplateContent),
	)

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		panic(err)
	}
	return sb.String()
}

func renderIssueRef(b *issue.Issue, asLink bool, linkPrefix string) string {
	if !asLink {
		return "(" + b.ID + ")"
	}
	if linkPrefix == "" {
		return fmt.Sprintf("([%s](%s))", b.ID, b.Path)
	}
	if !strings.HasSuffix(linkPrefix, "/") {
		linkPrefix += "/"
	}
	return fmt.Sprintf("([%s](%s%s))", b.ID, linkPrefix, b.Path)
}

func typeBadge(b *issue.Issue) string {
	if b.Type == "" {
		return ""
	}
	colors := map[string]string{
		todoconfig.TypeBug:       "d73a4a",
		todoconfig.TypeFeature:   "0e8a16",
		todoconfig.TypeTask:      "1d76db",
		todoconfig.TypeEpic:      "5319e7",
		todoconfig.TypeMilestone: "fbca04",
	}
	color := colors[b.Type]
	if color == "" {
		color = "gray"
	}
	return fmt.Sprintf("![%s](https://img.shields.io/badge/%s-%s?style=flat-square)", b.Type, b.Type, color)
}

func defaultLinkPrefix() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	rel, err := filepath.Rel(cwd, todoStore.Root())
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

func firstParagraph(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	lines := strings.Split(body, "\n")
	var para []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			break
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		para = append(para, strings.TrimSpace(line))
	}

	const maxDescriptionLen = 200

	result := strings.Join(para, " ")
	if len(result) > maxDescriptionLen {
		result = result[:maxDescriptionLen-3] + "..."
	}
	return result
}

func init() {
	roadmapCmd.Flags().BoolVar(&roadmapJSON, "json", false, "Output as JSON")
	roadmapCmd.Flags().BoolVar(&roadmapIncludeDone, "include-done", false, "Include completed items")
	roadmapCmd.Flags().StringArrayVar(&roadmapStatus, "status", nil, "Filter milestones by status (can be repeated)")
	roadmapCmd.Flags().StringArrayVar(&roadmapNoStatus, "no-status", nil, "Exclude milestones by status (can be repeated)")
	roadmapCmd.Flags().BoolVar(&roadmapNoLinks, "no-links", false, "Don't render issue IDs as markdown links")
	roadmapCmd.Flags().StringVar(&roadmapLinkPrefix, "link-prefix", "", "URL prefix for links")
	todoCmd.AddCommand(roadmapCmd)
}
