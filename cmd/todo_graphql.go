package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"
	"github.com/toba/jig/internal/todo/graph"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	queryJSON       bool
	queryVariables  string
	queryOperation  string
	querySchemaOnly bool
)

var graphqlCmd = &cobra.Command{
	Use:     "graphql <query>",
	Aliases: []string{"query"},
	Short:   "Execute a GraphQL query or mutation",
	Long: `Execute a GraphQL query or mutation against the issues data.

The argument should be a valid GraphQL query or mutation string.

Examples:
  # List all issues
  jig todo graphql '{ issues { id title status } }'

  # Get a specific issue
  jig todo graphql '{ issue(id: "abc") { title status body } }'

  # Use variables
  jig todo graphql -v '{"id": "abc"}' 'query GetIssue($id: ID!) { issue(id: $id) { title } }'

  # Read from stdin
  echo '{ issues { id title } }' | jig todo graphql

  # Print the schema
  jig todo graphql --schema`,
	Args: func(cmd *cobra.Command, args []string) error {
		if querySchemaOnly {
			return nil
		}
		if len(args) > 1 {
			return errors.New("accepts at most 1 argument (the GraphQL query)")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if querySchemaOnly {
			return printSchema()
		}

		var query string
		if len(args) == 1 {
			query = args[0]
		} else {
			stdinQuery, err := readFromStdin()
			if err != nil {
				return err
			}
			if stdinQuery == "" {
				return errors.New("no query provided (pass as argument or pipe to stdin)")
			}
			query = stdinQuery
		}

		var variables map[string]any
		if queryVariables != "" {
			if err := json.Unmarshal([]byte(queryVariables), &variables); err != nil {
				return fmt.Errorf("invalid variables JSON: %w", err)
			}
		}

		result, err := executeQuery(query, variables, queryOperation)
		if err != nil {
			return err
		}

		if queryJSON {
			fmt.Println(string(pretty.Pretty(result)))
		} else {
			fmt.Println(string(pretty.Color(pretty.Pretty(result), nil)))
		}

		return nil
	},
}

func readFromStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("checking stdin: %w", err)
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func executeQuery(query string, variables map[string]any, operationName string) ([]byte, error) {
	es := graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{Core: todoStore},
	})

	exec := executor.New(es)

	ctx := graphql.StartOperationTrace(context.Background())
	params := &graphql.RawParams{
		Query:         query,
		Variables:     variables,
		OperationName: operationName,
	}

	opCtx, errs := exec.CreateOperationContext(ctx, params)
	if errs != nil {
		return nil, formatGraphQLErrors(errs)
	}

	ctx = graphql.WithOperationContext(ctx, opCtx)
	handler, ctx := exec.DispatchOperation(ctx, opCtx)
	resp := handler(ctx)

	if len(resp.Errors) > 0 {
		return nil, formatGraphQLErrors(resp.Errors)
	}

	return resp.Data, nil
}

func formatGraphQLErrors(errs gqlerror.List) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return fmt.Errorf("graphql: %s", errs[0].Message)
	}
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, e.Message)
	}
	return fmt.Errorf("graphql errors:\n  %s", strings.Join(msgs, "\n  "))
}

func printSchema() error {
	fmt.Print(GetGraphQLSchema())
	return nil
}

// GetGraphQLSchema returns the GraphQL schema as a string.
func GetGraphQLSchema() string {
	es := graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{Core: todoStore},
	})

	var buf bytes.Buffer
	f := formatter.NewFormatter(&buf, formatter.WithIndent("  "))
	f.FormatSchema(es.Schema())

	return buf.String()
}

func init() {
	graphqlCmd.Flags().BoolVar(&queryJSON, "json", false, "Output JSON without colors (for piping)")
	graphqlCmd.Flags().StringVarP(&queryVariables, "variables", "v", "", "Query variables as JSON string")
	graphqlCmd.Flags().StringVarP(&queryOperation, "operation", "o", "", "Operation name (for multi-operation documents)")
	graphqlCmd.Flags().BoolVar(&querySchemaOnly, "schema", false, "Print the GraphQL schema and exit")
	todoCmd.AddCommand(graphqlCmd)
}
