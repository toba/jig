package cite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/toba/jig/internal/config"
	"github.com/toba/jig/internal/github"
)

// DoctorOpts holds the inputs for cite doctor.
type DoctorOpts struct {
	Sources []config.Source
	Client  github.Client // nil if gh is unavailable
}

// licenseFiles are the filenames to search for attribution.
var licenseFiles = []string{
	"LICENSE",
	"LICENSE.md",
	"LICENSE.txt",
	"LICENCE",
	"LICENCE.md",
	"LICENCE.txt",
	"NOTICE",
	"NOTICE.md",
	"NOTICE.txt",
	"THIRD_PARTY",
	"THIRD_PARTY.md",
	"THIRD_PARTY.txt",
	"THIRD-PARTY-NOTICES",
	"THIRD-PARTY-NOTICES.md",
	"THIRD-PARTY-NOTICES.txt",
	"COPYING",
}

// RunDoctor checks that each configured citation has attribution in local
// license/notice files. Returns 0 on success, 1 on any FAIL.
func RunDoctor(opts DoctorOpts) int {
	if len(opts.Sources) == 0 {
		fmt.Fprintf(os.Stderr, "OK:   no citations configured (nothing to check)\n")
		return 0
	}

	// Discover local license files.
	found := discoverLicenseFiles()
	if len(found) == 0 {
		fmt.Fprintf(os.Stderr, "WARN: no LICENSE/NOTICE files found in project root\n")
		fmt.Fprintf(os.Stderr, "      cannot verify citation attribution\n")
		return 0
	}
	fmt.Fprintf(os.Stderr, "OK:   found license files: %s\n", strings.Join(fileNames(found), ", "))

	// Read all license file contents once.
	var contents []string
	for _, f := range found {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: could not read %s: %v\n", f, err)
			continue
		}
		contents = append(contents, strings.ToLower(string(data)))
	}

	ok := true
	for _, src := range opts.Sources {
		ok = checkAttribution(src, contents, opts.Client) && ok
	}

	if !ok {
		return 1
	}
	return 0
}

// checkAttribution verifies a single citation is mentioned in license files.
func checkAttribution(src config.Source, contents []string, client github.Client) bool {
	// Build search terms from the repo slug.
	owner, repo := splitRepo(src.Repo)
	terms := []string{strings.ToLower(src.Repo)}
	if repo != "" {
		terms = append(terms, strings.ToLower(repo))
	}
	if owner != "" {
		terms = append(terms, strings.ToLower(owner))
	}

	// Try to get upstream license info from GitHub.
	var licenseName string
	if client != nil && isGitHubRepo(src.Repo) {
		info, err := client.GetLicense(src.Repo)
		if err == nil && info.License.Name != "" {
			licenseName = info.License.Name
		}
	}

	// Search for any mention in local license files.
	for _, term := range terms {
		for _, content := range contents {
			if strings.Contains(content, term) {
				msg := fmt.Sprintf("OK:   %s — found attribution", src.Repo)
				if licenseName != "" {
					msg += fmt.Sprintf(" (upstream: %s)", licenseName)
				}
				fmt.Fprintln(os.Stderr, msg)
				return true
			}
		}
	}

	// Not found — report as failure.
	msg := fmt.Sprintf("FAIL: %s — no attribution found in license files", src.Repo)
	if licenseName != "" {
		msg += fmt.Sprintf(" (upstream license: %s)", licenseName)
	}
	fmt.Fprintln(os.Stderr, msg)
	return false
}

// discoverLicenseFiles returns paths to license-related files in the
// current working directory.
func discoverLicenseFiles() []string {
	var found []string
	for _, name := range licenseFiles {
		matches, _ := filepath.Glob(name)
		found = append(found, matches...)
	}
	// Deduplicate (glob shouldn't produce dupes, but be safe).
	seen := make(map[string]bool, len(found))
	deduped := found[:0]
	for _, f := range found {
		if !seen[f] {
			seen[f] = true
			deduped = append(deduped, f)
		}
	}
	return deduped
}

// splitRepo splits "owner/repo" into its parts.
func splitRepo(slug string) (owner, repo string) {
	parts := strings.SplitN(slug, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", slug
}

// isGitHubRepo returns true if the repo slug looks like a GitHub repo
// (owner/name format without other URL components).
func isGitHubRepo(repo string) bool {
	parts := strings.SplitN(repo, "/", 3)
	// GitHub slugs are exactly "owner/repo" with no further slashes.
	return len(parts) == 2 && !strings.Contains(repo, "://")
}

// fileNames extracts just the base names from paths.
func fileNames(paths []string) []string {
	names := make([]string, len(paths))
	for i, p := range paths {
		names[i] = filepath.Base(p)
	}
	return names
}
