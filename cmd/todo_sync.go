package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/integration"
	"github.com/toba/jig/internal/todo/issue"
)

var (
	syncDryRun          bool
	syncForce           bool
	syncNoRelationships bool
	syncJSON            bool
)

var todoSyncCmd = &cobra.Command{
	Use:   "sync [issue-id...]",
	Short: "Sync issues to external integrations",
	Long: `Syncs issues to an external integration configured in .jig.yaml.

If issue IDs are provided, only those issues are synced. Otherwise, all issues
matching the sync filter are synced.

To enable sync, add a sync section to your .jig.yaml config file under todo:

  todo:
    sync:
      github:
        repo: owner/repo`,
	RunE: runSync,
}

func init() {
	todoSyncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Show what would be done without making changes")
	todoSyncCmd.Flags().BoolVar(&syncForce, "force", false, "Force update even if unchanged")
	todoSyncCmd.Flags().BoolVar(&syncNoRelationships, "no-relationships", false, "Skip syncing blocking relationships as dependencies")
	todoSyncCmd.Flags().BoolVar(&syncJSON, "json", false, "Output results as JSON")
	todoCmd.AddCommand(todoSyncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	integ, err := integration.Detect(todoCfg.Sync, todoStore)
	if err != nil {
		return fmt.Errorf("detecting integration: %w", err)
	}
	if integ == nil {
		if syncJSON {
			return outputSyncJSON(nil)
		}
		fmt.Println("No integration configured. Add a sync section (clickup or github) to .jig.yaml.")
		return nil
	}

	var issueList []*issue.Issue
	if len(args) == 0 {
		issueList = todoStore.All()
	}
	if len(args) > 0 {
		for _, id := range args {
			b, err := todoStore.Get(id)
			if err != nil {
				return fmt.Errorf("issue not found: %s", id)
			}
			issueList = append(issueList, b)
		}
	}

	if len(issueList) == 0 {
		if syncJSON {
			return outputSyncJSON(nil)
		}
		fmt.Println("No issues to sync")
		return nil
	}

	opts := integration.SyncOptions{
		DryRun:          syncDryRun,
		Force:           syncForce,
		NoRelationships: syncNoRelationships,
	}

	if !syncJSON {
		fmt.Printf("Syncing %d issues to %s", len(issueList), integ.Name())
		if len(issueList) >= 5 {
			fmt.Print(" ")
			opts.OnProgress = func(result integration.SyncResult, completed, total int) {
				if result.Error != nil {
					fmt.Print("x")
				} else {
					fmt.Print(".")
				}
			}
		}
	}

	results, err := integ.Sync(ctx, issueList, opts)

	if !syncJSON {
		fmt.Println()
	}

	if err != nil {
		return err
	}

	if results == nil {
		if syncJSON {
			return outputSyncJSON(nil)
		}
		fmt.Println("All issues up to date")
		return nil
	}

	if syncJSON {
		return outputSyncJSON(results)
	}
	return outputSyncText(results)
}

func outputSyncJSON(results []integration.SyncResult) error {
	type jsonResult struct {
		IssueID     string `json:"issue_id"`
		IssueTitle  string `json:"issue_title"`
		ExternalID  string `json:"external_id,omitempty"`
		ExternalURL string `json:"external_url,omitempty"`
		Action      string `json:"action"`
		Error       string `json:"error,omitempty"`
	}

	if results == nil {
		fmt.Println("[]")
		return nil
	}

	jsonResults := make([]jsonResult, len(results))
	for i, r := range results {
		jsonResults[i] = jsonResult{
			IssueID:     r.IssueID,
			IssueTitle:  r.IssueTitle,
			ExternalID:  r.ExternalID,
			ExternalURL: r.ExternalURL,
			Action:      r.Action,
		}
		if r.Error != nil {
			jsonResults[i].Error = r.Error.Error()
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonResults)
}

func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen] + "\u2026"
}

func outputSyncText(results []integration.SyncResult) error {
	var created, updated, unchanged, skipped, errors int

	for _, r := range results {
		switch r.Action {
		case integration.ActionCreated:
			created++
			fmt.Printf("  Created: %s \u2192 %s \"%s\"\n", r.IssueID, r.ExternalURL, truncateTitle(r.IssueTitle, 20))
		case integration.ActionUpdated:
			updated++
			fmt.Printf("  Updated: %s \u2192 %s \"%s\"\n", r.IssueID, r.ExternalURL, truncateTitle(r.IssueTitle, 20))
		case integration.ActionUnchanged:
			unchanged++
		case integration.ActionSkipped:
			skipped++
		case integration.ActionWouldCreate:
			fmt.Printf("  Would create: %s - %s\n", r.IssueID, r.IssueTitle)
		case integration.ActionWouldUpdate:
			fmt.Printf("  Would update: %s - %s\n", r.IssueID, r.IssueTitle)
		case integration.ActionError:
			errors++
			fmt.Printf("  Error: %s - %v\n", r.IssueID, r.Error)
		}
	}

	fmt.Printf("\nSummary: %d created, %d updated, %d unchanged, %d skipped, %d errors\n",
		created, updated, unchanged, skipped, errors)
	return nil
}
