package cmd

import (
	_ "embed"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
)

//go:embed todo_prompt.tmpl
var agentPromptTemplate string

// promptData holds all data needed to render the prompt template.
type promptData struct {
	Types      []todoconfig.TypeConfig
	Statuses   []todoconfig.StatusConfig
	Priorities []todoconfig.PriorityConfig
	Tags       []todoconfig.TagConfig
	HasSync    bool
	SyncNames  []string
}

var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Output instructions for AI coding agents",
	Long:  `Outputs a prompt that primes AI coding agents on how to use the issues CLI to manage project issues.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var primeCfg *todoconfig.Config
		if todoDataPath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil
			}
			configFile, err := todoconfig.FindConfig(cwd)
			if err != nil || configFile == "" {
				return nil
			}
			primeCfg, _ = todoconfig.Load(configFile)
		} else {
			cp := configPath()
			primeCfg, _ = todoconfig.Load(cp)
		}

		tmpl, err := template.New("prompt").Parse(agentPromptTemplate)
		if err != nil {
			return err
		}

		data := promptData{
			Types:      todoconfig.DefaultTypes,
			Statuses:   todoconfig.DefaultStatuses,
			Priorities: todoconfig.DefaultPriorities,
		}

		if primeCfg != nil && len(primeCfg.Tags) > 0 {
			data.Tags = primeCfg.Tags
		}

		if primeCfg != nil && primeCfg.Sync != nil {
			data.HasSync = true
			for name := range primeCfg.Sync {
				data.SyncNames = append(data.SyncNames, name)
			}
		}

		return tmpl.Execute(os.Stdout, data)
	},
}

func init() {
	rootCmd.AddCommand(primeCmd)
}
