package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/config"
	"github.com/toba/skill/internal/github"
)

var markCmd = &cobra.Command{
	Use:   "mark [source]",
	Short: "Update last_checked_sha/date to current HEAD",
	Long:  "Mark upstream sources as reviewed by updating their SHA and date to the current HEAD.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runMark,
}

func init() {
	rootCmd.AddCommand(markCmd)
}

func runMark(cmd *cobra.Command, args []string) error {
	client := github.NewClient()

	sources := cfg.Sources
	if len(args) > 0 {
		src := config.FindSource(cfg, args[0])
		if src == nil {
			return fmt.Errorf("source %q not found in config", args[0])
		}
		sources = []config.Source{*src}
	}

	for _, src := range sources {
		sha, err := client.GetHeadSHA(src.Repo, src.Branch)
		if err != nil {
			return fmt.Errorf("getting HEAD for %s: %w", src.Repo, err)
		}

		// Find and update the source in the original config (not the filtered copy).
		origSrc := config.FindSource(cfg, src.Repo)
		if origSrc == nil {
			return fmt.Errorf("source %s not found in config", src.Repo)
		}
		config.MarkSource(origSrc, sha)
		fmt.Printf("marked %s at %s\n", src.Repo, sha[:min(7, len(sha))])
	}

	if err := config.Save(cfgDoc, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}
