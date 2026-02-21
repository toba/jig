package update

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateTodoConfig(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    func(t *testing.T, output string)
		wantMig bool
	}{
		{
			name: "restructures issues and sync into todo section",
			input: `upstream:
    sources: []

issues:
  default_status: draft
  default_type: task
  path: .issues
sync:
  clickup:
    list_id: "123"
`,
			wantMig: true,
			want: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "todo:") {
					t.Error("missing todo: key")
				}
				if sectionExists(splitLines(output), "issues") {
					t.Error("issues: should have been removed")
				}
				if sectionExists(splitLines(output), "sync") {
					t.Error("top-level sync: should have been removed")
				}
				if !strings.Contains(output, "upstream:") {
					t.Error("upstream section was lost")
				}
				if !strings.Contains(output, "default_status: draft") {
					t.Error("default_status was lost")
				}
				if !strings.Contains(output, "list_id:") {
					t.Error("sync config was lost")
				}
			},
		},
		{
			name: "issues only, no sync",
			input: `issues:
  default_status: ready
  path: .issues
`,
			wantMig: true,
			want: func(t *testing.T, output string) {
				t.Helper()
				if !strings.Contains(output, "todo:") {
					t.Error("missing todo: key")
				}
				if sectionExists(splitLines(output), "issues") {
					t.Error("issues: should have been removed")
				}
				if !strings.Contains(output, "default_status: ready") {
					t.Error("default_status was lost")
				}
			},
		},
		{
			name:    "skips when todo already exists",
			input:   "todo:\n  path: .issues\n",
			wantMig: false,
		},
		{
			name:    "skips when no issues key",
			input:   "upstream:\n  sources: []\n",
			wantMig: false,
		},
		{
			name:    "skips when file does not exist",
			input:   "", // won't write a file
			wantMig: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tobaPath := filepath.Join(dir, ".toba.yaml")

			if tc.input != "" {
				writeFile(t, tobaPath, tc.input)
			}

			migrated, err := migrateTodoConfig(tobaPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if migrated != tc.wantMig {
				t.Fatalf("migrated = %v, want %v", migrated, tc.wantMig)
			}
			if tc.want != nil {
				output := readFile(t, tobaPath)
				tc.want(t, output)
			}
		})
	}
}
