package commit

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitignoreCandidates returns untracked files that match gitignore patterns.
func GitignoreCandidates() ([]string, error) {
	out, err := exec.Command("git", "ls-files", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil, fmt.Errorf("listing untracked files: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}

	var candidates []string
	for _, file := range strings.Split(raw, "\n") {
		if matchesGitignorePattern(file) {
			candidates = append(candidates, file)
		}
	}
	return candidates, nil
}

// matchesGitignorePattern reports whether a file path matches any built-in
// gitignore candidate pattern.
func matchesGitignorePattern(path string) bool {
	for _, re := range gitignorePatterns {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// StageAll runs git add -A and returns the short status output.
func StageAll() (string, error) {
	if err := exec.Command("git", "add", "-A").Run(); err != nil {
		return "", fmt.Errorf("git add -A: %w", err)
	}
	out, err := exec.Command("git", "status", "--short").Output()
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// TodoSync runs "todo sync" in the background, ignoring errors.
// Returns immediately.
func TodoSync() {
	cmd := exec.Command("todo", "sync")
	// Fire and forget — don't wait for completion, ignore errors.
	_ = cmd.Start()
}
