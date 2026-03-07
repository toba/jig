package changelog

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/issue"
)

// Commit represents a single git commit.
type Commit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// TimeRange represents the time window for the changelog.
type TimeRange struct {
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

// Issues groups issues by how they relate to the time range.
type Issues struct {
	Created   []*issue.Issue `json:"created"`
	Updated   []*issue.Issue `json:"updated"`
	Completed []*issue.Issue `json:"completed"`
}

// Result is the full changelog output.
type Result struct {
	GitHub  string    `json:"github,omitempty"`
	Range   TimeRange `json:"range"`
	Issues  Issues    `json:"issues"`
	Commits []Commit  `json:"commits,omitempty"`
}

// Options configures what to gather.
type Options struct {
	Since      time.Time
	Until      time.Time
	IncludeGit bool
}

// Gather filters issues into created/updated/completed buckets based on the time range.
func Gather(all []*issue.Issue, opts Options) *Result {
	r := &Result{
		Range: TimeRange{Since: opts.Since, Until: opts.Until},
		Issues: Issues{
			Created:   []*issue.Issue{},
			Updated:   []*issue.Issue{},
			Completed: []*issue.Issue{},
		},
	}

	for _, iss := range all {
		inCreated := iss.CreatedAt != nil && !iss.CreatedAt.Before(opts.Since) && iss.CreatedAt.Before(opts.Until)
		inUpdated := iss.UpdatedAt != nil && !iss.UpdatedAt.Before(opts.Since) && iss.UpdatedAt.Before(opts.Until)
		isCompleted := (iss.Status == config.StatusCompleted || iss.Status == config.StatusReview) && inUpdated

		switch {
		case isCompleted:
			r.Issues.Completed = append(r.Issues.Completed, iss)
		case inCreated:
			r.Issues.Created = append(r.Issues.Created, iss)
		case inUpdated:
			r.Issues.Updated = append(r.Issues.Updated, iss)
		}
	}

	return r
}

// GitCommits returns commits in the given time range.
func GitCommits(since, until time.Time) ([]Commit, error) {
	args := []string{
		"log",
		"--after=" + since.Format(time.RFC3339),
		"--before=" + until.Format(time.RFC3339),
		"--format=%H\t%s\t%aI",
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var commits []Commit
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		commits = append(commits, Commit{
			Hash:    parts[0][:minInt(12, len(parts[0]))],
			Subject: parts[1],
			Date:    parts[2],
		})
	}
	return commits, nil
}

// CommitTimeRange returns the time range spanned by the last N commits.
// Returns zero times if there are no commits.
func CommitTimeRange(n int) (since, until time.Time, err error) {
	args := []string{
		"log",
		fmt.Sprintf("-%d", n),
		"--format=%aI",
	}
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("git log: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return time.Time{}, time.Time{}, errors.New("no commits found")
	}

	// Most recent commit is first, oldest is last
	until, err = time.Parse(time.RFC3339, lines[0])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing newest commit date: %w", err)
	}

	since, err = time.Parse(time.RFC3339, lines[len(lines)-1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("parsing oldest commit date: %w", err)
	}

	// When there's only one commit, since == until creates a zero-width range
	// where no timestamps can match. Extend until by 1 second to include it.
	if !until.After(since) {
		until = since.Add(time.Second)
	}

	return since, until, nil
}

// ChangelogLastModified returns the author date of the last git commit that
// touched the given file. Returns zero time if the file is untracked or missing.
func ChangelogLastModified(path string) (time.Time, error) {
	out, err := exec.Command("git", "log", "-1", "--format=%aI", "--", path).Output()
	if err != nil {
		return time.Time{}, nil //nolint:nilerr // not a git repo or git not available
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return time.Time{}, nil // file untracked or never committed
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing changelog last modified date: %w", err)
	}
	return t, nil
}

// MarkdownOptions configures markdown output.
type MarkdownOptions struct {
	// Mode controls the section header style.
	// "weekly" → "## Week of Feb 16 – Feb 22, 2026"
	// "daily"  → "## Feb 20, 2026"
	// "since"  → "## Since Feb 20, 2026"
	// "append" → no new header, entries only (for --commits mode)
	Mode string
}

// FormatMarkdown produces ready-to-paste changelog markdown from a Result.
// It groups completed issues by type, formats GitHub links, and generates
// the appropriate section header. Issues whose IDs appear in existingContent
// are excluded.
func FormatMarkdown(r *Result, opts MarkdownOptions, existingContent string) string {
	completed := r.Issues.Completed
	if existingContent != "" {
		completed = excludeExisting(completed, existingContent)
	}
	if len(completed) == 0 {
		return ""
	}

	// Group by category.
	type category struct {
		heading string
		issues  []*issue.Issue
	}
	categories := []category{
		{"### ✨ Features", nil},
		{"### 🐞 Fixes", nil},
		{"### 🗜️ Tweaks", nil},
	}
	for _, iss := range completed {
		switch iss.Type {
		case "feature":
			categories[0].issues = append(categories[0].issues, iss)
		case "bug":
			categories[1].issues = append(categories[1].issues, iss)
		default: // task, epic, milestone
			categories[2].issues = append(categories[2].issues, iss)
		}
	}

	var b strings.Builder

	// Section header.
	if opts.Mode != "append" {
		b.WriteString(sectionHeader(r.Range, opts.Mode))
		b.WriteString("\n\n")
	}

	first := true
	for _, cat := range categories {
		if len(cat.issues) == 0 {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false
		b.WriteString(cat.heading)
		b.WriteString("\n\n")
		for _, iss := range cat.issues {
			b.WriteString(formatEntry(iss, r.GitHub))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func sectionHeader(tr TimeRange, mode string) string {
	switch mode {
	case "weekly":
		// Find Sunday of the week containing Since.
		sun := tr.Since
		for sun.Weekday() != time.Sunday {
			sun = sun.AddDate(0, 0, -1)
		}
		sat := sun.AddDate(0, 0, 6)
		return fmt.Sprintf("## Week of %s – %s",
			shortDate(sun), shortDateWithYear(sat))
	case "daily":
		return "## " + shortDateWithYear(tr.Since)
	case "since":
		return "## Since " + shortDateWithYear(tr.Since)
	default:
		return "## " + shortDateWithYear(tr.Since)
	}
}

func shortDate(t time.Time) string {
	return t.Format("Jan 2")
}

func shortDateWithYear(t time.Time) string {
	return t.Format("Jan 2, 2006")
}

func formatEntry(iss *issue.Issue, githubURL string) string {
	ghNum := iss.GithubIssueNumber()
	if ghNum > 0 && githubURL != "" {
		return fmt.Sprintf("- %s ([#%d](%s/issues/%d))", iss.Title, ghNum, githubURL, ghNum)
	}
	return "- " + iss.Title
}

// excludeExisting filters out issues whose ID already appears in the content.
func excludeExisting(issues []*issue.Issue, content string) []*issue.Issue {
	var filtered []*issue.Issue
	for _, iss := range issues {
		if !strings.Contains(content, iss.ID) {
			filtered = append(filtered, iss)
		}
	}
	return filtered
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
