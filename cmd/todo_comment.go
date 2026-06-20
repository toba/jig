package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/graph"
	"github.com/toba/jig/internal/todo/graph/model"
	"github.com/toba/jig/internal/todo/issue"
	"github.com/toba/jig/internal/todo/output"
	"github.com/toba/jig/internal/todo/ui"
)

var todoCommentJSON bool

// commentIssue appends text to an issue's body via the same body-modification
// path as `update --append-body`, so etag checks, timestamps, and sync all run.
// It is the implementation behind the discoverable `comment` verb that agents
// reach for by analogy with gh/git, instead of editing `.issues/*.md` directly.
func commentIssue(id, text string) (*issue.Issue, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("comment text is empty")
	}

	ctx := context.Background()
	resolver := &graph.Resolver{Core: todoStore}

	b, err := resolver.Query().Issue(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find issue: %w", err)
	}
	if b == nil {
		return nil, fmt.Errorf("issue not found: %s", id)
	}

	input := model.UpdateIssueInput{BodyMod: &model.BodyModification{Append: &text}}
	return resolver.Mutation().UpdateIssue(ctx, b.ID, input)
}

var todoCommentCmd = &cobra.Command{
	Use:   "comment <id> <text>",
	Short: "Append a note to an issue's body",
	Long: `Appends text to an issue's body, separated from existing content by a blank line.

This is a discoverable alias for 'update --append-body'. Use it instead of
editing the issue's markdown file directly, which bypasses concurrency (etag)
checks, the updated timestamp, and external sync.

Pass '-' as the text to read the comment from stdin (best for multi-line
content with backticks):

  jig todo comment abc - <<'EOF'
  ## Summary
  Fixed the thing.
  EOF`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		var text string
		if len(args) > 1 {
			text = strings.Join(args[1:], " ")
		}
		resolved, err := resolveAppendContent(text)
		if err != nil {
			return cmdError(todoCommentJSON, output.ErrFileError, "%s", err)
		}

		b, err := commentIssue(id, resolved)
		if err != nil {
			return cmdError(todoCommentJSON, output.ErrValidation, "%s", err)
		}

		if todoCommentJSON {
			return output.Success(b, "Comment added")
		}

		fmt.Println(ui.Success.Render("Commented on ") + ui.ID.Render(b.ID) + " " + ui.Muted.Render(b.Path))
		return nil
	},
}

func init() {
	todoCommentCmd.Flags().BoolVar(&todoCommentJSON, "json", false, "Output as JSON")
	todoCmd.AddCommand(todoCommentCmd)
}
