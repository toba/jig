package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/changelog"
	"github.com/toba/jig/internal/todo/issue"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Gather recent issues and commits for changelog generation",
	Long:  `Collects issues created, updated, or completed within a time range, optionally with git commits.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initTodoCore(cmd)
	},
	RunE: runChangelog,
}

func init() {
	changelogCmd.Flags().Int("days", 0, "include issues from the last N days (default 7)")
	changelogCmd.Flags().Int("commits", 0, "include issues within the last N git commits' time range")
	changelogCmd.Flags().String("since", "", "explicit start date (YYYY-MM-DD, overrides --days/--commits)")
	changelogCmd.Flags().Bool("git", false, "include git commits in output")
	rootCmd.AddCommand(changelogCmd)
}

func runChangelog(cmd *cobra.Command, _ []string) error {
	days, _ := cmd.Flags().GetInt("days")
	commits, _ := cmd.Flags().GetInt("commits")
	sinceStr, _ := cmd.Flags().GetString("since")
	includeGit, _ := cmd.Flags().GetBool("git")

	now := time.Now()
	var since, until time.Time
	until = now

	switch {
	case sinceStr != "":
		t, err := time.Parse("2006-01-02", sinceStr)
		if err != nil {
			return fmt.Errorf("invalid --since date %q: expected YYYY-MM-DD", sinceStr)
		}
		since = t
	case commits > 0:
		var err error
		since, until, err = changelog.CommitTimeRange(commits)
		if err != nil {
			return fmt.Errorf("determining commit time range: %w", err)
		}
	default:
		if days == 0 {
			days = 7
		}
		since = now.AddDate(0, 0, -days)
	}

	all := todoStore.All()
	opts := changelog.Options{
		Since:      since,
		Until:      until,
		IncludeGit: includeGit,
	}
	result := changelog.Gather(all, opts)

	// Add GitHub repo URL from sync config if available.
	if ghCfg := todoCfg.SyncConfig("github"); ghCfg != nil {
		if repo, ok := ghCfg["repo"].(string); ok && repo != "" {
			result.GitHub = "https://github.com/" + repo
		}
	}

	if includeGit {
		gitCommits, err := changelog.GitCommits(since, until)
		if err != nil {
			return err
		}
		result.Commits = gitCommits
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	return printChangelogText(result)
}

func printChangelogText(r *changelog.Result) error {
	fmt.Printf("Changelog: %s to %s\n\n",
		r.Range.Since.Format("2006-01-02"),
		r.Range.Until.Format("2006-01-02"))

	printIssueSection("Completed", r.Issues.Completed)
	printIssueSection("Created", r.Issues.Created)
	printIssueSection("Updated", r.Issues.Updated)

	if len(r.Commits) > 0 {
		fmt.Println("## Commits")
		for _, c := range r.Commits {
			fmt.Printf("  %s %s\n", c.Hash, c.Subject)
		}
		fmt.Println()
	}

	return nil
}

func printIssueSection(heading string, issues []*issue.Issue) {
	if len(issues) == 0 {
		return
	}
	fmt.Printf("## %s\n", heading)
	for _, iss := range issues {
		prefix := ""
		if iss.Type != "" {
			prefix = fmt.Sprintf("[%s] ", iss.Type)
		}
		fmt.Printf("  %s%s (%s)\n", prefix, iss.Title, iss.ID)
	}
	fmt.Println()
}
