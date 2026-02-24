package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/integration"
)

var syncLinkJSON bool

var syncLinkCmd = &cobra.Command{
	Use:   "link <issue-id> <external-id>",
	Short: "Link an issue to an existing external task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueID := args[0]
		externalID := args[1]
		ctx := context.Background()

		integ, err := integration.Detect(todoCfg.Sync, todoStore)
		if err != nil {
			return fmt.Errorf("detecting integration: %w", err)
		}
		if integ == nil {
			return errors.New("no integration configured")
		}

		result, err := integ.Link(ctx, issueID, externalID)
		if err != nil {
			return err
		}

		b, getErr := todoStore.Get(issueID)
		title := issueID
		if getErr == nil {
			title = b.Title
		}

		if syncLinkJSON {
			return outputLinkJSON(issueID, title, externalID, result.Action)
		}

		switch result.Action {
		case integration.ActionAlreadyLinked:
			fmt.Printf("Skipped: %s already linked to %s\n", issueID, externalID)
		case integration.ActionLinked:
			fmt.Printf("Linked: %s → %s\n", issueID, externalID)
		}
		return nil
	},
}

func init() {
	syncLinkCmd.Flags().BoolVar(&syncLinkJSON, "json", false, "Output as JSON")
	todoSyncCmd.AddCommand(syncLinkCmd)
}

func outputLinkJSON(issueID, issueTitle, externalID, action string) error {
	result := map[string]string{
		"issue_id":    issueID,
		"issue_title": issueTitle,
		"external_id": externalID,
		"action":      action,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
