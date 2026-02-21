package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/integration"
)

var syncUnlinkJSON bool

var syncUnlinkCmd = &cobra.Command{
	Use:   "unlink <issue-id>",
	Short: "Remove the link between an issue and its external task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		ctx := context.Background()

		integ, err := integration.Detect(todoCfg.Sync, todoStore)
		if err != nil {
			return fmt.Errorf("detecting integration: %w", err)
		}
		if integ == nil {
			return fmt.Errorf("no integration configured")
		}

		result, err := integ.Unlink(ctx, issueID)
		if err != nil {
			return err
		}

		b, getErr := todoStore.Get(issueID)
		title := issueID
		if getErr == nil {
			title = b.Title
		}

		if syncUnlinkJSON {
			return outputUnlinkJSON(issueID, title, result.ExternalID, result.Action)
		}

		switch result.Action {
		case integration.ActionNotLinked:
			fmt.Printf("Skipped: %s is not linked to an external task\n", issueID)
		case integration.ActionUnlinked:
			fmt.Printf("Unlinked: %s (was %s)\n", issueID, result.ExternalID)
		}
		return nil
	},
}

func init() {
	syncUnlinkCmd.Flags().BoolVar(&syncUnlinkJSON, "json", false, "Output as JSON")
	todoSyncCmd.AddCommand(syncUnlinkCmd)
}

func outputUnlinkJSON(issueID, issueTitle, externalID, action string) error {
	result := map[string]string{
		"issue_id":    issueID,
		"issue_title": issueTitle,
		"action":      action,
	}
	if externalID != "" {
		result["external_id"] = externalID
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
