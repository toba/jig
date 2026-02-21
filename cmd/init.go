package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const starterConfig = `citations:
  - repo: owner/repo
    branch: main
    notes: ""
    paths:
      high:
        - "src/**/*.go"
      medium:
        - "go.mod"
        - "go.sum"
      low:
        - ".github/**"
        - "README.md"
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Add starter citations section to .jig.yaml",
	Long:  "Create or update .jig.yaml with a starter citations configuration section.",
	RunE:  runInit,
}

func init() {
	citeCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	path := configPath()

	// Check if file already exists.
	if data, err := os.ReadFile(path); err == nil {
		// File exists — check if it already has a citations section.
		content := string(data)
		if len(content) > 0 {
			// Simple check for existing citations section.
			if slices.Contains(strings.Split(content, "\n"), "citations:") {
				return fmt.Errorf("%s already contains a 'citations' section", path)
			}
			// Append citations section to existing file.
			if content[len(content)-1] != '\n' {
				content += "\n"
			}
			content += "\n" + starterConfig
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("updating %s: %w", path, err)
			}
			fmt.Printf("added citations section to %s\n", path)
			return nil
		}
	}

	// Create new file.
	if err := os.WriteFile(path, []byte(starterConfig), 0o644); err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	fmt.Printf("created %s\n", path)
	return nil
}
