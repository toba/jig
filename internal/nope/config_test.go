package nope

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("testdata/nope.yaml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if len(cfg.Rules) != 24 {
		t.Fatalf("expected 24 rules, got %d", len(cfg.Rules))
	}

	// Verify first rule is the multiline builtin
	r := cfg.Rules[0]
	if r.Name != "multiline-commands" {
		t.Errorf("first rule name = %q, want multiline-commands", r.Name)
	}
	if r.Builtin != "multiline" {
		t.Errorf("first rule builtin = %q, want multiline", r.Builtin)
	}

	// Verify a regex rule
	r = cfg.Rules[1]
	if r.Name != "git-checkout-switch" {
		t.Errorf("second rule name = %q, want git-checkout-switch", r.Name)
	}
	if r.Pattern == "" {
		t.Error("second rule should have a pattern")
	}
}

func TestCompileRules(t *testing.T) {
	cfg, err := LoadConfig("testdata/nope.yaml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	rules, err := CompileRules(cfg.Rules)
	if err != nil {
		t.Fatalf("CompileRules: %v", err)
	}
	if len(rules) != 24 {
		t.Fatalf("expected 24 compiled rules, got %d", len(rules))
	}
}

func TestCompileRulesValidation(t *testing.T) {
	tests := []struct {
		name    string
		rules   []RuleDef
		wantErr string
	}{
		{
			name:    "both pattern and builtin",
			rules:   []RuleDef{{Name: "bad", Pattern: "foo", Builtin: "multiline", Message: "msg"}},
			wantErr: "mutually exclusive",
		},
		{
			name:    "neither pattern nor builtin",
			rules:   []RuleDef{{Name: "bad", Message: "msg"}},
			wantErr: "must have pattern or builtin",
		},
		{
			name:    "missing message",
			rules:   []RuleDef{{Name: "bad", Pattern: "foo"}},
			wantErr: "message is required",
		},
		{
			name:    "bad regex",
			rules:   []RuleDef{{Name: "bad", Pattern: "[invalid", Message: "msg"}},
			wantErr: "bad pattern",
		},
		{
			name:    "unknown builtin",
			rules:   []RuleDef{{Name: "bad", Builtin: "nope", Message: "msg"}},
			wantErr: "unknown builtin",
		},
		{
			name:    "builtin with non-Bash tool",
			rules:   []RuleDef{{Name: "bad", Builtin: "multiline", Tools: []string{"Write"}, Message: "msg"}},
			wantErr: "builtin rules only support Bash tool",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompileRules(tt.rules)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := LoadConfig("testdata/nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte(":\n  :\n    - ["), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadConfigNoNopeSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".toba.yaml")
	if err := os.WriteFile(path, []byte("upstream:\n  sources: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for missing nope section")
	}
	if !strings.Contains(err.Error(), "no 'nope' section") {
		t.Errorf("error = %q, want substring %q", err.Error(), "no 'nope' section")
	}
}
