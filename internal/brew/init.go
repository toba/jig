package brew

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toba/skill/internal/companion"
)

// InitOpts holds the inputs for brew init.
type InitOpts struct {
	Tap     string // e.g. "toba/homebrew-todo"
	Tag     string // e.g. "v1.2.3" (empty = latest)
	Repo    string // e.g. "toba/todo" (empty = detect)
	Desc    string // formula description (empty = detect)
	License string // license identifier (empty = detect)
	DryRun  bool
}

// InitResult describes what was done (or would be done).
type InitResult struct {
	Tap         string `json:"tap"`
	Repo        string `json:"repo"`
	Tool        string `json:"tool"`
	Tag         string `json:"tag"`
	Asset       string `json:"asset"`
	SHA256      string `json:"sha256"`
	Desc        string `json:"desc"`
	License     string `json:"license"`
	Formula     string `json:"formula"`
	Readme      string `json:"readme"`
	WorkflowJob string `json:"workflow_job"`
	TapCreated  bool   `json:"tap_created"`
	TapPushed   bool   `json:"tap_pushed"`
	WorkflowMod bool   `json:"workflow_modified"`
}

// RunInit performs the full brew tap setup workflow.
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

	// Step 3: Derive asset name.
	asset := tool + "_darwin_arm64.tar.gz"

	// Step 4: Resolve SHA256.
	sha, err := ResolveSHA256(repo, tag, asset)
	if err != nil {
		return nil, fmt.Errorf("resolving SHA256 for %s: %w", asset, err)
	}

	// Step 5: Generate formula.
	formulaContent := GenerateFormula(FormulaParams{
		Tool:    tool,
		Desc:    desc,
		Repo:    repo,
		Tag:     tag,
		Asset:   asset,
		SHA256:  sha,
		License: license,
	})

	// Step 6: Generate README.
	readmeContent := generateReadme(tool, org, desc)

	// Step 7: Generate workflow job.
	workflowJob := GenerateWorkflowJob(WorkflowParams{
		Tool:    tool,
		Org:     org,
		Desc:    desc,
		License: license,
		Asset:   asset,
	})

	result := &InitResult{
		Tap:         opts.Tap,
		Repo:        repo,
		Tool:        tool,
		Tag:         tag,
		Asset:       asset,
		SHA256:      sha,
		Desc:        desc,
		License:     license,
		Formula:     formulaContent,
		Readme:      readmeContent,
		WorkflowJob: workflowJob,
	}

	if opts.DryRun {
		return result, nil
	}

	// Step 8: Create tap repo on GitHub.
	if err := createTapRepo(opts.Tap, tool); err != nil {
		return nil, fmt.Errorf("creating tap repo: %w", err)
	}
	result.TapCreated = true

	// Step 9: Push initial content.
	if err := pushInitialContent(opts.Tap, tool, formulaContent, readmeContent); err != nil {
		return nil, fmt.Errorf("pushing initial content: %w", err)
	}
	result.TapPushed = true

	// Step 10: Inject workflow job into release.yml.
	workflowPath := companion.WorkflowPath
	if err := injectWorkflow(workflowPath, WorkflowParams{
		Tool:    tool,
		Org:     org,
		Desc:    desc,
		License: license,
		Asset:   asset,
	}); err != nil {
		// Non-fatal — print warning but don't fail.
		fmt.Fprintf(os.Stderr, "Warning: could not inject workflow job: %v\n", err)
	} else {
		result.WorkflowMod = true
	}

	return result, nil
}


func generateReadme(tool, org, desc string) string {
	return fmt.Sprintf(`# Homebrew Tap for %s

This is the official Homebrew tap for [%s](https://github.com/%s/%s), %s.

## Installation

`+"```bash"+`
brew tap %s/%s
brew install %s
`+"```"+`

## Usage

`+"```bash"+`
%s version
`+"```"+`

## Updating

`+"```bash"+`
brew update
brew upgrade %s
`+"```"+`

## Uninstalling

`+"```bash"+`
brew uninstall %s
brew untap %s/%s
`+"```"+`

## Requirements

- macOS (Apple Silicon)

## Issues

Report issues at [%s/%s](https://github.com/%s/%s/issues).
`, tool, tool, org, tool, desc,
		org, tool, tool,
		tool,
		tool,
		tool, org, tool,
		org, tool, org, tool)
}

func createTapRepo(tap, tool string) error {
	cmd := exec.Command("gh", "repo", "create", tap, "--public",
		"--description", fmt.Sprintf("Homebrew tap for %s", tool))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func pushInitialContent(tap, tool, formula, readme string) error {
	tmp, err := os.MkdirTemp("", "brew-init-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	// Clone the empty repo.
	cmd := exec.Command("gh", "repo", "clone", tap, tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cloning: %s", strings.TrimSpace(string(out)))
	}

	// Write formula.
	formulaDir := filepath.Join(tmp, "Formula")
	if err := os.MkdirAll(formulaDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(formulaDir, tool+".rb"), []byte(formula), 0o644); err != nil {
		return err
	}

	// Write README.
	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte(readme), 0o644); err != nil {
		return err
	}

	// Commit and push.
	cmds := [][]string{
		{"git", "-C", tmp, "add", "."},
		{"git", "-C", tmp, "commit", "-m", "initial formula and README"},
		{"git", "-C", tmp, "push"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
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
