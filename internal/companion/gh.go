// Package companion provides shared helpers for brew and zed companion setup.
package companion

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// RepoInfo holds common fields from gh repo view.
type RepoInfo struct {
	NameWithOwner string `json:"nameWithOwner"`
	Description   string `json:"description"`
	LicenseInfo   struct {
		SpdxID string `json:"spdxId"`
	} `json:"licenseInfo"`
}

// DetectRepoInfo fetches repo metadata via gh. If repo is empty, it detects
// from the current directory. The fields parameter controls which JSON fields
// are requested (e.g. "nameWithOwner,description,licenseInfo").
func DetectRepoInfo(repo, fields string) (*RepoInfo, error) {
	args := []string{"repo", "view", "--json", fields}
	if repo != "" {
		args = append(args, repo)
	}
	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh repo view: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, err
	}
	var info RepoInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parsing repo info: %w", err)
	}
	return &info, nil
}

// DetectLatestTag returns the tag name of the most recent release for repo.
func DetectLatestTag(repo string) (string, error) {
	cmd := exec.Command("gh", "release", "list", "--repo", repo, "--limit", "1", "--json", "tagName", "--jq", ".[0].tagName")
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh release list: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	tag := strings.TrimSpace(string(out))
	if tag == "" {
		return "", fmt.Errorf("no releases found for %s", repo)
	}
	return tag, nil
}
