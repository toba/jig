package cmd

import (
	"errors"
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
	Use:   "ja",
	Short: "Multi-tool CLI for upstream monitoring and Claude Code security guard",
	Long:  "Multi-tool CLI combining upstream repo monitoring, companion management, and Claude Code security guard.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
