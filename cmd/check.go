package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/classify"
	"github.com/toba/jig/internal/config"
	"github.com/toba/jig/internal/display"
	"github.com/toba/jig/internal/github"
	"golang.org/x/sync/errgroup"
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

type checkResult struct {
	display display.SourceResult
	headSHA string // latest commit SHA from fetched data
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
	results := make([]checkResult, len(sources))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, src := range sources {
		wg.Go(func() {
			result, headSHA, err := checkSource(client, src)
			if err != nil {
				mu.Lock()
				fmt.Fprintf(os.Stderr, "warning: %s: %v\n", src.Repo, err)
				mu.Unlock()
				results[i] = checkResult{display: display.SourceResult{Source: src}}
				return
			}
			results[i] = checkResult{display: *result, headSHA: headSHA}
		})
	}
	wg.Wait()

	// Update last_checked for sources that returned data.
	dirty := false
	for _, r := range results {
		if r.headSHA == "" {
			continue
		}
		origSrc := config.FindSource(cfg, r.display.Source.Repo)
		if origSrc == nil {
			continue
		}
		config.MarkSource(origSrc, r.headSHA)
		dirty = true
	}
	if dirty {
		if err := config.Save(cfgDoc, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: saving last_checked: %v\n", err)
		}
	}

	displayResults := make([]display.SourceResult, len(results))
	for i, r := range results {
		displayResults[i] = r.display
	}

	if jsonOut {
		return display.RenderJSON(os.Stdout, displayResults)
	}
	display.RenderText(os.Stdout, displayResults)
	return nil
}

func checkSource(client github.Client, src config.Source) (*display.SourceResult, string, error) {
	result := &display.SourceResult{Source: src}

	var commits []github.Commit
	var aggregateFiles []github.File

	if src.LastCheckedSHA == "" {
		// First run: fetch last 30 commits.
		var err error
		commits, err = client.GetCommits(src.Repo, src.Branch, 30)
		if err != nil {
			return nil, "", fmt.Errorf("fetching commits: %w", err)
		}

		// Fetch file details for the most recent 5 commits in parallel.
		aggregateFiles = fetchCommitDetails(client, src.Repo, commits, min(5, len(commits)))
	} else {
		// Compare since last checked SHA.
		cmp, err := client.Compare(src.Repo, src.LastCheckedSHA, src.Branch)
		if err != nil {
			// Fallback: force-push scenario (404 from compare).
			if strings.Contains(err.Error(), "404") && src.LastCheckedDate != "" {
				commits, err = client.GetCommitsSince(src.Repo, src.Branch, src.LastCheckedDate, 100)
				if err != nil {
					return nil, "", fmt.Errorf("fetching commits since date: %w", err)
				}
				// Fetch file details for up to 5 commits in parallel.
				aggregateFiles = fetchCommitDetails(client, src.Repo, commits, min(5, len(commits)))
			} else {
				return nil, "", fmt.Errorf("comparing: %w", err)
			}
		} else {
			commits = cmp.Commits
			aggregateFiles = cmp.Files
		}
	}

	if len(commits) == 0 {
		return result, "", nil
	}

	// The most recent commit SHA becomes the new last_checked reference.
	headSHA := commits[0].SHA

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

	return result, headSHA, nil
}

// fetchCommitDetails fetches file details for up to limit commits in parallel,
// setting each commit's Files field and returning the aggregate file list.
func fetchCommitDetails(client github.Client, repo string, commits []github.Commit, limit int) []github.File {
	type indexedFiles struct {
		idx   int
		files []github.File
	}

	var g errgroup.Group
	ch := make(chan indexedFiles, limit)

	for i := range limit {
		g.Go(func() error {
			detail, err := client.GetCommitDetail(repo, commits[i].SHA)
			if err != nil {
				return nil // non-fatal
			}
			ch <- indexedFiles{idx: i, files: detail.Files}
			return nil
		})
	}

	go func() {
		_ = g.Wait()
		close(ch)
	}()

	var aggregateFiles []github.File
	for result := range ch {
		commits[result.idx].Files = result.files
		aggregateFiles = append(aggregateFiles, result.files...)
	}

	return aggregateFiles
}
