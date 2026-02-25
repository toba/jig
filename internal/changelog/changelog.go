package changelog

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

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
		isCompleted := iss.Status == "completed" && inUpdated

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
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
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
		return time.Time{}, time.Time{}, fmt.Errorf("no commits found")
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

	return since, until, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
