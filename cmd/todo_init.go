package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/output"
)

var todoInitJSON bool

var todoInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a todo project",
	Long:  `Creates a data directory and todo config section in .jig.yaml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectDir string
		var dataDir string

		if todoDataPath != "" {
			dataDir = todoDataPath
			projectDir = filepath.Dir(dataDir)
			c := core.New(dataDir, nil)
			if err := c.Init(); err != nil {
				if todoInitJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			dir, err := os.Getwd()
			if err != nil {
				if todoInitJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return err
			}

			if err := core.Init(dir); err != nil {
				if todoInitJSON {
					return output.Error(output.ErrFileError, err.Error())
				}
				return fmt.Errorf("failed to initialize: %w", err)
			}

			projectDir = dir
			dataDir = filepath.Join(dir, todoconfig.DefaultDataPath)
		}

		// Create default config
		defaultCfg := todoconfig.Default()
		defaultCfg.SetConfigDir(projectDir)
		if err := defaultCfg.Save(projectDir); err != nil {
			if todoInitJSON {
				return output.Error(output.ErrFileError, err.Error())
			}
			return fmt.Errorf("failed to create config: %w", err)
		}

		if todoInitJSON {
			return output.SuccessInit(dataDir)
		}

		fmt.Println("Initialized todo project")
		return nil
	},
}

func init() {
	todoInitCmd.Flags().BoolVar(&todoInitJSON, "json", false, "Output as JSON")
	todoCmd.AddCommand(todoInitCmd)
}
