package brew

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"go.mod", []string{"go.mod"}, "go"},
		{"goreleaser yaml", []string{".goreleaser.yaml"}, "go"},
		{"goreleaser yml", []string{".goreleaser.yml"}, "go"},
		{"swift", []string{"Package.swift"}, "swift"},
		{"rust", []string{"Cargo.toml"}, "rust"},
		{"empty defaults to go", nil, "go"},
		{"goreleaser wins over swift", []string{".goreleaser.yaml", "Package.swift"}, "go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, f), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			orig, _ := os.Getwd()
			t.Cleanup(func() { os.Chdir(orig) })
			os.Chdir(dir)

			lang := DetectLanguage()
			if lang.Name != tt.expected {
				t.Errorf("DetectLanguage() = %q, want %q", lang.Name, tt.expected)
			}
		})
	}
}

func TestAssetName(t *testing.T) {
	tests := []struct {
		lang     string
		tool     string
		tag      string
		expected string
	}{
		{"go", "skill", "v1.0.0", "skill_darwin_arm64.tar.gz"},
		{"swift", "mytool", "v2.1.0", "mytool-v2.1.0-arm64.tar.gz"},
		{"rust", "rstool", "v0.3.0", "rstool-v0.3.0-aarch64-apple-darwin.tar.gz"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			l := Language{Name: tt.lang}
			got := l.AssetName(tt.tool, tt.tag)
			if got != tt.expected {
				t.Errorf("AssetName(%q, %q) = %q, want %q", tt.tool, tt.tag, got, tt.expected)
			}
		})
	}
}

func TestChecksumMode(t *testing.T) {
	tests := []struct {
		lang     string
		expected string
	}{
		{"go", "checksums.txt"},
		{"swift", "sidecar"},
		{"rust", "sidecar"},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			l := Language{Name: tt.lang}
			got := l.ChecksumMode()
			if got != tt.expected {
				t.Errorf("ChecksumMode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWorkflowBuildMarkers(t *testing.T) {
	tests := []struct {
		lang    string
		markers []string
	}{
		{"go", []string{"goreleaser/goreleaser-action"}},
		{"swift", []string{"swift build"}},
		{"rust", []string{"cargo build", "cross"}},
	}
	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			l := Language{Name: tt.lang}
			got := l.WorkflowBuildMarkers()
			if len(got) != len(tt.markers) {
				t.Fatalf("WorkflowBuildMarkers() returned %d markers, want %d", len(got), len(tt.markers))
			}
			for i, m := range got {
				if m != tt.markers[i] {
					t.Errorf("marker[%d] = %q, want %q", i, m, tt.markers[i])
				}
			}
		})
	}
}
