package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
)

var (
	todoStore    *core.Core
	todoCfg      *todoconfig.Config
	todoDataPath string
)

// loadConfigWithFallback loads todo config from the given path, falling back
// to searching upward from the current directory.
func loadConfigWithFallback(cfgPath string) (*todoconfig.Config, error) {
	if _, statErr := os.Stat(cfgPath); statErr == nil {
		cfg, err := todoconfig.Load(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("loading config from %s: %w", cfgPath, err)
		}
		return cfg, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %w", err)
	}
	cfg, err := todoconfig.LoadFromDirectory(cwd)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return cfg, nil
}

// initTodoCore loads config, resolves data dir, and creates the core.
// Extracted from todo's rootCmd.PersistentPreRunE.
func initTodoCore(_ *cobra.Command) error {
	var err error

	todoCfg, err = loadConfigWithFallback(configPath())
	if err != nil {
		return err
	}

	// Determine data directory
	var root string
	if todoDataPath != "" {
		root = todoDataPath
		if info, statErr := os.Stat(root); statErr != nil || !info.IsDir() {
			return fmt.Errorf("data path does not exist or is not a directory: %s", root)
		}
	} else {
		root = todoCfg.ResolveDataPath()
		if info, statErr := os.Stat(root); statErr != nil || !info.IsDir() {
			return fmt.Errorf("no data directory found at %s (run 'jig todo init' to create one)", root)
		}
	}

	todoStore = core.New(root, todoCfg)
	if err := todoStore.Load(); err != nil {
		return fmt.Errorf("loading issues: %w", err)
	}

	return nil
}

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "File-based issue tracker for AI-first workflows",
	Long: `Todo is a lightweight issue tracker that stores issues as markdown files.
Track your work alongside your code and supercharge your coding agent with
a full view of your project.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip core initialization for init, prime, and refry commands
		if cmd.Name() == "init" || cmd.Name() == "prime" || cmd.Name() == "refry" || cmd.Name() == "import" {
			return nil
		}
		return initTodoCore(cmd)
	},
}

func init() {
	todoCmd.PersistentFlags().StringVar(&todoDataPath, "data-path", "", "Path to data directory (overrides config)")
	rootCmd.AddCommand(todoCmd)
}
