package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/classify"
	"github.com/toba/skill/internal/config"
	"github.com/toba/skill/internal/display"
	"github.com/toba/skill/internal/github"
)

var checkCmd = &cobra.Command{
	Use:   "check [source]",
	Short: "Fetch and display changes grouped by relevance",
	Long:  "Check upstream repositories for changes since last review. Optionally filter to a specific source.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCheck,
}

func init() {
	upstreamCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	sources := cfg.Sources
	if len(args) > 0 {
		src := config.FindSource(cfg, args[0])
		if src == nil {
			return fmt.Errorf("source %q not found in config", args[0])
		}
		sources = []config.Source{*src}
	}

	client := github.NewClient()
	results := make([]display.SourceResult, len(sources))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, src := range sources {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := checkSource(client, src)
			if err != nil {
				mu.Lock()
				fmt.Fprintf(os.Stderr, "warning: %s: %v\n", src.Repo, err)
				mu.Unlock()
				results[i] = display.SourceResult{Source: src}
				return
			}
			results[i] = *result
		}()
	}
	wg.Wait()

	if jsonOut {
		return display.RenderJSON(os.Stdout, results)
	}
	display.RenderText(os.Stdout, results)
	return nil
}

func checkSource(client github.Client, src config.Source) (*display.SourceResult, error) {
	result := &display.SourceResult{Source: src}

	var commits []github.Commit
	var aggregateFiles []github.File

	if src.LastCheckedSHA == "" {
		// First run: fetch last 30 commits.
		var err error
		commits, err = client.GetCommits(src.Repo, src.Branch, 30)
		if err != nil {
			return nil, fmt.Errorf("fetching commits: %w", err)
		}

		// Fetch file details for the most recent 5 commits only.
		limit := min(5, len(commits))
		for i := range limit {
			detail, err := client.GetCommitDetail(src.Repo, commits[i].SHA)
			if err != nil {
				continue
			}
			commits[i].Files = detail.Files
			aggregateFiles = append(aggregateFiles, detail.Files...)
		}
	} else {
		// Compare since last checked SHA.
		cmp, err := client.Compare(src.Repo, src.LastCheckedSHA, src.Branch)
		if err != nil {
			// Fallback: force-push scenario (404 from compare).
			if strings.Contains(err.Error(), "404") && src.LastCheckedDate != "" {
				commits, err = client.GetCommitsSince(src.Repo, src.Branch, src.LastCheckedDate, 100)
				if err != nil {
					return nil, fmt.Errorf("fetching commits since date: %w", err)
				}
				// Fetch file details for up to 5 commits.
				limit := min(5, len(commits))
				for i := range limit {
					detail, derr := client.GetCommitDetail(src.Repo, commits[i].SHA)
					if derr != nil {
						continue
					}
					commits[i].Files = detail.Files
					aggregateFiles = append(aggregateFiles, detail.Files...)
				}
			} else {
				return nil, fmt.Errorf("comparing: %w", err)
			}
		} else {
			commits = cmp.Commits
			aggregateFiles = cmp.Files
		}
	}

	if len(commits) == 0 {
		return result, nil
	}

	// Classify aggregate files.
	filePaths := make([]string, 0, len(aggregateFiles))
	for _, f := range aggregateFiles {
		filePaths = append(filePaths, f.Filename)
	}
	fileResults := classify.Classify(filePaths, src.Paths)

	// Build file results for JSON output.
	for _, fr := range fileResults {
		result.Files = append(result.Files, display.FileResult{
			Path:  fr.Path,
			Level: fr.Level,
		})
	}

	// Classify each commit by its highest-level file.
	for _, c := range commits {
		level := classify.Unclassified
		if len(c.Files) > 0 {
			cFilePaths := make([]string, 0, len(c.Files))
			for _, f := range c.Files {
				cFilePaths = append(cFilePaths, f.Filename)
			}
			cResults := classify.Classify(cFilePaths, src.Paths)
			level = classify.MaxLevel(cResults)
		} else {
			// No per-commit files: derive from aggregate classification.
			level = classify.MaxLevel(fileResults)
		}

		result.Commits = append(result.Commits, display.CommitResult{
			SHA:     c.SHA,
			Message: c.Message,
			Author:  c.Author,
			Date:    c.Date.Format("2006-01-02"),
			Level:   level,
		})
	}

	return result, nil
}
