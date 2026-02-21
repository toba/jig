package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/config"
)

var brewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Homebrew tap management",
	Long:  "Commands for managing Homebrew tap formulas.",
}

func init() {
	rootCmd.AddCommand(brewCmd)
}

// tapFromCompanions reads the companions.brew git URL from .toba.yaml
// and extracts the "owner/repo" tap identifier. Returns "" on any failure.
func tapFromCompanions(cfgPath string) string {
	doc, err := config.LoadDocument(cfgPath)
	if err != nil {
		return ""
	}
	c := config.LoadCompanions(doc)
	if c == nil || c.Brew == "" {
		return ""
	}
	return repoFromGitURL(c.Brew)
}

// resolveTap determines the tap repo using (in order):
//  1. explicit --tap flag
//  2. companions.brew from .toba.yaml
//  3. convention: owner/homebrew-<name> derived from the current GitHub repo
func resolveTap(flag, cfgPath string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if tap := tapFromCompanions(cfgPath); tap != "" {
		return tap, nil
	}
	if tap := tapFromConvention(); tap != "" {
		return tap, nil
	}
	return "", fmt.Errorf("--tap is required (or set companions.brew in %s)", cfgPath)
}

// tapFromConvention derives "owner/homebrew-name" from the current GitHub repo.
func tapFromConvention() string {
	out, err := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner").Output()
	if err != nil {
		return ""
	}
	repo := strings.TrimSpace(string(out))
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0] + "/homebrew-" + parts[1]
}

// repoFromGitURL extracts "owner/repo" from a git URL.
// Handles https://github.com/owner/repo.git and git@github.com:owner/repo.git.
func repoFromGitURL(u string) string {
	// Strip trailing .git
	u = strings.TrimSuffix(u, ".git")

	// HTTPS: https://github.com/owner/repo
	if strings.Contains(u, "://") {
		parts := strings.Split(u, "/")
		if len(parts) >= 2 {
			return parts[len(parts)-2] + "/" + parts[len(parts)-1]
		}
		return ""
	}

	// SSH: git@github.com:owner/repo
	if _, after, ok := strings.Cut(u, ":"); ok {
		return after
	}

	return ""
}
