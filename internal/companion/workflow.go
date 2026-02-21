package companion

import "strings"

// DetectLastJob finds the last top-level job name in a GitHub Actions workflow.
// It looks for lines indented exactly 2 spaces followed by a word and colon,
// which is the YAML pattern for job definitions under the "jobs:" key.
// Returns "release" if no jobs are found.
func DetectLastJob(content string) string {
	last := "release"
	inJobs := false
	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "jobs:" {
			inJobs = true
			continue
		}
		if !inJobs {
			continue
		}
		// A top-level job is indented exactly 2 spaces (not more).
		if len(line) > 2 && line[0] == ' ' && line[1] == ' ' && line[2] != ' ' {
			name := strings.TrimSuffix(trimmed, ":")
			if strings.HasSuffix(trimmed, ":") && !strings.Contains(name, " ") {
				last = name
			}
		}
	}
	return last
}
