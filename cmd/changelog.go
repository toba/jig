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
	Long:  `Collects issues created, updated, or completed within a time range, optionally with git commits. By default, uses the last commit that touched CHANGELOG.md as the start date, falling back to 7 days.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initTodoCore(cmd)
	},
	RunE: runChangelog,
}

func init() {
	changelogCmd.Flags().Int("days", 0, "include issues from the last N days")
	changelogCmd.Flags().Int("commits", 0, "include issues within the last N git commits' time range")
	changelogCmd.Flags().String("since", "", "explicit start date (YYYY-MM-DD, overrides --days/--commits)")
	changelogCmd.Flags().Bool("git", false, "include git commits in output")
	changelogCmd.Flags().Bool("markdown", false, "output formatted markdown ready to paste into CHANGELOG.md")
	changelogCmd.Flags().String("changelog-file", "CHANGELOG.md", "path to existing changelog (for --markdown dedup)")
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
	case days > 0:
		since = now.AddDate(0, 0, -days)
	default:
		// Default: since last CHANGELOG.md commit, falling back to 7 days.
		lastMod, err := changelog.ChangelogLastModified("CHANGELOG.md")
		if err != nil {
			return err
		}
		if lastMod.IsZero() {
			since = now.AddDate(0, 0, -7)
		} else {
			since = lastMod
		}
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

	if includeGit || commits > 0 {
		gitCommits, err := changelog.GitCommits(since, until)
		if err != nil {
			return err
		}
		result.Commits = gitCommits
	}

	markdown, _ := cmd.Flags().GetBool("markdown")

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if markdown {
		return printChangelogMarkdown(cmd, result, commits > 0)
	}

	return printChangelogText(result)
}

func printChangelogMarkdown(cmd *cobra.Command, r *changelog.Result, isCommitMode bool) error {
	changelogFile, _ := cmd.Flags().GetString("changelog-file")

	// Determine mode for section header.
	mode := "weekly"
	if isCommitMode {
		mode = "append"
	} else if sinceStr, _ := cmd.Flags().GetString("since"); sinceStr != "" {
		mode = "since"
	} else if days, _ := cmd.Flags().GetInt("days"); days > 0 && days <= 1 {
		mode = "daily"
	}

	// Read existing changelog for dedup.
	var existing string
	if data, err := os.ReadFile(changelogFile); err == nil {
		existing = string(data)
	}

	md := changelog.FormatMarkdown(r, changelog.MarkdownOptions{Mode: mode}, existing)
	if md == "" {
		fmt.Fprintln(os.Stderr, "No new completed issues to add.")
		return nil
	}

	fmt.Print(md)
	return nil
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
