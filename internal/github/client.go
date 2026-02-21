package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Client defines the interface for fetching data from GitHub.
type Client interface {
	GetCommits(repo, branch string, perPage int) ([]Commit, error)
	GetCommitsSince(repo, branch, since string, perPage int) ([]Commit, error)
	Compare(repo, base, head string) (*CompareResponse, error)
	GetCommitDetail(repo, sha string) (*Commit, error)
	GetHeadSHA(repo, branch string) (string, error)
	GetRepo(repo string) (*RepoInfo, error)
	GetTree(repo, branch string) (*TreeResponse, error)
	GetLicense(repo string) (*LicenseInfo, error)
}

// GHClient implements Client by shelling out to the gh CLI.
type GHClient struct{}

func NewClient() *GHClient {
	return &GHClient{}
}

func (c *GHClient) GetCommits(repo, branch string, perPage int) ([]Commit, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/commits?per_page=%d&sha=%s", repo, perPage, branch))
	if err != nil {
		return nil, err
	}

	var commits []Commit
	if err := json.Unmarshal(out, &commits); err != nil {
		return nil, fmt.Errorf("parsing commits: %w", err)
	}
	for i := range commits {
		commits[i].Normalize()
	}
	return commits, nil
}

func (c *GHClient) GetCommitsSince(repo, branch, since string, perPage int) ([]Commit, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/commits?since=%s&sha=%s&per_page=%d", repo, since, branch, perPage))
	if err != nil {
		return nil, err
	}

	var commits []Commit
	if err := json.Unmarshal(out, &commits); err != nil {
		return nil, fmt.Errorf("parsing commits: %w", err)
	}
	for i := range commits {
		commits[i].Normalize()
	}
	return commits, nil
}

func (c *GHClient) Compare(repo, base, head string) (*CompareResponse, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/compare/%s...%s", repo, base, head))
	if err != nil {
		return nil, err
	}

	var resp CompareResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parsing compare response: %w", err)
	}
	for i := range resp.Commits {
		resp.Commits[i].Normalize()
	}
	return &resp, nil
}

func (c *GHClient) GetCommitDetail(repo, sha string) (*Commit, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/commits/%s", repo, sha))
	if err != nil {
		return nil, err
	}

	var commit Commit
	if err := json.Unmarshal(out, &commit); err != nil {
		return nil, fmt.Errorf("parsing commit detail: %w", err)
	}
	commit.Normalize()
	return &commit, nil
}

func (c *GHClient) GetHeadSHA(repo, branch string) (string, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/commits/%s", repo, branch), "--jq", ".sha")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *GHClient) GetRepo(repo string) (*RepoInfo, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s", repo))
	if err != nil {
		return nil, err
	}
	var info RepoInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parsing repo info: %w", err)
	}
	return &info, nil
}

func (c *GHClient) GetTree(repo, branch string) (*TreeResponse, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/git/trees/%s?recursive=1", repo, branch))
	if err != nil {
		return nil, err
	}
	var resp TreeResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parsing tree response: %w", err)
	}
	return &resp, nil
}

func (c *GHClient) GetLicense(repo string) (*LicenseInfo, error) {
	out, err := gh("api", fmt.Sprintf("repos/%s/license", repo))
	if err != nil {
		return nil, err
	}
	var info LicenseInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parsing license info: %w", err)
	}
	return &info, nil
}

func gh(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh %s: %s", strings.Join(args, " "), string(ee.Stderr))
		}
		return nil, fmt.Errorf("gh %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}
