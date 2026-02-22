package cmd

import "github.com/spf13/cobra"

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage project tag registry",
}

func init() {
	todoCmd.AddCommand(tagsCmd)
}
