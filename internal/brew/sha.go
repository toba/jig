package brew

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveSHA256 tries three strategies to get the SHA256 for an asset in a release:
//  1. Download <asset>.sha256 sidecar file
//  2. Download checksums.txt and grep for the asset
//  3. Download the asset itself and compute SHA256
func ResolveSHA256(repo, tag, asset string) (string, error) {
	// Strategy 1: .sha256 sidecar
	if sha, err := downloadSidecar(repo, tag, asset); err == nil && sha != "" {
		return sha, nil
	}

	// Strategy 2: checksums.txt
	if sha, err := grepChecksums(repo, tag, asset); err == nil && sha != "" {
		return sha, nil
	}

	// Strategy 3: download and compute
	return computeSHA(repo, tag, asset)
}

func downloadSidecar(repo, tag, asset string) (string, error) {
	out, err := ghRelease("download", tag, "--repo", repo, "--pattern", asset+".sha256", "-O", "-")
	if err != nil {
		return "", err
	}
	// Sidecar files typically contain just the hash, or "hash  filename"
	s := strings.TrimSpace(string(out))
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty sidecar file")
	}
	return fields[0], nil
}

func grepChecksums(repo, tag, asset string) (string, error) {
	out, err := ghRelease("download", tag, "--repo", repo, "--pattern", "checksums.txt", "-O", "-")
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.Contains(line, asset) {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				return fields[0], nil
			}
		}
	}
	return "", fmt.Errorf("asset %s not found in checksums.txt", asset)
}

func computeSHA(repo, tag, asset string) (string, error) {
	tmp, err := os.MkdirTemp("", "brew-sha-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	dest := filepath.Join(tmp, asset)
	_, err = ghRelease("download", tag, "--repo", repo, "--pattern", asset, "-D", tmp)
	if err != nil {
		return "", fmt.Errorf("downloading asset: %w", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		return "", fmt.Errorf("reading downloaded asset: %w", err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum), nil
}

// ghRelease shells out to `gh release`. Separated for testability.
var ghRelease = func(args ...string) ([]byte, error) {
	full := append([]string{"release"}, args...)
	cmd := exec.Command("gh", full...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh release %s: %s", strings.Join(args, " "), string(ee.Stderr))
		}
		return nil, fmt.Errorf("gh release %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}
