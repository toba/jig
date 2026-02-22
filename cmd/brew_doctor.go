package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/brew"
	"github.com/toba/jig/internal/nope"
)

var brewDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify brew tap setup is healthy",
	RunE: func(cmd *cobra.Command, args []string) error {
		tap, err := resolveTap("", configPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "OK:   companions.brew not configured (nothing to check)\n")
			return nil
		}

		// Detect source repo.
		out, err := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner").Output()
		if err != nil {
			return fmt.Errorf("detecting source repo: %w", err)
		}
		repo := strings.TrimSpace(string(out))

		// Derive tool name from tap repo (homebrew-<tool>).
		tapParts := strings.SplitN(tap, "/", 2)
		if len(tapParts) != 2 {
			return fmt.Errorf("unexpected tap format: %s", tap)
		}
		tool := strings.TrimPrefix(tapParts[1], "homebrew-")

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
