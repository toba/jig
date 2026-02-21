package cmd

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var helpAllCmd = &cobra.Command{
	Use:   "help-all",
	Short: "Show all commands and flags in agent-friendly format",
	Long:  "Print a compact reference of every command, subcommand, and flag. Designed for consumption by AI agents and scripts.",
	Run: func(cmd *cobra.Command, args []string) {
		printCommandTree(rootCmd, "")
	},
}

func init() {
	rootCmd.AddCommand(helpAllCmd)
}

func printCommandTree(cmd *cobra.Command, prefix string) {
	// Skip hidden commands, completion, and help (agents don't need those).
	if cmd.Hidden || cmd.Name() == "completion" || cmd.Name() == "help-all" || (cmd.Name() == "help" && cmd.Parent() == rootCmd) {
		return
	}

	name := cmd.Name()
	if prefix != "" {
		name = prefix + " " + name
	}

	// Print command line.
	desc := cmd.Short
	if desc == "" {
		desc = cmd.Long
	}
	fmt.Printf("%s — %s\n", name, desc)

	// Print usage if it has args beyond the command name.
	use := cmd.Use
	if _, after, ok := strings.Cut(use, " "); ok {
		argSpec := after
		fmt.Printf("  usage: %s %s\n", name, argSpec)
	}

	// Print local flags (skip help).
	flags := collectFlags(cmd)
	if len(flags) > 0 {
		fmt.Printf("  flags: %s\n", strings.Join(flags, ", "))
	}

	// Recurse into subcommands.
	subs := cmd.Commands()
	slices.SortFunc(subs, func(a, b *cobra.Command) int { return cmp.Compare(a.Name(), b.Name()) })
	for _, sub := range subs {
		printCommandTree(sub, name)
	}
}

func collectFlags(cmd *cobra.Command) []string {
	var flags []string
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		entry := "--" + f.Name
		if f.Shorthand != "" {
			entry = "-" + f.Shorthand + "/" + entry
		}
		if f.Value.Type() != "bool" {
			entry += " <" + f.Value.Type() + ">"
		}
		if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
			entry += " (default: " + f.DefValue + ")"
		}
		flags = append(flags, entry)
	})
	return flags
}
