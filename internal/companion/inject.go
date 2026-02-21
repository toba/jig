package companion

import (
	"fmt"
	"strings"
)

// InjectJob appends a CI job to an existing workflow string. It checks whether
// the job already exists using jobMarker (e.g. "update-homebrew:"), optionally
// detects the last job name to fill in the needs field via needsPtr, and calls
// generate to produce the YAML block to append.
func InjectJob(content, jobMarker string, needsPtr *string, generate func() string) (string, error) {
	if strings.Contains(content, jobMarker) {
		return "", fmt.Errorf("workflow already contains a %s job", strings.TrimSuffix(jobMarker, ":"))
	}

	if needsPtr != nil && *needsPtr == "" {
		*needsPtr = DetectLastJob(content)
	}

	job := generate()

	// Ensure the file ends with a newline before appending.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + job, nil
}
