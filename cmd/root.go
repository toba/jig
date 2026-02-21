package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/config"
	"github.com/toba/skill/internal/nope"
)

var (
	cfgPath string
	jsonOut bool
	cfg     *config.Config
	cfgDoc  *config.Document
)

var rootCmd = &cobra.Command{
	Use:   "skill",
	Short: "Check upstream repos for changes",
	Long:  "Monitor upstream repositories for changes, classify them by relevance, and track what you've reviewed.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for init, version, and nope commands.
		if cmd.Name() == "init" || cmd.Name() == "version" {
			return nil
		}
		if cmd.Name() == "nope" || (cmd.Parent() != nil && cmd.Parent().Name() == "nope") {
			return nil
		}
		path := cfgPath
		if path == "" {
			path = ".toba.yaml"
		}
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
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "path to config file (default .toba.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output as JSON")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var exitErr nope.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

func configPath() string {
	if cfgPath != "" {
		return cfgPath
	}
	return ".toba.yaml"
}
