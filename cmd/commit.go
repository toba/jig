package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	commitpkg "github.com/toba/skill/internal/commit"
	"github.com/toba/skill/internal/config"
	"github.com/toba/skill/internal/nope"
)

var commitCmd = &cobra.Command{
	Use:   "commit [push]",
	Short: "Stage changes and check for gitignore candidates",
	Long: `Checks untracked files against known gitignore patterns, stages all
changes, optionally triggers todo sync, and signals push intent.

Exit codes:
  0  Success — all changes staged
  2  Gitignore candidates found — review before committing`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		push := len(args) > 0 && args[0] == "push"

		// 1. Check for gitignore candidates.
		candidates, err := commitpkg.GitignoreCandidates()
		if err != nil {
			return err
		}
		if len(candidates) > 0 {
			fmt.Println("GITIGNORE_CANDIDATES:")
			for _, c := range candidates {
				fmt.Println(c)
			}
			fmt.Println()
			fmt.Println("These untracked files may belong in .gitignore.")
			return nope.ExitError{Code: 2}
		}

		// 2. Stage all changes.
		status, err := commitpkg.StageAll()
		if err != nil {
			return err
		}
		fmt.Println("Staged changes:")
		if status != "" {
			fmt.Println(status)
		}

		// 3. Todo sync (if configured).
		if hasTodoSync(configPath()) {
			commitpkg.TodoSync()
		}

		// 4. Push signal.
		if push {
			fmt.Println()
			fmt.Println("PUSH_AFTER_COMMIT")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

// hasTodoSync checks whether .toba.yaml has a todo.sync section.
func hasTodoSync(path string) bool {
	doc, err := config.LoadDocument(path)
	if err != nil {
		return false
	}
	todoNode := config.FindKey(doc.Root, "todo")
	if todoNode == nil {
		return false
	}
	return config.FindKey(todoNode, "sync") != nil
}
