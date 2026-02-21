package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/config"
)

var upstreamCmd = &cobra.Command{
	Use:   "upstream",
	Short: "Monitor upstream repositories for changes",
	Long:  "Commands for checking upstream repos, classifying changes by relevance, and tracking what you've reviewed.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// init doesn't need config loaded.
		if cmd.Name() == "init" {
			return nil
		}
		path := configPath()
		doc, c, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		cfg = c
		cfgDoc = doc
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upstreamCmd)
}
