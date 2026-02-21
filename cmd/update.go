package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/update"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Migrate legacy config files into .toba.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		return update.Run(configPath())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
