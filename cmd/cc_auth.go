package cmd

import "github.com/spf13/cobra"

var ccAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage per-alias Claude credentials",
}

func init() {
	ccCmd.AddCommand(ccAuthCmd)
}
