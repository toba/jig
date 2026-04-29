package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/cc"
)

var authStatusFiles = []string{
	".credentials.json",
	".claude.json",
	"policy-limits.json",
	"mcp-needs-auth-cache.json",
	"remote-settings.json",
}

var ccAuthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show per-alias presence of credential files",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cc.Load()
		if err != nil {
			return err
		}
		type result struct {
			Alias string          `json:"alias"`
			Files map[string]bool `json:"files"`
		}
		var out []result
		for _, n := range c.Names() {
			a := c.Aliases[n]
			files := map[string]bool{}
			for _, f := range authStatusFiles {
				p := filepath.Join(a.Path, f)
				info, err := os.Lstat(p)
				files[f] = err == nil && info.Mode()&os.ModeSymlink == 0
			}
			out = append(out, result{Alias: n, Files: files})
		}
		if jsonOut {
			return cc.EmitJSON(cc.JSONResponse{Success: true, Data: out})
		}
		for _, r := range out {
			fmt.Printf("%s\n", r.Alias)
			for _, f := range authStatusFiles {
				mark := "✗"
				if r.Files[f] {
					mark = "✓"
				}
				fmt.Printf("  %s %s\n", mark, f)
			}
		}
		return nil
	},
}

func init() {
	ccAuthCmd.AddCommand(ccAuthStatusCmd)
}
