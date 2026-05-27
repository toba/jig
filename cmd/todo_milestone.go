package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/issue"
	"github.com/toba/jig/internal/todo/output"
	"github.com/toba/jig/internal/todo/ui"
)

var (
	milestoneShort string
	milestoneName  string
	milestoneDue   string
	milestoneBody  string
	milestoneJSON  bool
)

var milestoneCmd = &cobra.Command{
	Use:     "milestone",
	Aliases: []string{"ms"},
	Short:   "Manage milestones",
	Long: `Manage milestones — lightweight, optional groupings an issue may be assigned to.

Milestones are stored as files under .issues/milestones/ and are not issues.
Assign an issue to a milestone with: jig todo update <id> --milestone <milestone-id>`,
}

var milestoneCreateCmd = &cobra.Command{
	Use:     "create [name]",
	Aliases: []string{"c", "new"},
	Short:   "Create a new milestone",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := milestoneName
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			return cmdError(milestoneJSON, output.ErrValidation, "milestone name is required (pass as argument or --name)")
		}
		if err := issue.ValidateShort(milestoneShort); err != nil {
			return cmdError(milestoneJSON, output.ErrValidation, "%s", err)
		}

		m := &issue.Milestone{
			Short:       milestoneShort,
			Name:        name,
			Description: milestoneBody,
		}
		if milestoneDue != "" {
			due, err := issue.ParseDueDate(milestoneDue)
			if err != nil {
				return cmdError(milestoneJSON, output.ErrValidation, "%s", err)
			}
			m.Due = due
		}

		if err := todoStore.CreateMilestone(m); err != nil {
			return cmdError(milestoneJSON, output.ErrFileError, "failed to create milestone: %v", err)
		}

		if milestoneJSON {
			return printMilestoneJSON(m)
		}
		fmt.Println(ui.Success.Render("Created milestone ") + ui.ID.Render(m.ID) +
			" " + ui.Muted.Render("["+m.Short+"] "+m.Name))
		return nil
	},
}

var milestoneListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List milestones (ordered by due date)",
	RunE: func(cmd *cobra.Command, args []string) error {
		milestones := todoStore.MilestonesSorted()
		if milestoneJSON {
			return printMilestonesJSON(milestones)
		}
		if len(milestones) == 0 {
			fmt.Println(ui.Muted.Render("No milestones. Create one with: jig todo milestone create <name> --short <s>"))
			return nil
		}
		for _, m := range milestones {
			due := ""
			if m.Due != nil {
				due = " " + ui.Muted.Render("due "+m.Due.String())
			}
			fmt.Println(ui.ID.Render(m.ID) + "  " + ui.Secondary.Render("["+m.Short+"]") + " " + m.Name + due)
		}
		return nil
	},
}

var milestoneShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a milestone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := todoStore.GetMilestone(args[0])
		if err != nil {
			return cmdError(milestoneJSON, output.ErrNotFound, "milestone not found: %s", args[0])
		}
		if milestoneJSON {
			return printMilestoneJSON(m)
		}
		fmt.Println(ui.ID.Render(m.ID) + "  " + ui.Secondary.Render("["+m.Short+"]") + " " + m.Name)
		if m.Due != nil {
			fmt.Println(ui.Muted.Render("Due: " + m.Due.String()))
		}
		if m.Description != "" {
			fmt.Println("\n" + m.Description)
		}
		return nil
	},
}

var milestoneUpdateCmd = &cobra.Command{
	Use:     "update <id>",
	Aliases: []string{"u"},
	Short:   "Update a milestone",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := todoStore.GetMilestone(args[0])
		if err != nil {
			return cmdError(milestoneJSON, output.ErrNotFound, "milestone not found: %s", args[0])
		}
		if cmd.Flags().Changed("short") {
			if err := issue.ValidateShort(milestoneShort); err != nil {
				return cmdError(milestoneJSON, output.ErrValidation, "%s", err)
			}
			m.Short = milestoneShort
		}
		if cmd.Flags().Changed("name") {
			m.Name = milestoneName
		}
		if cmd.Flags().Changed("body") {
			m.Description = milestoneBody
		}
		if cmd.Flags().Changed("due") {
			if milestoneDue == "" {
				m.Due = nil
			} else {
				due, err := issue.ParseDueDate(milestoneDue)
				if err != nil {
					return cmdError(milestoneJSON, output.ErrValidation, "%s", err)
				}
				m.Due = due
			}
		}
		if err := todoStore.UpdateMilestone(m); err != nil {
			return cmdError(milestoneJSON, output.ErrFileError, "failed to update milestone: %v", err)
		}
		if milestoneJSON {
			return printMilestoneJSON(m)
		}
		fmt.Println(ui.Success.Render("Updated milestone ") + ui.ID.Render(m.ID))
		return nil
	},
}

var milestoneDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a milestone (does not unassign issues)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := todoStore.DeleteMilestone(args[0]); err != nil {
			return cmdError(milestoneJSON, output.ErrNotFound, "failed to delete milestone: %v", err)
		}
		if milestoneJSON {
			return output.SuccessMessage("Milestone deleted")
		}
		fmt.Println(ui.Success.Render("Deleted milestone ") + ui.ID.Render(args[0]))
		return nil
	},
}

var milestoneMigrateDryRun bool

var milestoneMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Convert legacy milestone-type issues into milestone entities",
	Long: `Converts legacy issues of type "milestone" into first-class milestone entities.

For each milestone-type issue it creates a milestone (carrying over title, due date,
description, and any GitHub milestone link), reassigns the issue's direct children to
the new milestone (clearing their parent), and deletes the old issue.

Use --dry-run to preview the changes without writing anything.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		migs, err := todoStore.MigrateMilestoneTypeIssues(milestoneMigrateDryRun)
		if err != nil {
			return cmdError(milestoneJSON, output.ErrFileError, "migration failed: %v", err)
		}
		if milestoneJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(migs)
		}
		if len(migs) == 0 {
			fmt.Println(ui.Muted.Render("No milestone-type issues to migrate."))
			return nil
		}
		verb := "Migrated"
		if milestoneMigrateDryRun {
			verb = "Would migrate"
		}
		for _, m := range migs {
			line := fmt.Sprintf("%s %s [%s] %s (%d children)", verb, m.OldIssueID, m.Short, m.Name, len(m.ChildIDs))
			fmt.Println(ui.Success.Render(line))
		}
		return nil
	},
}

func printMilestoneJSON(m *issue.Milestone) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

func printMilestonesJSON(ms []*issue.Milestone) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if ms == nil {
		ms = []*issue.Milestone{}
	}
	return enc.Encode(ms)
}

func init() {
	milestoneCreateCmd.Flags().StringVar(&milestoneShort, "short", "", "Short name (2-3 chars, shown in TUI grid)")
	milestoneCreateCmd.Flags().StringVar(&milestoneName, "name", "", "Milestone name (or pass as argument)")
	milestoneCreateCmd.Flags().StringVar(&milestoneDue, "due", "", "Due date (YYYY-MM-DD)")
	milestoneCreateCmd.Flags().StringVarP(&milestoneBody, "body", "d", "", "Description")
	milestoneCreateCmd.Flags().BoolVar(&milestoneJSON, "json", false, "Output as JSON")
	_ = milestoneCreateCmd.MarkFlagRequired("short")

	milestoneListCmd.Flags().BoolVar(&milestoneJSON, "json", false, "Output as JSON")
	milestoneShowCmd.Flags().BoolVar(&milestoneJSON, "json", false, "Output as JSON")

	milestoneUpdateCmd.Flags().StringVar(&milestoneShort, "short", "", "Short name (2-3 chars)")
	milestoneUpdateCmd.Flags().StringVar(&milestoneName, "name", "", "Milestone name")
	milestoneUpdateCmd.Flags().StringVar(&milestoneDue, "due", "", "Due date (YYYY-MM-DD, empty to clear)")
	milestoneUpdateCmd.Flags().StringVarP(&milestoneBody, "body", "d", "", "Description")
	milestoneUpdateCmd.Flags().BoolVar(&milestoneJSON, "json", false, "Output as JSON")

	milestoneDeleteCmd.Flags().BoolVar(&milestoneJSON, "json", false, "Output as JSON")

	milestoneMigrateCmd.Flags().BoolVar(&milestoneMigrateDryRun, "dry-run", false, "Preview changes without writing")
	milestoneMigrateCmd.Flags().BoolVar(&milestoneJSON, "json", false, "Output as JSON")

	milestoneCmd.AddCommand(milestoneCreateCmd, milestoneListCmd, milestoneShowCmd, milestoneUpdateCmd, milestoneDeleteCmd, milestoneMigrateCmd)
	todoCmd.AddCommand(milestoneCmd)
}
