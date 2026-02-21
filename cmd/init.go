package cmd

import (
	"fmt"
	"os"
	"slices"

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
			if slices.Contains(splitLines(content), "citations:") {
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
