package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
	github "github.com/toba/jig/internal/todo/integration/github"
)

var tagsImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import GitHub labels as project tags",
	Long:  `Imports labels from the configured GitHub repository into the project tag registry in .jig.yaml.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cp := configPath()

		// Load config (we skip initTodoCore for "import")
		var err error
		if _, statErr := os.Stat(cp); statErr == nil {
			todoCfg, err = todoconfig.Load(cp)
			if err != nil {
				return fmt.Errorf("loading config from %s: %w", cp, err)
			}
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}
			todoCfg, err = todoconfig.LoadFromDirectory(cwd)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
		}

		// Parse GitHub sync config
		ghCfg, err := github.ParseConfig(todoCfg.SyncConfig("github"))
		if err != nil {
			return fmt.Errorf("parsing GitHub sync config: %w", err)
		}
		if ghCfg == nil {
			return fmt.Errorf("no GitHub sync configured in .jig.yaml (need sync.github.repo)")
		}

		// Get token
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN environment variable is required")
		}

		// Fetch labels
		client := github.NewClient(token, ghCfg.Owner, ghCfg.Repo)
		labels, err := client.ListLabels(context.Background())
		if err != nil {
			return fmt.Errorf("fetching labels: %w", err)
		}

		replace, _ := cmd.Flags().GetBool("replace")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		var imported, updated int

		if replace {
			todoCfg.Tags = nil
		}

		for _, label := range labels {
			name := strings.TrimSpace(label.Name)
			if name == "" {
				continue
			}

			existing := todoCfg.GetTag(name)
			if existing != nil {
				// Update description if it changed
				if existing.Description != label.Description {
					existing.Description = label.Description
					updated++
				}
			} else {
				todoCfg.Tags = append(todoCfg.Tags, todoconfig.TagConfig{
					Name:        name,
					Description: label.Description,
				})
				imported++
			}
		}

		if err := todoCfg.Save(""); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		if jsonOutput {
			result := map[string]any{
				"imported": imported,
				"updated":  updated,
				"total":    len(todoCfg.Tags),
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		fmt.Printf("Imported %d new tags, updated %d existing (%d total)\n", imported, updated, len(todoCfg.Tags))
		return nil
	},
}

func init() {
	tagsImportCmd.Flags().Bool("replace", false, "Clear existing tags before importing")
	tagsImportCmd.Flags().Bool("json", false, "Output results as JSON")
	tagsCmd.AddCommand(tagsImportCmd)
}
