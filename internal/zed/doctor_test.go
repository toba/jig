package zed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGoreleaserExists_Yaml(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	// No goreleaser file -> false.
	if checkGoreleaserExists() {
		t.Error("expected false when no goreleaser file exists")
	}

	// Create .goreleaser.yaml -> true.
	if err := os.WriteFile(filepath.Join(tmp, ".goreleaser.yaml"), []byte("builds: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !checkGoreleaserExists() {
		t.Error("expected true when .goreleaser.yaml exists")
	}
}

func TestCheckGoreleaserExists_Yml(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	// Create .goreleaser.yml (not .yaml) -> true.
	if err := os.WriteFile(filepath.Join(tmp, ".goreleaser.yml"), []byte("builds: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !checkGoreleaserExists() {
		t.Error("expected true when .goreleaser.yml exists")
	}
}

func TestCheckGoreleaserExists_BothVariants(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	// Create both variants; .yaml should be found first.
	os.WriteFile(filepath.Join(tmp, ".goreleaser.yaml"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(tmp, ".goreleaser.yml"), []byte("b"), 0o644)
	if !checkGoreleaserExists() {
		t.Error("expected true when both goreleaser files exist")
	}
}

func TestRunDoctor_InvalidExtFormat(t *testing.T) {
	// No slash in Ext -> should return 1 immediately.
	code := RunDoctor(DoctorOpts{Ext: "noSlash", Repo: "toba/lsp"})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunDoctor_EmptyExt(t *testing.T) {
	// Empty Ext string -> SplitN gives [""] which has len 1 -> invalid format.
	code := RunDoctor(DoctorOpts{Ext: "", Repo: "toba/lsp"})
	if code != 1 {
		t.Errorf("expected exit code 1 for empty ext, got %d", code)
	}
}

func TestDoctorOpts_Struct(t *testing.T) {
	opts := DoctorOpts{
		Ext:  "toba/gozer",
		Repo: "toba/go-template-lsp",
	}
	if opts.Ext != "toba/gozer" {
		t.Errorf("unexpected Ext: %s", opts.Ext)
	}
	if opts.Repo != "toba/go-template-lsp" {
		t.Errorf("unexpected Repo: %s", opts.Repo)
	}
}

func TestInjectWorkflow_WritesFile(t *testing.T) {
	tmp := t.TempDir()
	wfPath := filepath.Join(tmp, "release.yml")

	content := `name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
	if err := os.WriteFile(wfPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := WorkflowParams{Org: "toba", Ext: "gozer"}
	if err := injectWorkflow(wfPath, p); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)

	for _, want := range []string{"sync-extension:", "toba/gozer/dispatches", "EXTENSION_PAT"} {
		if !strings.Contains(got, want) {
			t.Errorf("injected workflow missing %q", want)
		}
	}
}

func TestInjectWorkflow_AlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	wfPath := filepath.Join(tmp, "release.yml")

	content := `jobs:
  release:
    runs-on: ubuntu-latest
  sync-extension:
    needs: release
`
	if err := os.WriteFile(wfPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := WorkflowParams{Org: "toba", Ext: "gozer"}
	err := injectWorkflow(wfPath, p)
	if err == nil {
		t.Error("expected error when sync-extension already exists")
	}
}

func TestInjectWorkflow_MissingFile(t *testing.T) {
	err := injectWorkflow("/nonexistent/path/release.yml", WorkflowParams{Org: "toba", Ext: "gozer"})
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// mockGH creates a fake gh script that always succeeds (exit 0) and writes it
// to a temp bin directory. Returns the bin dir to prepend to PATH.
func mockGH(t *testing.T, script string) string {
	t.Helper()
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ghPath := filepath.Join(binDir, "gh")
	if err := os.WriteFile(ghPath, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return binDir
}

func TestRunDoctor_WithMockGH_AllChecksPass(t *testing.T) {
	// Mock gh to always succeed for repo/api/release commands.
	releaseJSON, _ := json.Marshal([]map[string]string{{"tagName": "v1.0.0"}})
	assetJSON, _ := json.Marshal(map[string]interface{}{
		"assets": []map[string]string{
			{"name": "my-tool_darwin_arm64.tar.gz"},
			{"name": "my-tool_linux_amd64.tar.gz"},
		},
	})
	// Mock gh: for "release list" return tag JSON, for "release view" return asset JSON,
	// for everything else succeed silently.
	script := `
case "$*" in
  *"release list"*) echo '` + string(releaseJSON) + `' ;;
  *"release view"*) echo '` + string(assetJSON) + `' ;;
  *) echo '{}' ;;
esac
`
	binDir := mockGH(t, script)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	// Set up working directory with goreleaser and workflow files.
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	os.WriteFile(".goreleaser.yaml", []byte("builds: []\n"), 0o644)
	os.MkdirAll(".github/workflows", 0o755)
	wfContent := `name: Release

jobs:
  release:
    runs-on: ubuntu-latest
  sync-extension:
    needs: release
    steps:
      - name: Dispatch version bump to toba/gozer
        run: |
          gh api repos/toba/gozer/dispatches
        env:
          GH_TOKEN: ${{ secrets.EXTENSION_PAT }}
`
	os.WriteFile(".github/workflows/release.yml", []byte(wfContent), 0o644)

	code := RunDoctor(DoctorOpts{Ext: "toba/gozer", Repo: "toba/my-tool"})
	if code != 0 {
		t.Errorf("expected exit code 0 (all checks pass with mock), got %d", code)
	}
}

func TestRunDoctor_MissingWorkflow(t *testing.T) {
	// Mock gh to always succeed.
	releaseJSON, _ := json.Marshal([]map[string]string{{"tagName": "v1.0.0"}})
	assetJSON, _ := json.Marshal(map[string]interface{}{
		"assets": []map[string]string{
			{"name": "my-tool_darwin_arm64.tar.gz"},
			{"name": "my-tool_linux_amd64.tar.gz"},
		},
	})
	script := `
case "$*" in
  *"release list"*) echo '` + string(releaseJSON) + `' ;;
  *"release view"*) echo '` + string(assetJSON) + `' ;;
  *) echo '{}' ;;
esac
`
	binDir := mockGH(t, script)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	// Working directory with no workflow file and no goreleaser.
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	code := RunDoctor(DoctorOpts{Ext: "toba/gozer", Repo: "toba/my-tool"})
	if code != 1 {
		t.Errorf("expected exit code 1 (missing workflow), got %d", code)
	}
}

func TestRunDoctor_WorkflowMissingSyncJob(t *testing.T) {
	// Mock gh to succeed.
	releaseJSON, _ := json.Marshal([]map[string]string{{"tagName": "v1.0.0"}})
	assetJSON, _ := json.Marshal(map[string]interface{}{
		"assets": []map[string]string{
			{"name": "my-tool_darwin_arm64.tar.gz"},
			{"name": "my-tool_linux_amd64.tar.gz"},
		},
	})
	script := `
case "$*" in
  *"release list"*) echo '` + string(releaseJSON) + `' ;;
  *"release view"*) echo '` + string(assetJSON) + `' ;;
  *) echo '{}' ;;
esac
`
	binDir := mockGH(t, script)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	os.WriteFile(".goreleaser.yaml", []byte("builds: []\n"), 0o644)
	os.MkdirAll(".github/workflows", 0o755)
	// Workflow without sync-extension job.
	wfContent := `name: Release

jobs:
  release:
    runs-on: ubuntu-latest
`
	os.WriteFile(".github/workflows/release.yml", []byte(wfContent), 0o644)

	code := RunDoctor(DoctorOpts{Ext: "toba/gozer", Repo: "toba/my-tool"})
	if code != 1 {
		t.Errorf("expected exit code 1 (missing sync-extension job), got %d", code)
	}
}

func TestRunDoctor_WorkflowMissingExtensionPAT(t *testing.T) {
	releaseJSON, _ := json.Marshal([]map[string]string{{"tagName": "v1.0.0"}})
	assetJSON, _ := json.Marshal(map[string]interface{}{
		"assets": []map[string]string{
			{"name": "my-tool_darwin_arm64.tar.gz"},
			{"name": "my-tool_linux_amd64.tar.gz"},
		},
	})
	script := `
case "$*" in
  *"release list"*) echo '` + string(releaseJSON) + `' ;;
  *"release view"*) echo '` + string(assetJSON) + `' ;;
  *) echo '{}' ;;
esac
`
	binDir := mockGH(t, script)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	tmp := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	os.WriteFile(".goreleaser.yaml", []byte("builds: []\n"), 0o644)
	os.MkdirAll(".github/workflows", 0o755)
	// Has sync-extension and repo ref but no EXTENSION_PAT.
	wfContent := `name: Release

jobs:
  release:
    runs-on: ubuntu-latest
  sync-extension:
    needs: release
    steps:
      - name: Dispatch to toba/gozer
        run: echo hello
`
	os.WriteFile(".github/workflows/release.yml", []byte(wfContent), 0o644)

	code := RunDoctor(DoctorOpts{Ext: "toba/gozer", Repo: "toba/my-tool"})
	if code != 1 {
		t.Errorf("expected exit code 1 (missing EXTENSION_PAT), got %d", code)
	}
}

func TestCheckReleaseAssets_BothPlatforms(t *testing.T) {
	assetJSON, _ := json.Marshal(map[string]interface{}{
		"assets": []map[string]string{
			{"name": "mytool_darwin_arm64.tar.gz"},
			{"name": "mytool_linux_amd64.tar.gz"},
		},
	})
	script := `echo '` + string(assetJSON) + `'`
	binDir := mockGH(t, script)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	ok := true
	checkReleaseAssets("toba/mytool", "v1.0.0", "mytool", &ok)
	if !ok {
		t.Error("expected ok=true when both platform assets exist")
	}
}

func TestCheckReleaseAssets_MissingDarwin(t *testing.T) {
	assetJSON, _ := json.Marshal(map[string]interface{}{
		"assets": []map[string]string{
			{"name": "mytool_linux_amd64.tar.gz"},
		},
	})
	script := `echo '` + string(assetJSON) + `'`
	binDir := mockGH(t, script)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	ok := true
	checkReleaseAssets("toba/mytool", "v1.0.0", "mytool", &ok)
	if ok {
		t.Error("expected ok=false when darwin asset missing")
	}
}

func TestCheckReleaseAssets_GHFails(t *testing.T) {
	binDir := mockGH(t, "exit 1")
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	ok := true
	checkReleaseAssets("toba/mytool", "v1.0.0", "mytool", &ok)
	if ok {
		t.Error("expected ok=false when gh command fails")
	}
}

func TestCheckReleaseAssets_InvalidJSON(t *testing.T) {
	binDir := mockGH(t, `echo 'not json'`)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", binDir+":"+origPath)
	defer os.Setenv("PATH", origPath)

	ok := true
	checkReleaseAssets("toba/mytool", "v1.0.0", "mytool", &ok)
	if ok {
		t.Error("expected ok=false for invalid JSON")
	}
}

func TestInjectWorkflow_PreservesOriginalContent(t *testing.T) {
	tmp := t.TempDir()
	wfPath := filepath.Join(tmp, "release.yml")

	content := `name: Release

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
`
	if err := os.WriteFile(wfPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := injectWorkflow(wfPath, WorkflowParams{Org: "toba", Ext: "gozer"}); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(wfPath)
	got := string(data)

	// Original content should still be present.
	if !strings.Contains(got, "name: Release") {
		t.Error("original workflow name lost after injection")
	}
	if !strings.Contains(got, "actions/checkout@v4") {
		t.Error("original steps lost after injection")
	}
}
