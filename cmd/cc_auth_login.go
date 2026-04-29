package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/cc"
	"github.com/toba/jig/internal/nope"
)

var ccAuthLoginCmd = &cobra.Command{
	Use:   "login <alias>",
	Short: "Run `<cli> /login` for the given alias",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := cc.Load()
		if err != nil {
			return err
		}
		code, err := cc.Launch(c, args[0], []string{"/login"})
		if err != nil {
			return err
		}
		if code != 0 {
			return nope.ExitError{Code: code}
		}
		return nil
	},
}

func init() {
	ccAuthCmd.AddCommand(ccAuthLoginCmd)
}
