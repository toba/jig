package zed

import (
	"strings"
	"testing"
)

var testParams = ExtensionParams{
	ExtID:     "gozer",
	ExtName:   "Gozer",
	Version:   "0.14.0",
	Desc:      "Go template support with LSP",
	Org:       "toba",
	ExtRepo:   "gozer",
	LSPRepo:   "toba/go-template-lsp",
	LSPName:   "go-template-lsp",
	Languages: []string{"Go Text Template", "Go HTML Template"},
}

func TestGenerateExtensionToml(t *testing.T) {
	got := GenerateExtensionToml(testParams)

	checks := []string{
		`id = "gozer"`,
		`name = "Gozer"`,
		`version = "0.14.0"`,
		`schema_version = 1`,
		`authors = ["Jason Abbott"]`,
		`description = "Go template support with LSP"`,
		`repository = "https://github.com/toba/gozer"`,
		`[language_servers.gozer]`,
		`name = "Gozer"`,
		`languages = ["Go Text Template", "Go HTML Template"]`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("extension.toml missing %q", want)
		}
	}
}

func TestGenerateCargoToml(t *testing.T) {
	got := GenerateCargoToml(testParams)

	checks := []string{
		`name = "go-template-lsp"`,
		`version = "0.14.0"`,
		`edition = "2021"`,
		`license = "MIT"`,
		`crate-type = ["cdylib"]`,
		`zed_extension_api = "0.7.0"`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("Cargo.toml missing %q", want)
		}
	}
}

func TestGenerateLibRs(t *testing.T) {
	got := GenerateLibRs(testParams)

	checks := []string{
		`const GITHUB_REPO: &str = "toba/go-template-lsp"`,
		`const BINARY_NAME: &str = "go-template-lsp"`,
		"struct GozerExtension",
		"impl GozerExtension",
		"impl zed::Extension for GozerExtension",
		"zed::register_extension!(GozerExtension)",
		"cached_binary_path",
		"language_server_binary_path",
		"language_server_command",
		"zed::latest_github_release",
		"zed::download_file",
		"zed::make_file_executable",
		`go-template-lsp_{os}_{arch}.{ext}`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("lib.rs missing %q", want)
		}
	}
}

func TestGenerateLibRsHyphenatedName(t *testing.T) {
	p := testParams
	p.ExtID = "my-cool-ext"
	got := GenerateLibRs(p)

	if !strings.Contains(got, "struct MyCoolExtExtension") {
		t.Error("expected PascalCase struct name for hyphenated id")
	}
}

func TestGenerateBumpVersionScript(t *testing.T) {
	got := GenerateBumpVersionScript()

	checks := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"extension.toml",
		"Cargo.toml",
		"cargo generate-lockfile",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("bump-version.sh missing %q", want)
		}
	}
}

func TestGenerateBumpVersionWorkflow(t *testing.T) {
	got := GenerateBumpVersionWorkflow()

	checks := []string{
		"name: Bump Version",
		"repository_dispatch",
		"bump-version",
		"actions/checkout@v4",
		"dtolnay/rust-toolchain@stable",
		"bump-version.sh",
		"git commit",
		"git tag",
		"git push",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("bump-version.yml missing %q", want)
		}
	}
}

func TestGenerateLicense(t *testing.T) {
	got := GenerateLicense()

	checks := []string{
		"MIT License",
		"Copyright (c)",
		"Toba",
		"Permission is hereby granted",
		"WITHOUT WARRANTY OF ANY KIND",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("LICENSE missing %q", want)
		}
	}
}

func TestGenerateReadme(t *testing.T) {
	got := GenerateReadme(testParams)

	checks := []string{
		"# Gozer",
		"Go template support with LSP",
		"go-template-lsp",
		"github.com/toba/go-template-lsp",
		"Cmd+Shift+X",
		"zed: install dev extension",
		"cargo build --target wasm32-wasip1",
		"MIT License",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("README missing %q", want)
		}
	}
}

func TestPascalCase(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"gozer", "Gozer"},
		{"gossamer", "Gossamer"},
		{"my-ext", "MyExt"},
		{"my_cool_ext", "MyCoolExt"},
		{"already", "Already"},
	}
	for _, tt := range tests {
		got := pascalCase(tt.in)
		if got != tt.want {
			t.Errorf("pascalCase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGenerateExtensionTomlSingleLanguage(t *testing.T) {
	p := testParams
	p.Languages = []string{"CSS"}
	got := GenerateExtensionToml(p)

	if !strings.Contains(got, `languages = ["CSS"]`) {
		t.Error("expected single language in brackets")
	}
}
