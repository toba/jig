package cmd

import (
	"github.com/spf13/cobra"
)

// syncAliasCmd is a top-level alias for "jig todo sync".
var syncAliasCmd = &cobra.Command{
	Use:   "sync [issue-id...]",
	Short: "Sync issues to external integrations (alias for 'jig todo sync')",
	Long: `Syncs issues to an external integration configured in .jig.yaml.

This is an alias for 'jig todo sync'. See 'jig todo sync --help' for full details.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initTodoCore(cmd)
	},
	RunE: runSync,
}

// syncAliasCheckCmd is a top-level alias for "jig todo sync check".
var syncAliasCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Verify integration configuration and sync state health",
	RunE:  syncCheckCmd.RunE,
}

// syncAliasLinkCmd is a top-level alias for "jig todo sync link".
var syncAliasLinkCmd = &cobra.Command{
	Use:   "link <issue-id> <external-id>",
	Short: "Link an issue to an existing external task",
	Args:  cobra.ExactArgs(2),
	RunE:  syncLinkCmd.RunE,
}

// syncAliasUnlinkCmd is a top-level alias for "jig todo sync unlink".
var syncAliasUnlinkCmd = &cobra.Command{
	Use:   "unlink <issue-id>",
	Short: "Remove the link between an issue and its external task",
	Args:  cobra.ExactArgs(1),
	RunE:  syncUnlinkCmd.RunE,
}

func init() {
	syncAliasCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Show what would be done without making changes")
	syncAliasCmd.Flags().BoolVar(&syncForce, "force", false, "Force update even if unchanged")
	syncAliasCmd.Flags().BoolVar(&syncNoRelationships, "no-relationships", false, "Skip syncing blocking relationships as dependencies")
	syncAliasCmd.Flags().BoolVar(&syncJSON, "json", false, "Output results as JSON")

	syncAliasCheckCmd.Flags().BoolVar(&syncCheckSkipAPI, "skip-api", false, "Skip API checks (offline validation only)")
	syncAliasCheckCmd.Flags().BoolVar(&syncCheckJSON, "json", false, "Output as JSON")

	syncAliasLinkCmd.Flags().BoolVar(&syncLinkJSON, "json", false, "Output as JSON")
	syncAliasUnlinkCmd.Flags().BoolVar(&syncUnlinkJSON, "json", false, "Output as JSON")

	syncAliasCmd.AddCommand(syncAliasCheckCmd)
	syncAliasCmd.AddCommand(syncAliasLinkCmd)
	syncAliasCmd.AddCommand(syncAliasUnlinkCmd)
	rootCmd.AddCommand(syncAliasCmd)
}
