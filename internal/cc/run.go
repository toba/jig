package cc

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Launch resolves an alias and executes its CLI with the given pass-through
// arguments. CLAUDE_CONFIG_DIR is set for non-source aliases. Stdio is
// inherited; this call blocks until the child exits and returns its exit
// code (0 on success, otherwise wraps the error).
func Launch(c *Config, query string, extraArgs []string) (int, error) {
	name, a, err := c.Resolve(query)
	if err != nil {
		return 1, err
	}

	// For non-source aliases, ensure symlinks exist before launching so the
	// CLI sees a complete config dir.
	if !a.IsSource {
		if _, err := Sync(c, name); err != nil {
			return 1, fmt.Errorf("syncing %s: %w", name, err)
		}
	}

	bin, err := exec.LookPath(a.CLI)
	if err != nil {
		return 1, fmt.Errorf("cli %q not found in PATH: %w", a.CLI, err)
	}
	cmd := exec.Command(bin, extraArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	if !a.IsSource {
		cmd.Env = append(cmd.Env, "CLAUDE_CONFIG_DIR="+a.Path)
	}

	// Record history (best-effort).
	if cwd, err := os.Getwd(); err == nil {
		_ = RecordHistory(cwd, name)
	}

	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if asExitError(err, &ee) {
			return ee.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

func asExitError(err error, target **exec.ExitError) bool {
	for err != nil {
		ee := &exec.ExitError{}
		if errors.As(err, &ee) {
			*target = ee
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
