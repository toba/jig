package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/config"
)

var scoopCmd = &cobra.Command{
	Use:   "scoop",
	Short: "Scoop bucket management",
	Long:  "Commands for managing Scoop bucket manifests.",
}

func init() {
	rootCmd.AddCommand(scoopCmd)
}

// bucketFromCompanions reads the companions.scoop git URL from .jig.yaml
// and extracts the "owner/repo" bucket identifier. Returns "" on any failure.
func bucketFromCompanions(cfgPath string) string {
	doc, err := config.LoadDocument(cfgPath)
	if err != nil {
		return ""
	}
	c := config.LoadCompanions(doc)
	if c == nil || c.Scoop == "" {
		return ""
	}
	return repoFromGitURL(c.Scoop)
}

// resolveBucket determines the bucket repo using (in order):
//  1. explicit --bucket flag
//  2. companions.scoop from .jig.yaml
//  3. convention: owner/scoop-<name> derived from the current GitHub repo
func resolveBucket(flag, cfgPath string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if bucket := bucketFromCompanions(cfgPath); bucket != "" {
		return bucket, nil
	}
	if bucket := bucketFromConvention(); bucket != "" {
		return bucket, nil
	}
	return "", fmt.Errorf("--bucket is required (or set companions.scoop in %s)", cfgPath)
}

// bucketFromConvention derives "owner/scoop-name" from the current GitHub repo.
func bucketFromConvention() string {
	out, err := exec.Command("gh", "repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner").Output()
	if err != nil {
		return ""
	}
	repo := strings.TrimSpace(string(out))
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[0] + "/scoop-" + parts[1]
}
