package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	todoconfig "github.com/toba/jig/internal/todo/config"
	"github.com/toba/jig/internal/todo/core"
	"github.com/toba/jig/internal/todo/ui"
)

var (
	todoCheckJSON bool
	todoCheckFix  bool
)

type todoCheckResult struct {
	Success      bool                  `json:"success"`
	ConfigErrors []string              `json:"config_errors"`
	LinkIssues   *core.LinkCheckResult `json:"link_issues,omitempty"`
	Fixed        int                   `json:"fixed,omitempty"`
}

var todoCheckCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate configuration and issue integrity",
	Long: `Checks configuration and issue integrity, including:
- Configuration settings (colors, default type)
- Sync integration configuration (unknown or multiple integrations)
- Broken links (links to non-existent issues)
- Self-references (issues linking to themselves)
- Circular dependencies (cycles in blocks/parent relationships)

Use --fix to automatically remove broken links and self-references.
Note: Cycles cannot be auto-fixed and require manual intervention.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var configErrors []string
		var fixed int

		// === Configuration checks ===
		if !todoCheckJSON {
			fmt.Println(ui.Bold.Render("Configuration"))
		}

		// 1. Check statuses are defined (always true since hardcoded)
		if !todoCheckJSON {
			fmt.Printf("  %s Statuses defined (%d hardcoded)\n", ui.Success.Render("✓"), len(todoconfig.DefaultStatuses))
		}

		// 2. Check default_status exists in statuses (always true since hardcoded)
		if !todoCheckJSON {
			fmt.Printf("  %s Default status '%s' exists\n", ui.Success.Render("✓"), todoCfg.GetDefaultStatus())
		}

		// 2b. Check default_type is a valid hardcoded type
		if todoCfg.GetDefaultType() != "" && !todoCfg.IsValidType(todoCfg.GetDefaultType()) {
			configErrors = append(configErrors, fmt.Sprintf("default_type '%s' is not a valid type", todoCfg.GetDefaultType()))
		} else if todoCfg.GetDefaultType() != "" {
			if !todoCheckJSON {
				fmt.Printf("  %s Default type '%s' is valid\n", ui.Success.Render("✓"), todoCfg.GetDefaultType())
			}
		}

		// 3. Check all status colors are valid (hardcoded statuses)
		for _, s := range todoconfig.DefaultStatuses {
			if !ui.IsValidColor(s.Color) {
				configErrors = append(configErrors, fmt.Sprintf("invalid color '%s' for status '%s'", s.Color, s.Name))
			}
		}
		if !todoCheckJSON {
			colorErrors := 0
			for _, e := range configErrors {
				if len(e) > 13 && e[:13] == "invalid color" {
					colorErrors++
				}
			}
			if colorErrors == 0 {
				fmt.Printf("  %s All status colors valid\n", ui.Success.Render("✓"))
			}
		}

		// 4. Check all type colors are valid (hardcoded types)
		for _, t := range todoconfig.DefaultTypes {
			if !ui.IsValidColor(t.Color) {
				configErrors = append(configErrors, fmt.Sprintf("invalid color '%s' for type '%s'", t.Color, t.Name))
			}
		}
		if !todoCheckJSON {
			typeColorErrors := 0
			for _, e := range configErrors {
				if len(e) > 13 && e[:13] == "invalid color" {
					typeColorErrors++
				}
			}
			if typeColorErrors == 0 {
				fmt.Printf("  %s All type colors valid\n", ui.Success.Render("✓"))
			}
		}

		// 5. Check sync configuration
		knownIntegrations := map[string]bool{"clickup": true, "github": true}
		if len(todoCfg.Sync) > 0 {
			var configuredIntegrations []string
			for name := range todoCfg.Sync {
				if !knownIntegrations[name] {
					configErrors = append(configErrors, fmt.Sprintf("unknown sync integration '%s' (known: clickup, github)", name))
				} else {
					configuredIntegrations = append(configuredIntegrations, name)
				}
			}
			if len(configuredIntegrations) > 1 {
				slices.Sort(configuredIntegrations)
				configErrors = append(configErrors, fmt.Sprintf("multiple sync integrations configured (%s); only one is supported at a time", strings.Join(configuredIntegrations, ", ")))
			} else if len(configuredIntegrations) == 1 && !todoCheckJSON {
				fmt.Printf("  %s Sync integration '%s' configured\n", ui.Success.Render("✓"), configuredIntegrations[0])
			}
		}

		// Print config errors in human-readable mode
		if !todoCheckJSON {
			for _, e := range configErrors {
				fmt.Printf("  %s %s\n", ui.Danger.Render("✗"), e)
			}
		}

		// === Issue link checks ===
		if !todoCheckJSON {
			fmt.Println()
			fmt.Println(ui.Bold.Render("Issue Links"))
		}

		linkResult := todoStore.CheckAllLinks()

		// Handle --fix mode
		if todoCheckFix && (len(linkResult.BrokenLinks) > 0 || len(linkResult.SelfLinks) > 0) {
			fixedCount, err := todoStore.FixBrokenLinks()
			if err != nil {
				return fmt.Errorf("fixing broken links: %w", err)
			}
			fixed = fixedCount

			if !todoCheckJSON {
				for _, bl := range linkResult.BrokenLinks {
					fmt.Printf("  %s %s: removed broken link %s:%s\n", ui.Success.Render("✓"), bl.IssueID, bl.LinkType, bl.Target)
				}
				for _, sl := range linkResult.SelfLinks {
					fmt.Printf("  %s %s: removed self-reference in %s link\n", ui.Success.Render("✓"), sl.IssueID, sl.LinkType)
				}
			}

			// Clear the fixed issues from the result
			linkResult.BrokenLinks = []core.BrokenLink{}
			linkResult.SelfLinks = []core.SelfLink{}
		} else if !todoCheckJSON {
			// Report issues without fixing
			for _, bl := range linkResult.BrokenLinks {
				fmt.Printf("  %s %s: broken link %s:%s\n", ui.Danger.Render("✗"), bl.IssueID, bl.LinkType, bl.Target)
			}
			for _, sl := range linkResult.SelfLinks {
				fmt.Printf("  %s %s: self-reference in %s link\n", ui.Danger.Render("✗"), sl.IssueID, sl.LinkType)
			}
		}

		// Cycles cannot be auto-fixed
		if !todoCheckJSON {
			for _, c := range linkResult.Cycles {
				if todoCheckFix {
					fmt.Printf("  %s Cannot auto-fix cycle: %s (via %s)\n", ui.Warning.Render("!"), formatCycle(c.Path), c.LinkType)
				} else {
					fmt.Printf("  %s Circular dependency: %s (via %s)\n", ui.Danger.Render("✗"), formatCycle(c.Path), c.LinkType)
				}
			}
		}

		// Show success if no issues
		if !todoCheckJSON && !linkResult.HasIssues() && fixed == 0 {
			fmt.Printf("  %s No link issues found\n", ui.Success.Render("✓"))
		}

		// === Summary ===
		totalIssues := len(configErrors) + linkResult.TotalIssues()

		if todoCheckJSON {
			result := todoCheckResult{
				Success:      totalIssues == 0,
				ConfigErrors: configErrors,
				LinkIssues:   linkResult,
				Fixed:        fixed,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println()
			if totalIssues == 0 && fixed == 0 {
				fmt.Println(ui.Success.Render("All checks passed"))
			} else if totalIssues == 0 && fixed > 0 {
				fmt.Println(ui.Success.Render(fmt.Sprintf("Fixed %d issue(s)", fixed)))
			} else if fixed > 0 {
				fmt.Println(ui.Warning.Render(fmt.Sprintf("Fixed %d issue(s), %d require manual intervention", fixed, totalIssues)))
			} else if totalIssues == 1 {
				fmt.Println(ui.Danger.Render("1 issue found"))
			} else {
				fmt.Println(ui.Danger.Render(fmt.Sprintf("%d issues found", totalIssues)))
			}
		}

		// Exit with error code if validation failed
		if totalIssues > 0 {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	todoCheckCmd.Flags().BoolVar(&todoCheckJSON, "json", false, "Output as JSON")
	todoCheckCmd.Flags().BoolVar(&todoCheckFix, "fix", false, "Automatically fix broken links and self-references")
	todoCmd.AddCommand(todoCheckCmd)
}
