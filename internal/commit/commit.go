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

// Diff returns the staged diff output.
func Diff() (string, error) {
	out, err := exec.Command("git", "diff", "--staged").Output()
	if err != nil {
		return "", fmt.Errorf("git diff --staged: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// LatestTag returns the latest v* tag by version sort, or "" if none exist.
func LatestTag() (string, error) {
	out, err := exec.Command("git", "tag", "-l", "v*", "--sort=version:refname").Output()
	if err != nil {
		return "", fmt.Errorf("git tag -l: %w", err)
	}
	lines := strings.TrimSpace(string(out))
	if lines == "" {
		return "", nil
	}
	parts := strings.Split(lines, "\n")
	return parts[len(parts)-1], nil
}

// LogSinceTag returns the oneline log of commits since the given tag.
// If tag is empty, returns the last 10 commits.
func LogSinceTag(tag string) (string, error) {
	var cmd *exec.Cmd
	if tag == "" {
		cmd = exec.Command("git", "log", "--oneline", "-10")
	} else {
		cmd = exec.Command("git", "log", tag+"..HEAD", "--oneline")
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// Commit creates a git commit with the given message.
func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Tag creates a git tag with the given name.
func Tag(version string) error {
	if err := exec.Command("git", "tag", version).Run(); err != nil {
		return fmt.Errorf("git tag %s: %w", version, err)
	}
	return nil
}

// Push runs git push && git push --tags.
func Push() error {
	if err := exec.Command("git", "push").Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	if err := exec.Command("git", "push", "--tags").Run(); err != nil {
		return fmt.Errorf("git push --tags: %w", err)
	}
	return nil
}

// Status returns the current git status output.
func Status() (string, error) {
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
