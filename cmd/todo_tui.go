package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/tui"
)

var todoTuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Open the interactive TUI",
	Long:  `Opens an interactive terminal user interface for browsing and managing issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run(todoStore, todoCfg)
	},
}

func init() {
	todoCmd.AddCommand(todoTuiCmd)
}
