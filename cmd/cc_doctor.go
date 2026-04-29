package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/cc"
	"github.com/toba/jig/internal/nope"
)

var ccDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check cc symlink health for every alias",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cc.Load()
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if jsonOut {
					return cc.EmitJSON(cc.JSONResponse{Success: true, Message: "no cc config; nothing to check"})
				}
				fmt.Fprintln(os.Stderr, "OK:   no cc config (nothing to check)")
				return nil
			}
			return err
		}
		results := map[string]*cc.Health{}
		bad := 0
		for _, n := range c.Names() {
			if c.Aliases[n].IsSource {
				continue
			}
			h, err := cc.CheckHealth(c, n)
			if err != nil {
				return err
			}
			results[n] = h
			if h.HasIssues() {
				bad++
			}
		}
		if jsonOut {
			return cc.EmitJSON(cc.JSONResponse{Success: bad == 0, Data: results})
		}
		for n, h := range results {
			fmt.Printf("%s: %d valid, %d broken, %d missing, %d conflicts, %d orphaned\n",
				n, len(h.Valid), len(h.Broken), len(h.Missing), len(h.Conflicts), len(h.Orphaned))
			for _, x := range h.Broken {
				fmt.Printf("    broken: %s\n", x)
			}
			for _, x := range h.Missing {
				fmt.Printf("    missing: %s\n", x)
			}
			for _, x := range h.Conflicts {
				fmt.Printf("    conflict: %s\n", x)
			}
			for _, x := range h.Orphaned {
				fmt.Printf("    orphaned: %s\n", x)
			}
		}
		if bad > 0 {
			return nope.ExitError{Code: 1}
		}
		return nil
	},
}

func init() {
	ccCmd.AddCommand(ccDoctorCmd)
}
