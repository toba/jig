package zed

import (
	"strings"
	"testing"
)

func TestParseLanguages(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"CSS", []string{"CSS"}},
		{"Go Text Template,Go HTML Template", []string{"Go Text Template", "Go HTML Template"}},
		{" CSS , HTML ", []string{"CSS", "HTML"}},
		{"", nil},
		{",,,", nil},
	}
	for _, tt := range tests {
		got := parseLanguages(tt.in)
		if len(got) != len(tt.want) {
			t.Errorf("parseLanguages(%q) = %v, want %v", tt.in, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseLanguages(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
			}
		}
	}
}

func TestParseLanguages_SingleTrailingComma(t *testing.T) {
	got := parseLanguages("Go,")
	if len(got) != 1 || got[0] != "Go" {
		t.Errorf("parseLanguages(\"Go,\") = %v, want [Go]", got)
	}
}

func TestParseLanguages_WhitespaceOnly(t *testing.T) {
	got := parseLanguages("  ,  ,  ")
	if got != nil {
		t.Errorf("parseLanguages(\"  ,  ,  \") = %v, want nil", got)
	}
}

func TestInitOpts_Struct(t *testing.T) {
	opts := InitOpts{
		Ext:       "toba/gozer",
		Tag:       "v0.14.0",
		Repo:      "toba/go-template-lsp",
		Desc:      "Test desc",
		LSPName:   "go-template-lsp",
		Languages: "CSS",
		DryRun:    true,
	}
	if opts.Ext != "toba/gozer" {
		t.Errorf("expected ext toba/gozer, got %s", opts.Ext)
	}
	if !opts.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestInitResult_Struct(t *testing.T) {
	r := InitResult{
		Ext:         "toba/gozer",
		Repo:        "toba/go-template-lsp",
		Tag:         "v0.14.0",
		LSPName:     "go-template-lsp",
		ExtCreated:  true,
		ExtPushed:   true,
		WorkflowMod: false,
	}
	if r.Ext != "toba/gozer" {
		t.Errorf("unexpected Ext: %s", r.Ext)
	}
	if !r.ExtCreated {
		t.Error("expected ExtCreated true")
	}
	if r.WorkflowMod {
		t.Error("expected WorkflowMod false")
	}
}

func TestInjectSyncExtensionJob_DetectsLastJob(t *testing.T) {
	// Workflow with multiple jobs; sync-extension should depend on the last one.
	existing := `name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
  checksums:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - run: echo done
`
	p := WorkflowParams{Org: "toba", Ext: "gozer"}
	result, err := InjectSyncExtensionJob(existing, p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "needs: checksums") {
		t.Error("expected sync-extension to depend on checksums (last job)")
	}
}

func TestInjectSyncExtensionJob_NoTrailingNewline(t *testing.T) {
	// Content without trailing newline should still work.
	existing := `jobs:
  release:
    runs-on: ubuntu-latest`

	p := WorkflowParams{Org: "toba", Ext: "gozer"}
	result, err := InjectSyncExtensionJob(existing, p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "sync-extension:") {
		t.Error("expected sync-extension job in result")
	}
}
