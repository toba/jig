package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	commitpkg "github.com/toba/skill/internal/commit"
	"github.com/toba/skill/internal/config"
	"github.com/toba/skill/internal/nope"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Two-phase commit workflow: gather context, then apply",
}

var gatherCmd = &cobra.Command{
	Use:   "gather",
	Short: "Stage changes and output context for commit message authoring",
	Long: `Stages all changes (git add -A), checks for gitignore candidates,
then outputs staged files, diff, latest tag, and recent log.

Exit codes:
  0  Success — context printed
  2  Gitignore candidates found — review before committing`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		fmt.Println("STAGED:")
		if status != "" {
			fmt.Println(status)
		}

		// 3. Diff.
		diff, err := commitpkg.Diff()
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("DIFF:")
		if diff != "" {
			fmt.Println(diff)
		}

		// 4. Latest tag.
		tag, err := commitpkg.LatestTag()
		if err != nil {
			return err
		}
		fmt.Println()
		if tag == "" {
			fmt.Println("LATEST_TAG: (none)")
		} else {
			fmt.Println("LATEST_TAG:", tag)
		}

		// 5. Log since tag.
		log, err := commitpkg.LogSinceTag(tag)
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("LOG_SINCE_TAG:")
		if log != "" {
			fmt.Println(log)
		}

		// 6. Todo sync (if configured).
		if hasTodoSync(configPath()) {
			commitpkg.TodoSync()
		}

		return nil
	},
}

var (
	applyMessage string
	applyVersion string
	applyPush    bool
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Commit staged changes with optional tag and push",
	Long: `Creates a git commit from staged changes. Optionally tags a version
and pushes to the remote.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Commit.
		if err := commitpkg.Commit(applyMessage); err != nil {
			return err
		}
		fmt.Println("Committed.")

		// 2. Tag if version provided.
		if applyVersion != "" {
			if err := commitpkg.Tag(applyVersion); err != nil {
				return err
			}
			fmt.Printf("Tagged %s.\n", applyVersion)
		}

		// 3. Push if requested.
		if applyPush {
			if err := commitpkg.Push(); err != nil {
				return err
			}
			fmt.Println("Pushed.")

			// Todo sync after push if configured.
			if hasTodoSync(configPath()) {
				commitpkg.TodoSync()
			}
		}

		// 4. Final status.
		status, err := commitpkg.Status()
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("STATUS:")
		if status != "" {
			fmt.Println(status)
		} else {
			fmt.Println("(clean)")
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().StringVarP(&applyMessage, "message", "m", "", "commit message (required)")
	_ = applyCmd.MarkFlagRequired("message")
	applyCmd.Flags().StringVarP(&applyVersion, "version", "v", "", "version tag to create")
	applyCmd.Flags().BoolVar(&applyPush, "push", false, "push commits and tags after committing")

	commitCmd.AddCommand(gatherCmd)
	commitCmd.AddCommand(applyCmd)
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
