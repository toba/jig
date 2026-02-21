package brew

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/toba/jig/internal/companion"
	"gopkg.in/yaml.v3"
)

// DoctorOpts holds the inputs for brew doctor.
type DoctorOpts struct {
	Tap  string // e.g. "toba/homebrew-skill"
	Repo string // e.g. "toba/skill"
	Tool string // e.g. "skill"
}

// RunDoctor validates the brew tap chain is healthy.
// Returns 0 on success, 1 on any FAIL.
func RunDoctor(opts DoctorOpts) int {
	ok := true

	// 1. companions.brew configured
	if opts.Tap == "" {
		fmt.Fprintf(os.Stderr, "FAIL: companions.brew not configured in .jig.yaml\n")
		return 1
	}
	fmt.Fprintf(os.Stderr, "OK:   companions.brew configured: %s\n", opts.Tap)

	lang := DetectLanguage()
	fmt.Fprintf(os.Stderr, "OK:   detected language: %s\n", lang.Name)

	// 2. .goreleaser.yaml exists and is configured correctly (Go only)
	if lang.HasBuildToolCheck() {
		ok = checkGoreleaser(opts.Tool) && ok
	}

	// 3. tap repo exists on GitHub
	cmd := exec.Command("gh", "repo", "view", opts.Tap)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: tap repo %s not found on GitHub: %s\n", opts.Tap, strings.TrimSpace(string(out)))
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "OK:   tap repo exists: %s\n", opts.Tap)
	}

	// 4. formula exists in tap
	formulaPath := fmt.Sprintf("repos/%s/contents/Formula/%s.rb", opts.Tap, opts.Tool)
	cmd = exec.Command("gh", "api", formulaPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: formula not found at Formula/%s.rb in %s: %s\n", opts.Tool, opts.Tap, strings.TrimSpace(string(out)))
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "OK:   formula exists: Formula/%s.rb\n", opts.Tool)
	}

	// 5. source repo has releases
	tag := ""
	cmd = exec.Command("gh", "release", "list", "--repo", opts.Repo, "--limit", "1", "--json", "tagName", "--jq", ".[0].tagName")
	if out, err := cmd.Output(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: no releases found for %s\n", opts.Repo)
		ok = false
	} else {
		tag = strings.TrimSpace(string(out))
		if tag == "" {
			fmt.Fprintf(os.Stderr, "FAIL: no releases found for %s\n", opts.Repo)
			ok = false
		} else {
			fmt.Fprintf(os.Stderr, "OK:   latest release: %s\n", tag)
		}
	}

	// 6. latest release has darwin arm64 asset
	expectedAsset := lang.AssetName(opts.Tool, tag)
	if tag != "" {
		cmd = exec.Command("gh", "release", "view", tag, "--repo", opts.Repo, "--json", "assets")
		if out, err := cmd.Output(); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: could not fetch release %s assets\n", tag)
			ok = false
		} else {
			var release struct {
				Assets []struct {
					Name string `json:"name"`
				} `json:"assets"`
			}
			if err := json.Unmarshal(out, &release); err != nil {
				fmt.Fprintf(os.Stderr, "FAIL: could not parse release assets: %v\n", err)
				ok = false
			} else {
				foundAsset := false
				hasChecksums := false
				hasSidecar := false
				sidecarName := expectedAsset + ".sha256"
				for _, a := range release.Assets {
					if a.Name == expectedAsset {
						foundAsset = true
					}
					if a.Name == "checksums.txt" {
						hasChecksums = true
					}
					if a.Name == sidecarName {
						hasSidecar = true
					}
				}
				if !foundAsset {
					fmt.Fprintf(os.Stderr, "FAIL: release %s missing asset %s\n", tag, expectedAsset)
					ok = false
				} else {
					fmt.Fprintf(os.Stderr, "OK:   release has asset: %s\n", expectedAsset)
				}
				// 7. checksum verification
				switch lang.ChecksumMode() {
				case "checksums.txt":
					if !hasChecksums {
						fmt.Fprintf(os.Stderr, "FAIL: release %s missing checksums.txt (goreleaser checksum output)\n", tag)
						ok = false
					} else {
						fmt.Fprintf(os.Stderr, "OK:   release has checksums.txt\n")
					}
				case "sidecar":
					if !hasSidecar {
						fmt.Fprintf(os.Stderr, "FAIL: release %s missing %s\n", tag, sidecarName)
						ok = false
					} else {
						fmt.Fprintf(os.Stderr, "OK:   release has %s\n", sidecarName)
					}
				}
			}
		}
	}

	// 8. release workflow exists locally
	workflowPath := companion.WorkflowPath
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %s not found\n", workflowPath)
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "OK:   workflow exists: %s\n", workflowPath)

		workflowStr := string(content)

		// 9. workflow uses expected build tool
		buildFound := false
		for _, marker := range lang.WorkflowBuildMarkers() {
			if strings.Contains(workflowStr, marker) {
				buildFound = true
				break
			}
		}
		if !buildFound {
			fmt.Fprintf(os.Stderr, "FAIL: workflow missing %s\n", lang.WorkflowBuildLabel())
			ok = false
		} else {
			fmt.Fprintf(os.Stderr, "OK:   workflow uses %s\n", lang.WorkflowBuildLabel())
		}

		// 10. workflow has update-homebrew job
		if !strings.Contains(workflowStr, "update-homebrew:") {
			fmt.Fprintf(os.Stderr, "FAIL: workflow missing update-homebrew job\n")
			ok = false
		} else {
			fmt.Fprintf(os.Stderr, "OK:   workflow has update-homebrew job\n")
		}

		// 11. workflow references correct tap repo
		tapRepo := filepath.Base(opts.Tap) // e.g. "homebrew-skill"
		if !strings.Contains(workflowStr, tapRepo) {
			fmt.Fprintf(os.Stderr, "FAIL: workflow does not reference %s\n", tapRepo)
			ok = false
		} else {
			fmt.Fprintf(os.Stderr, "OK:   workflow references %s\n", tapRepo)
		}

		// 12. workflow references expected asset name
		if !strings.Contains(workflowStr, expectedAsset) {
			fmt.Fprintf(os.Stderr, "FAIL: workflow does not reference asset %s\n", expectedAsset)
			ok = false
		} else {
			fmt.Fprintf(os.Stderr, "OK:   workflow references asset %s\n", expectedAsset)
		}
	}

	if !ok {
		return 1
	}
	return 0
}

// goreleaserConfig is the subset of .goreleaser.yaml we validate.
type goreleaserConfig struct {
	Builds []struct {
		Binary string   `yaml:"binary"`
		Goos   []string `yaml:"goos"`
		Goarch []string `yaml:"goarch"`
	} `yaml:"builds"`
	Archives []struct {
		Formats      []string `yaml:"formats"`
		NameTemplate string   `yaml:"name_template"`
	} `yaml:"archives"`
	Checksum struct {
		NameTemplate string `yaml:"name_template"`
	} `yaml:"checksum"`
}

// checkGoreleaser validates .goreleaser.yaml is present and correctly
// configured for brew distribution. Returns true if all checks pass.
func checkGoreleaser(tool string) bool {
	ok := true

	data, err := os.ReadFile(".goreleaser.yaml")
	if err != nil {
		// Try alternate extension.
		data, err = os.ReadFile(".goreleaser.yml")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: .goreleaser.yaml not found\n")
		return false
	}
	fmt.Fprintf(os.Stderr, "OK:   .goreleaser.yaml exists\n")

	var cfg goreleaserConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: .goreleaser.yaml parse error: %v\n", err)
		return false
	}

	// Check builds include darwin + arm64.
	hasDarwin, hasArm64 := false, false
	for _, b := range cfg.Builds {
		for _, os := range b.Goos {
			if os == "darwin" {
				hasDarwin = true
			}
		}
		for _, arch := range b.Goarch {
			if arch == "arm64" {
				hasArm64 = true
			}
		}
	}
	if !hasDarwin {
		fmt.Fprintf(os.Stderr, "FAIL: goreleaser builds missing goos: darwin\n")
		ok = false
	} else if !hasArm64 {
		fmt.Fprintf(os.Stderr, "FAIL: goreleaser builds missing goarch: arm64\n")
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "OK:   goreleaser builds darwin/arm64\n")
	}

	// Check archive produces tar.gz with expected naming.
	if len(cfg.Archives) == 0 {
		fmt.Fprintf(os.Stderr, "WARN: goreleaser has no archives section (using defaults)\n")
	} else {
		a := cfg.Archives[0]
		hasTarGz := false
		for _, f := range a.Formats {
			if f == "tar.gz" {
				hasTarGz = true
			}
		}
		if !hasTarGz {
			fmt.Fprintf(os.Stderr, "FAIL: goreleaser archive format does not include tar.gz\n")
			ok = false
		} else {
			fmt.Fprintf(os.Stderr, "OK:   goreleaser archive format includes tar.gz\n")
		}

		// The name template must produce {tool}_{os}_{arch}.
		// The standard template is "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}".
		tmpl := a.NameTemplate
		if tmpl != "" && !strings.Contains(tmpl, ".Os") {
			fmt.Fprintf(os.Stderr, "WARN: goreleaser name_template %q may not produce %s_darwin_arm64 assets\n", tmpl, tool)
		} else if tmpl != "" {
			fmt.Fprintf(os.Stderr, "OK:   goreleaser name_template: %s\n", tmpl)
		}
	}

	// Check checksums.txt generation is enabled.
	if cfg.Checksum.NameTemplate == "" {
		fmt.Fprintf(os.Stderr, "WARN: goreleaser checksum name_template not set (defaults to checksums.txt)\n")
	} else if cfg.Checksum.NameTemplate != "checksums.txt" {
		fmt.Fprintf(os.Stderr, "FAIL: goreleaser checksum name_template is %q, expected checksums.txt\n", cfg.Checksum.NameTemplate)
		ok = false
	} else {
		fmt.Fprintf(os.Stderr, "OK:   goreleaser checksum: checksums.txt\n")
	}

	return ok
}
