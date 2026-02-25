package scoop

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toba/jig/internal/brew"
	"github.com/toba/jig/internal/companion"
)

// InitOpts holds the inputs for scoop init.
type InitOpts struct {
	Bucket  string // e.g. "toba/scoop-jig"
	Tag     string // e.g. "v1.2.3" (empty = latest)
	Repo    string // e.g. "toba/jig" (empty = detect)
	Desc    string // manifest description (empty = detect)
	License string // license identifier (empty = detect)
	DryRun  bool
}

// InitResult describes what was done (or would be done).
type InitResult struct {
	Bucket      string `json:"bucket"`
	Repo        string `json:"repo"`
	Tool        string `json:"tool"`
	Tag         string `json:"tag"`
	AssetAMD64  string `json:"asset_amd64"`
	AssetARM64  string `json:"asset_arm64"`
	SHA256AMD64 string `json:"sha256_amd64"`
	SHA256ARM64 string `json:"sha256_arm64"`
	Desc        string `json:"desc"`
	License     string `json:"license"`
	Manifest    string `json:"manifest"`
	Readme      string `json:"readme"`
	WorkflowJob string `json:"workflow_job"`
	BucketCreated bool `json:"bucket_created"`
	BucketPushed  bool `json:"bucket_pushed"`
	WorkflowMod   bool `json:"workflow_modified"`
}

// RunInit performs the full scoop bucket setup workflow.
func RunInit(opts InitOpts) (*InitResult, error) {
	// Step 1: Detect source repo info.
	info, err := companion.DetectRepoInfo(opts.Repo, "nameWithOwner,description,licenseInfo")
	if err != nil {
		return nil, fmt.Errorf("detecting repo info: %w", err)
	}

	repo := info.NameWithOwner
	desc := cmp.Or(opts.Desc, info.Description)
	license := cmp.Or(opts.License, info.LicenseInfo.SpdxID, "NOASSERTION")

	// Derive tool name and org from repo.
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo format %q, expected owner/name", repo)
	}
	org := parts[0]
	tool := parts[1]

	// Step 2: Detect latest release if tag not given.
	tag := opts.Tag
	if tag == "" {
		tag, err = companion.DetectLatestTag(repo)
		if err != nil {
			return nil, fmt.Errorf("detecting latest release: %w", err)
		}
	}

	// Step 3: Derive asset names.
	assetAMD64 := tool + "_windows_amd64.zip"
	assetARM64 := tool + "_windows_arm64.zip"

	// Step 4: Resolve SHA256 for both architectures.
	shaAMD64, err := brew.ResolveSHA256(repo, tag, assetAMD64)
	if err != nil {
		return nil, fmt.Errorf("resolving SHA256 for %s: %w", assetAMD64, err)
	}
	shaARM64, err := brew.ResolveSHA256(repo, tag, assetARM64)
	if err != nil {
		return nil, fmt.Errorf("resolving SHA256 for %s: %w", assetARM64, err)
	}

	// Step 5: Generate manifest.
	manifestContent := GenerateManifest(ManifestParams{
		Tool:        tool,
		Desc:        desc,
		Homepage:    "https://github.com/" + repo,
		License:     license,
		Tag:         tag,
		Repo:        repo,
		SHA256AMD64: shaAMD64,
		SHA256ARM64: shaARM64,
	})

	// Step 6: Generate README.
	readmeContent := generateReadme(tool, org, desc)

	// Step 7: Generate workflow job.
	workflowJob := GenerateWorkflowJob(WorkflowParams{
		Tool:    tool,
		Org:     org,
		Desc:    desc,
		License: license,
	})

	result := &InitResult{
		Bucket:      opts.Bucket,
		Repo:        repo,
		Tool:        tool,
		Tag:         tag,
		AssetAMD64:  assetAMD64,
		AssetARM64:  assetARM64,
		SHA256AMD64: shaAMD64,
		SHA256ARM64: shaARM64,
		Desc:        desc,
		License:     license,
		Manifest:    manifestContent,
		Readme:      readmeContent,
		WorkflowJob: workflowJob,
	}

	if opts.DryRun {
		return result, nil
	}

	// Step 8: Create bucket repo on GitHub.
	if err := createBucketRepo(opts.Bucket, tool); err != nil {
		return nil, fmt.Errorf("creating bucket repo: %w", err)
	}
	result.BucketCreated = true

	// Step 9: Push initial content.
	if err := pushInitialContent(opts.Bucket, tool, manifestContent, readmeContent); err != nil {
		return nil, fmt.Errorf("pushing initial content: %w", err)
	}
	result.BucketPushed = true

	// Step 10: Inject workflow job into release.yml.
	workflowPath := companion.WorkflowPath
	if err := injectWorkflow(workflowPath, WorkflowParams{
		Tool:    tool,
		Org:     org,
		Desc:    desc,
		License: license,
	}); err != nil {
		// Non-fatal — print warning but don't fail.
		fmt.Fprintf(os.Stderr, "Warning: could not inject workflow job: %v\n", err)
	} else {
		result.WorkflowMod = true
	}

	return result, nil
}

func generateReadme(tool, org, desc string) string {
	return fmt.Sprintf(`# Scoop Bucket for %s

This is the official Scoop bucket for [%s](https://github.com/%s/%s), %s.

## Installation

`+"`"+`powershell
scoop bucket add %s https://github.com/%s/scoop-%s
scoop install %s
`+"`"+`

## Usage

`+"`"+`powershell
%s version
`+"`"+`

## Updating

`+"`"+`powershell
scoop update %s
`+"`"+`

## Uninstalling

`+"`"+`powershell
scoop uninstall %s
scoop bucket rm %s
`+"`"+`

## Requirements

- Windows (amd64 or arm64)

## Issues

Report issues at [%s/%s](https://github.com/%s/%s/issues).
`, tool, tool, org, tool, desc,
		org, org, tool, tool,
		tool,
		tool,
		tool, org,
		org, tool, org, tool)
}

func createBucketRepo(bucket, tool string) error {
	cmd := exec.Command("gh", "repo", "create", bucket, "--public", //nolint:gosec // gh CLI wrapper
		"--description", "Scoop bucket for "+tool)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func pushInitialContent(bucket, tool, manifest, readme string) error {
	tmp, err := os.MkdirTemp("", "scoop-init-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmp) //nolint:errcheck // best-effort cleanup

	// Clone the empty repo.
	cmd := exec.Command("gh", "repo", "clone", bucket, tmp) //nolint:gosec // gh CLI wrapper
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cloning: %s", strings.TrimSpace(string(out)))
	}

	// Write manifest.
	bucketDir := filepath.Join(tmp, "bucket")
	if err := os.MkdirAll(bucketDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(bucketDir, tool+".json"), []byte(manifest), 0o644); err != nil {
		return err
	}

	// Write README.
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte(readme), 0o644); err != nil {
		return err
	}

	// Commit and push.
	cmds := [][]string{
		{"git", "-C", tmp, "add", "."},
		{"git", "-C", tmp, "commit", "-m", "initial manifest and README"},
		{"git", "-C", tmp, "push"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...) //nolint:gosec // gh CLI wrapper
		if out, err := c.CombinedOutput(); err != nil {
			return fmt.Errorf("%s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func injectWorkflow(path string, p WorkflowParams) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	modified, err := InjectWorkflowJob(string(data), p)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(modified), 0o644)
}
