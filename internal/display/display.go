package display

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/toba/jig/internal/classify"
	"github.com/toba/jig/internal/config"
)

var (
	repoStyle    = lipgloss.NewStyle().Bold(true)
	branchStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))   // cyan
	relStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // gray
	highStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // red
	mediumStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // yellow
	lowStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // gray
	unclassStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	shaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))   // yellow
	authorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))   // cyan
	dateStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // gray
	notesStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	noChangeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	sepStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// SourceResult holds the check results for a single upstream source.
type SourceResult struct {
	Source  config.Source   `json:"source"`
	Commits []CommitResult `json:"commits"`
	Files   []FileResult   `json:"files,omitempty"`
}

// CommitResult holds a commit with its classified files.
type CommitResult struct {
	SHA     string              `json:"sha"`
	Message string              `json:"message"`
	Author  string              `json:"author"`
	Date    string              `json:"date"`
	Level   classify.Level      `json:"level"`
}

// FileResult holds a file with its classification.
type FileResult struct {
	Path  string         `json:"path"`
	Level classify.Level `json:"level"`
}

// RenderText writes styled terminal output for check results.
func RenderText(w io.Writer, results []SourceResult) {
	for i, r := range results {
		if i > 0 {
			fmt.Fprintln(w, sepStyle.Render("---"))
			fmt.Fprintln(w)
		}

		// Header: repo  branch  (relationship)
		header := repoStyle.Render(r.Source.Repo) + "  " +
			branchStyle.Render(r.Source.Branch) + "  " +
			relStyle.Render("("+r.Source.Relationship+")")
		fmt.Fprintln(w, header)

		if r.Source.Notes != "" {
			fmt.Fprintln(w, "  "+notesStyle.Render(r.Source.Notes))
		}

		if len(r.Commits) == 0 {
			since := "last check"
			if r.Source.LastCheckedDate != "" {
				since = r.Source.LastCheckedDate[:10]
			}
			fmt.Fprintln(w, "  "+noChangeStyle.Render("No new commits since "+since))
			fmt.Fprintln(w)
			continue
		}

		since := "first check"
		if r.Source.LastCheckedDate != "" {
			since = r.Source.LastCheckedDate[:10]
		}
		fmt.Fprintf(w, "  %d new commits since %s\n", len(r.Commits), since)
		fmt.Fprintln(w)

		// Group commits by level, display in order: HIGH, MEDIUM, LOW, UNCLASSIFIED.
		grouped := groupCommitsByLevel(r.Commits)
		for _, level := range []classify.Level{classify.High, classify.Medium, classify.Low, classify.Unclassified} {
			commits, ok := grouped[level]
			if !ok {
				continue
			}
			label := levelLabel(level, len(commits))
			fmt.Fprintln(w, "  "+label)
			for _, c := range commits {
				short := c.SHA[:min(7, len(c.SHA))]
				msg := truncate(c.Message, 50)
				line := fmt.Sprintf("    %s  %s  %s  %s",
					shaStyle.Render(short),
					msg,
					authorStyle.Render("("+c.Author+")"),
					dateStyle.Render(c.Date),
				)
				fmt.Fprintln(w, line)
			}
			fmt.Fprintln(w)
		}
	}
}

// RenderJSON writes JSON output for check results.
func RenderJSON(w io.Writer, results []SourceResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func levelLabel(level classify.Level, count int) string {
	label := fmt.Sprintf("%s (%d commits)", level.String(), count)
	if count == 1 {
		label = fmt.Sprintf("%s (1 commit)", level.String())
	}
	switch level {
	case classify.High:
		return highStyle.Render(label)
	case classify.Medium:
		return mediumStyle.Render(label)
	case classify.Low:
		return lowStyle.Render(label)
	default:
		return unclassStyle.Render(label)
	}
}

func groupCommitsByLevel(commits []CommitResult) map[classify.Level][]CommitResult {
	grouped := make(map[classify.Level][]CommitResult)
	for _, c := range commits {
		grouped[c.Level] = append(grouped[c.Level], c)
	}
	return grouped
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
