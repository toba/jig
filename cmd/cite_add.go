package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/cite"
	"github.com/toba/jig/internal/config"
	"github.com/toba/jig/internal/github"
)

var addWrite bool

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Inspect a repository and suggest citation config",
	Long:  "Inspect a repository via GitHub API or git clone, detect its language, and suggest path classification globs for the citations config.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().BoolVarP(&addWrite, "write", "w", false, "append the suggested source to .jig.yaml")
	citeCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	arg := cite.ParseRepoArg(args[0])
	client := github.NewClient()

	src, err := cite.Inspect(client, arg)
	if err != nil {
		return err
	}

	yamlStr, err := cite.FormatSourceYAML(src)
	if err != nil {
		return err
	}

	if addWrite {
		path := configPath()
		doc, loadErr := config.LoadDocument(path)
		if loadErr != nil {
			// File doesn't exist — create a minimal document.
			if os.IsNotExist(loadErr) {
				if err := os.WriteFile(path, []byte("citations: []\n"), 0o644); err != nil {
					return fmt.Errorf("creating %s: %w", path, err)
				}
				doc, loadErr = config.LoadDocument(path)
			}
			if loadErr != nil {
				return fmt.Errorf("loading config: %w", loadErr)
			}
		}
		if err := config.AppendSource(doc, *src); err != nil {
			return fmt.Errorf("appending source: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Added %s to %s\n", src.Repo, path)
		return nil
	}

	fmt.Print(yamlStr)
	return nil
}
