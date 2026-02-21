package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

const starterConfig = `upstream:
  sources:
    - repo: owner/repo
      branch: main
      relationship: derived    # derived | dependency | watch
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
	Short: "Add starter upstream section to .toba.yaml",
	Long:  "Create or update .toba.yaml with a starter upstream configuration section.",
	RunE:  runInit,
}

func init() {
	upstreamCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	path := configPath()

	// Check if file already exists.
	if data, err := os.ReadFile(path); err == nil {
		// File exists — check if it already has an upstream section.
		content := string(data)
		if len(content) > 0 {
			// Simple check for existing upstream section.
			if slices.Contains(splitLines(content), "upstream:") {
				return fmt.Errorf("%s already contains an 'upstream' section", path)
			}
			// Append upstream section to existing file.
			if content[len(content)-1] != '\n' {
				content += "\n"
			}
			content += "\n" + starterConfig
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return fmt.Errorf("updating %s: %w", path, err)
			}
			fmt.Printf("added upstream section to %s\n", path)
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

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := range len(s) {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
