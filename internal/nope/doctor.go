package nope

import (
	"fmt"
	"os"
)

// RunDoctor validates the nope configuration.
func RunDoctor() int {
	ok := true

	// Check config file exists
	path, err := FindConfigPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "OK:   config found: %s\n", path)

	// Check config parses (includes nope: section check)
	cfg, err := LoadConfig(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "OK:   nope section parses successfully\n")
	}

	// Check rules compile
	if cfg != nil {
		if len(cfg.Rules) == 0 {
			fmt.Fprintf(os.Stderr, "WARN: no rules defined\n")
		} else {
			_, err := CompileRules(cfg.Rules)
			if err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
				ok = false
			} else {
				fmt.Fprintf(os.Stderr, "OK:   %d rule(s) compile successfully\n", len(cfg.Rules))
			}
		}
	}

	if !ok {
		return 1
	}
	return 0
}
