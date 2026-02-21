package brew

import (
	"strings"
	"testing"

	"github.com/toba/jig/internal/companion"
)

func TestGenerateWorkflowJob(t *testing.T) {
	job := GenerateWorkflowJob(WorkflowParams{
		Tool:    "todo",
		Org:     "toba",
		Desc:    "Issue tracker",
		License: "Apache-2.0",
		Asset:   "todo_darwin_arm64.tar.gz",
	})

	checks := []string{
		"update-homebrew:",
		"needs: release",
		"HOMEBREW_TAP_TOKEN",
		"todo_darwin_arm64.tar.gz",
		"toba/homebrew-todo",
		"Formula/todo.rb",
		`class Todo < Formula`,
		`desc "Issue tracker"`,
		`license "Apache-2.0"`,
		`bin.install "todo"`,
		"bump to ${VERSION}",
	}
	for _, want := range checks {
		if !strings.Contains(job, want) {
			t.Errorf("job missing %q", want)
		}
	}
}

func TestGenerateWorkflowJobCustomNeeds(t *testing.T) {
	job := GenerateWorkflowJob(WorkflowParams{
		Tool:  "tool",
		Org:   "org",
		Asset: "tool_darwin_arm64.tar.gz",
		Needs: "checksums",
	})
	if !strings.Contains(job, "needs: checksums") {
		t.Error("expected needs: checksums")
	}
}

func TestInjectWorkflowJob(t *testing.T) {
	existing := `name: Release

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

	p := WorkflowParams{
		Tool:    "todo",
		Org:     "toba",
		Desc:    "Issue tracker",
		License: "Apache-2.0",
		Asset:   "todo_darwin_arm64.tar.gz",
	}

	result, err := InjectWorkflowJob(existing, p)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "update-homebrew:") {
		t.Error("result missing update-homebrew job")
	}
	if !strings.Contains(result, "needs: release") {
		t.Error("result should depend on release job")
	}
}

func TestInjectWorkflowJobAlreadyExists(t *testing.T) {
	existing := `jobs:
  release:
    runs-on: ubuntu-latest
  update-homebrew:
    needs: release
`
	_, err := InjectWorkflowJob(existing, WorkflowParams{Tool: "todo", Org: "toba"})
	if err == nil {
		t.Error("expected error for existing update-homebrew job")
	}
}

func TestDetectLastJob(t *testing.T) {
	content := `jobs:
  build:
    runs-on: ubuntu-latest
  checksums:
    needs: build
  release:
    needs: checksums
`
	got := companion.DetectLastJob(content)
	if got != "release" {
		t.Errorf("DetectLastJob = %q, want release", got)
	}
}
