package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/brew"
)

var (
	brewInitTap     string
	brewInitTag     string
	brewInitRepo    string
	brewInitDesc    string
	brewInitLicense string
	brewInitDryRun  bool
)

var brewInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up a new Homebrew tap with automated releases",
	Long: `Creates a companion Homebrew tap repo, pushes an initial formula and README,
and injects an update-homebrew job into the source repo's release workflow.

This is a one-time setup command. After running it, tap updates happen
automatically via CI when you push a new tag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tap, err := resolveTap(brewInitTap, configPath())
		if err != nil {
			return err
		}

		opts := brew.InitOpts{
			Tap:     tap,
			Tag:     brewInitTag,
			Repo:    brewInitRepo,
			Desc:    brewInitDesc,
			License: brewInitLicense,
			DryRun:  brewInitDryRun,
		}

		result, err := brew.RunInit(opts)
		if err != nil {
			return err
		}

		if jsonOut {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		if brewInitDryRun {
			header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
			dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

			fmt.Println(header.Render("=== Formula ==="))
			fmt.Println(result.Formula)

			fmt.Println(header.Render("=== README ==="))
			fmt.Println(result.Readme)

			fmt.Println(header.Render("=== Workflow Job ==="))
			fmt.Println(result.WorkflowJob)

			fmt.Println(dim.Render("(dry run — nothing was created)"))
			return nil
		}

		ok := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

		fmt.Println(ok.Render("Tap initialized"))
		fmt.Printf("  tap      %s\n", result.Tap)
		fmt.Printf("  repo     %s\n", result.Repo)
		fmt.Printf("  tag      %s\n", result.Tag)
		fmt.Printf("  sha256   %s\n", dim.Render(result.SHA256[:12]+"…"))

		if result.WorkflowMod {
			fmt.Println(ok.Render("\nWorkflow updated"))
			fmt.Println("  .github/workflows/release.yml now includes update-homebrew job")
		}

		fmt.Println(dim.Render("\nReminder: add HOMEBREW_TAP_TOKEN secret to " + result.Repo))
		fmt.Println(dim.Render("  GitHub PAT with Contents write access to " + result.Tap))

		return nil
	},
}

func init() {
	brewInitCmd.Flags().StringVar(&brewInitTap, "tap", "", "tap repo (default: companions.brew from .toba.yaml)")
	brewInitCmd.Flags().StringVar(&brewInitTag, "tag", "", "release tag (default: latest release)")
	brewInitCmd.Flags().StringVar(&brewInitRepo, "repo", "", "source repo (default: current repo via gh)")
	brewInitCmd.Flags().StringVar(&brewInitDesc, "desc", "", "formula description (default: repo description)")
	brewInitCmd.Flags().StringVar(&brewInitLicense, "license", "", "license identifier (default: from repo)")
	brewInitCmd.Flags().BoolVar(&brewInitDryRun, "dry-run", false, "show what would be created without doing it")
	brewCmd.AddCommand(brewInitCmd)
}
