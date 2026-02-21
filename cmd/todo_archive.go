package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/issue"
	"github.com/toba/jig/internal/todo/output"
)

var archiveJSON bool

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Move completed/scrapped issues to the archive",
	Long: `Moves all issues with status "completed" or "scrapped" to the archive directory (.issues/archive/).
archived issues are preserved for project memory and remain visible in all queries.
The archive keeps the main data directory tidy while preserving project history.

Relationships (parent, blocking) are preserved in archived issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allIssues := todoStore.All()

		// Find issues with any archive status
		var archiveIssues []*issue.Issue
		for _, b := range allIssues {
			if todoCfg.IsArchiveStatus(b.Status) {
				archiveIssues = append(archiveIssues, b)
			}
		}

		if len(archiveIssues) == 0 {
			if archiveJSON {
				return output.SuccessMessage("No issues to archive")
			}
			fmt.Println("No issues with archive status to archive.")
			return nil
		}

		// Sort issues for consistent display
		issue.SortByStatusPriorityAndType(archiveIssues, todoCfg.StatusNames(), todoCfg.PriorityNames(), todoCfg.TypeNames())

		// Archive all issues with archive status
		var archived []string
		for _, b := range archiveIssues {
			if err := todoStore.Archive(b.ID); err != nil {
				if archiveJSON {
					return output.Error(output.ErrFileError, fmt.Sprintf("failed to archive issue %s: %s", b.ID, err.Error()))
				}
				return fmt.Errorf("failed to archive issue %s: %w", b.ID, err)
			}
			archived = append(archived, b.ID)
		}

		if archiveJSON {
			return output.SuccessMessage(fmt.Sprintf("Archived %d issue(s) to .issues/archive/", len(archived)))
		}

		fmt.Printf("Archived %d issue(s) to .issues/archive/\n", len(archived))
		return nil
	},
}

func init() {
	archiveCmd.Flags().BoolVar(&archiveJSON, "json", false, "Output as JSON")
	todoCmd.AddCommand(archiveCmd)
}
