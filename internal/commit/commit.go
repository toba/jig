package commit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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
	for file := range strings.SplitSeq(raw, "\n") {
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

// RecentCommits returns recent commits for style reference.
// If a tag is provided and there are commits since it, returns those.
// Otherwise returns the last 20 commits so agents always have style context.
func RecentCommits(tag string) (string, error) {
	if tag != "" {
		cmd := exec.Command("git", "log", tag+"..HEAD", "--format=%h %s") //nolint:gosec // args from internal config
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("git log: %w", err)
		}
		if result := strings.TrimRight(string(out), "\n"); result != "" {
			return result, nil
		}
	}
	// No tag or no commits since tag — show recent commits for style reference.
	out, err := exec.Command("git", "log", "--format=%h %s", "-20").Output()
	if err != nil {
		return "", fmt.Errorf("git log: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// HasStagedChanges reports whether there are any staged changes to commit.
func HasStagedChanges() (bool, error) {
	err := exec.Command("git", "diff", "--cached", "--quiet").Run()
	if err == nil {
		return false, nil // exit 0 = no differences
	}
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		return true, nil // exit 1 = differences exist
	}
	return false, fmt.Errorf("git diff --cached --quiet: %w", err)
}

// Commit creates a git commit with the given message.
// Stderr is captured and included in the error so hook failures are visible.
func Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message) //nolint:gosec // args from internal config
	cmd.WaitDelay = 10 * time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		if detail != "" {
			return fmt.Errorf("git commit: %w\n%s", err, detail)
		}
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Tag creates a git tag with the given name.
func Tag(version string) error {
	if err := exec.Command("git", "tag", version).Run(); err != nil { //nolint:gosec // args from internal config
		return fmt.Errorf("git tag %s: %w", version, err)
	}
	return nil
}

// Push pushes the current branch, then pushes any unpushed version tags
// individually in semver order. This prevents out-of-order GitHub releases
// when multiple tags are pushed simultaneously.
func Push() error {
	if err := exec.Command("git", "push").Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	tags, err := unpushedVersionTags()
	if err != nil {
		// Can't determine unpushed tags (e.g. no remote) — push all tags.
		if err := exec.Command("git", "push", "--tags").Run(); err != nil {
			return fmt.Errorf("git push --tags: %w", err)
		}
		return nil
	}

	for _, tag := range tags {
		if err := exec.Command("git", "push", "origin", tag).Run(); err != nil { //nolint:gosec // args from internal config
			return fmt.Errorf("git push origin %s: %w", tag, err)
		}
	}
	return nil
}

// unpushedVersionTags returns local v* tags not present on the remote,
// sorted in version order (oldest first).
func unpushedVersionTags() ([]string, error) {
	// Get local v* tags sorted by version.
	localOut, err := exec.Command("git", "tag", "-l", "v*", "--sort=version:refname").Output()
	if err != nil {
		return nil, fmt.Errorf("listing local tags: %w", err)
	}

	// Get remote tags.
	remoteOut, err := exec.Command("git", "ls-remote", "--tags", "origin").Output()
	if err != nil {
		return nil, fmt.Errorf("listing remote tags: %w", err)
	}

	remote := make(map[string]bool)
	for line := range strings.SplitSeq(strings.TrimSpace(string(remoteOut)), "\n") {
		// Format: "<sha>\trefs/tags/<name>" (skip ^{} derefs)
		if _, ref, ok := strings.Cut(line, "\t"); ok {
			tag := strings.TrimPrefix(ref, "refs/tags/")
			if !strings.HasSuffix(tag, "^{}") {
				remote[tag] = true
			}
		}
	}

	var unpushed []string
	for tag := range strings.SplitSeq(strings.TrimSpace(string(localOut)), "\n") {
		if tag != "" && !remote[tag] {
			unpushed = append(unpushed, tag)
		}
	}
	return unpushed, nil
}

// RestageIssues stages any modified .issues/ files so sync metadata
// changes are included in the upcoming commit. No-op if the directory
// doesn't exist or has no changes.
func RestageIssues() error {
	if !issuesDirExists() {
		return nil
	}
	return exec.Command("git", "add", "--", ".issues").Run()
}

func issuesDirExists() bool {
	_, err := os.Stat(".issues")
	return err == nil
}

// Status returns the current git status output.
func Status() (string, error) {
	out, err := exec.Command("git", "status", "--short").Output()
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
