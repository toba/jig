package scoop

import (
	"cmp"
	"fmt"

	"github.com/toba/jig/internal/companion"
)

// WorkflowParams holds the inputs needed to generate the update-scoop CI job.
type WorkflowParams struct {
	Tool    string // binary name, e.g. "jig"
	Org     string // GitHub org, e.g. "toba"
	Desc    string // one-line description
	License string // e.g. "Apache-2.0"
	Needs   string // job this depends on, e.g. "release"
}

// GenerateWorkflowJob produces the YAML block for an update-scoop job.
func GenerateWorkflowJob(p WorkflowParams) string {
	needs := cmp.Or(p.Needs, "release")
	return fmt.Sprintf(`
  update-scoop:
    needs: %[1]s
    runs-on: ubuntu-latest
    steps:
      - name: Update Scoop manifest
        env:
          GH_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
        run: |
          TAG="${GITHUB_REF#refs/tags/}"
          VERSION="${TAG#v}"

          CHECKSUMS=$(gh release download "$TAG" --repo "$GITHUB_REPOSITORY" --pattern checksums.txt -O -)

          SHA_AMD64=$(echo "$CHECKSUMS" | grep "%[2]s_windows_amd64.zip" | awk '{print $1}')
          SHA_ARM64=$(echo "$CHECKSUMS" | grep "%[2]s_windows_arm64.zip" | awk '{print $1}')

          if [ -z "$SHA_AMD64" ]; then
            echo "ERROR: Could not extract SHA256 for %[2]s_windows_amd64.zip"
            exit 1
          fi
          if [ -z "$SHA_ARM64" ]; then
            echo "ERROR: Could not extract SHA256 for %[2]s_windows_arm64.zip"
            exit 1
          fi

          git clone "https://x-access-token:${GH_TOKEN}@github.com/%[3]s/scoop-%[2]s.git" bucket
          cd bucket

          jq -n \
            --arg version "$VERSION" \
            --arg sha_amd64 "$SHA_AMD64" \
            --arg sha_arm64 "$SHA_ARM64" \
            --arg tag "$TAG" \
            --arg desc "%[4]s" \
            '{
              version: $version,
              description: $desc,
              homepage: "https://github.com/%[3]s/%[2]s",
              license: "%[5]s",
              architecture: {
                "64bit": {
                  url: ("https://github.com/%[3]s/%[2]s/releases/download/" + $tag + "/%[2]s_windows_amd64.zip"),
                  hash: $sha_amd64
                },
                arm64: {
                  url: ("https://github.com/%[3]s/%[2]s/releases/download/" + $tag + "/%[2]s_windows_arm64.zip"),
                  hash: $sha_arm64
                }
              },
              bin: ["%[2]s.exe"],
              checkver: { github: "https://github.com/%[3]s/%[2]s" },
              autoupdate: {
                architecture: {
                  "64bit": {
                    url: "https://github.com/%[3]s/%[2]s/releases/download/v$version/%[2]s_windows_amd64.zip"
                  },
                  arm64: {
                    url: "https://github.com/%[3]s/%[2]s/releases/download/v$version/%[2]s_windows_arm64.zip"
                  }
                }
              }
            }' > bucket/%[2]s.json

          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add bucket/%[2]s.json
          git commit -m "bump to ${VERSION}"
          git push
`, needs, p.Tool, p.Org, p.Desc, p.License)
}

// InjectWorkflowJob appends the update-scoop job to an existing workflow file.
// It returns the modified content or an error if the job already exists.
func InjectWorkflowJob(content string, p WorkflowParams) (string, error) {
	return companion.InjectJob(content, "update-scoop:", &p.Needs, func() string {
		return GenerateWorkflowJob(p)
	})
}
