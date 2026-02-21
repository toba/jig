package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/tui"
)

// tuiAliasCmd is a top-level alias for "jig todo tui".
var tuiAliasCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open the interactive TUI (alias for 'jig todo tui')",
	Long:  `Opens an interactive terminal user interface for browsing and managing issues.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return initTodoCore(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run(todoStore, todoCfg)
	},
}

func init() {
	rootCmd.AddCommand(tuiAliasCmd)
}
