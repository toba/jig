package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/toba/jig/internal/todo/integration"
)

var syncDoctorCmd = &cobra.Command{
	Use:    "doctor",
	Hidden: true,
	Short:  "Check sync configuration health",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for .jig.yml typo in current directory
		if _, err := os.Stat(".jig.yml"); err == nil {
			return errors.New("found .jig.yml but jig expects .jig.yaml — please rename it")
		}

		// If no sync config, that's fine — sync is optional
		if todoCfg == nil || todoCfg.Sync == nil {
			return nil
		}

		integ, err := integration.Detect(todoCfg.Sync, todoStore)
		if err != nil {
			return fmt.Errorf("detecting integration: %w", err)
		}
		if integ == nil {
			return nil
		}

		// Run offline check only (skip API calls for doctor)
		ctx := context.Background()
		report, err := integ.Check(ctx, integration.CheckOptions{SkipAPI: true})
		if err != nil {
			return err
		}
		if report.Summary.Failed > 0 {
			return fmt.Errorf("%d sync check(s) failed", report.Summary.Failed)
		}
		return nil
	},
}
