package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/brew"
	"github.com/toba/skill/internal/nope"
)

var brewDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify brew tap setup is healthy",
	RunE: func(cmd *cobra.Command, args []string) error {
		tap, err := resolveTap("", configPath())
		if err != nil {
			return err
		}

		// Detect source repo and tool name.
		out, err := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner").Output()
		if err != nil {
			return fmt.Errorf("detecting source repo: %w", err)
		}
		repo := strings.TrimSpace(string(out))
		parts := strings.SplitN(repo, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("unexpected repo format: %s", repo)
		}
		tool := parts[1]

		code := brew.RunDoctor(brew.DoctorOpts{
			Tap:  tap,
			Repo: repo,
			Tool: tool,
		})
		if code != 0 {
			return nope.ExitError{Code: code}
		}
		return nil
	},
}

func init() {
	brewCmd.AddCommand(brewDoctorCmd)
}
