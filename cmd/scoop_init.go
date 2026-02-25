package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/scoop"
)

var (
	scoopInitBucket  string
	scoopInitTag     string
	scoopInitRepo    string
	scoopInitDesc    string
	scoopInitLicense string
	scoopInitDryRun  bool
)

var scoopInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up a new Scoop bucket with automated releases",
	Long: `Creates a companion Scoop bucket repo, pushes an initial manifest and README,
and injects an update-scoop job into the source repo's release workflow.

This is a one-time setup command. After running it, bucket updates happen
automatically via CI when you push a new tag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bucket, err := resolveBucket(scoopInitBucket, configPath())
		if err != nil {
			return err
		}

		opts := scoop.InitOpts{
			Bucket:  bucket,
			Tag:     scoopInitTag,
			Repo:    scoopInitRepo,
			Desc:    scoopInitDesc,
			License: scoopInitLicense,
			DryRun:  scoopInitDryRun,
		}

		result, err := scoop.RunInit(opts)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		if scoopInitDryRun {
			header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
			dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

			fmt.Println(header.Render("=== Manifest ==="))
			fmt.Println(result.Manifest)

			fmt.Println(header.Render("=== README ==="))
			fmt.Println(result.Readme)

			fmt.Println(header.Render("=== Workflow Job ==="))
			fmt.Println(result.WorkflowJob)

			fmt.Println(dim.Render("(dry run — nothing was created)"))
			return nil
		}

		ok := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		fmt.Println(ok.Render("Bucket initialized"))
		fmt.Printf("  bucket   %s\n", result.Bucket)
		fmt.Printf("  repo     %s\n", result.Repo)
		fmt.Printf("  tag      %s\n", result.Tag)
		fmt.Printf("  sha256   %s (amd64)\n", dim.Render(result.SHA256AMD64[:12]+"…"))
		fmt.Printf("  sha256   %s (arm64)\n", dim.Render(result.SHA256ARM64[:12]+"…"))

		if result.WorkflowMod {
			fmt.Println(ok.Render("\nWorkflow updated"))
			fmt.Println("  .github/workflows/release.yml now includes update-scoop job")
		}

		fmt.Println(dim.Render("\nReminder: add HOMEBREW_TAP_TOKEN secret to " + result.Repo))
		fmt.Println(dim.Render("  GitHub PAT with Contents write access to " + result.Bucket))

		return nil
	},
}

func init() {
	scoopInitCmd.Flags().StringVar(&scoopInitBucket, "bucket", "", "bucket repo (default: companions.scoop from .jig.yaml)")
	scoopInitCmd.Flags().StringVar(&scoopInitTag, "tag", "", "release tag (default: latest release)")
	scoopInitCmd.Flags().StringVar(&scoopInitRepo, "repo", "", "source repo (default: current repo via gh)")
	scoopInitCmd.Flags().StringVar(&scoopInitDesc, "desc", "", "manifest description (default: repo description)")
	scoopInitCmd.Flags().StringVar(&scoopInitLicense, "license", "", "license identifier (default: from repo)")
	scoopInitCmd.Flags().BoolVar(&scoopInitDryRun, "dry-run", false, "show what would be created without doing it")
	scoopCmd.AddCommand(scoopInitCmd)
}
