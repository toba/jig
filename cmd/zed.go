package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/toba/skill/internal/config"
)

var zedCmd = &cobra.Command{
	Use:   "zed",
	Short: "Zed extension management",
	Long:  "Commands for managing Zed editor extension companion repos.",
}

func init() {
	rootCmd.AddCommand(zedCmd)
}

// extFromCompanions reads the companions.zed git URL from .toba.yaml
// and extracts the "owner/repo" identifier. Returns "" on any failure.
func extFromCompanions(cfgPath string) string {
	doc, err := config.LoadDocument(cfgPath)
	if err != nil {
		return ""
	}
	c := config.LoadCompanions(doc)
	if c == nil || c.Zed == "" {
		return ""
	}
	return repoFromGitURL(c.Zed)
}

// resolveExt determines the extension repo using (in order):
//  1. explicit --ext flag
//  2. companions.zed from .toba.yaml
func resolveExt(flag, cfgPath string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if ext := extFromCompanions(cfgPath); ext != "" {
		return ext, nil
	}
	return "", nil
}

// extNameFromRepo extracts just the repo name from "owner/repo".
func extNameFromRepo(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return repo
	}
	return parts[1]
}
