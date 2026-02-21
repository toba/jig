package zed

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/toba/skill/internal/companion"
)

// WorkflowParams holds the inputs needed to generate the sync-extension CI job.
type WorkflowParams struct {
	Org    string // GitHub org, e.g. "toba"
	Ext    string // extension repo name, e.g. "gozer"
	Needs  string // job this depends on, e.g. "release"
}

// GenerateSyncExtensionJob produces the YAML block for a sync-extension job.
func GenerateSyncExtensionJob(p WorkflowParams) string {
	needs := cmp.Or(p.Needs, "release")
	return fmt.Sprintf(`
  sync-extension:
    runs-on: ubuntu-latest
    needs: %s
    steps:
      - name: Dispatch version bump to %s/%s
        run: |
          gh api repos/%s/%s/dispatches \
            -f event_type=bump-version \
            -f 'client_payload[version]=${{ github.ref_name }}'
        env:
          GH_TOKEN: ${{ secrets.EXTENSION_PAT }}
`, needs, p.Org, p.Ext, p.Org, p.Ext)
}

// InjectSyncExtensionJob appends the sync-extension job to an existing workflow file.
// It returns the modified content or an error if the job already exists.
func InjectSyncExtensionJob(content string, p WorkflowParams) (string, error) {
	if strings.Contains(content, "sync-extension:") {
		return "", fmt.Errorf("workflow already contains a sync-extension job")
	}

	// Detect the "needs" job name from the existing workflow.
	if p.Needs == "" {
		p.Needs = companion.DetectLastJob(content)
	}

	job := GenerateSyncExtensionJob(p)

	// Ensure the file ends with a newline before appending.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + job, nil
}

