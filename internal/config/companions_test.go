package config

import (
	"os"
	"testing"
)

func TestLoadCompanions(t *testing.T) {
	yaml := `citations: []
companions:
  zed: https://github.com/toba/zed-skill.git
  brew: https://github.com/toba/homebrew-skill.git
`
	path := writeTempConfig(t, yaml)
	doc, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	c := LoadCompanions(doc)
	if c == nil {
		t.Fatal("expected companions, got nil")
	}
	if c.Zed != "https://github.com/toba/zed-skill.git" {
		t.Errorf("zed = %q", c.Zed)
	}
	if c.Brew != "https://github.com/toba/homebrew-skill.git" {
		t.Errorf("brew = %q", c.Brew)
	}
}

func TestLoadCompanionsPartial(t *testing.T) {
	yaml := `citations: []
companions:
  brew: https://github.com/toba/homebrew-skill.git
`
	path := writeTempConfig(t, yaml)
	doc, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	c := LoadCompanions(doc)
	if c == nil {
		t.Fatal("expected companions, got nil")
	}
	if c.Zed != "" {
		t.Errorf("zed = %q, want empty", c.Zed)
	}
	if c.Brew != "https://github.com/toba/homebrew-skill.git" {
		t.Errorf("brew = %q", c.Brew)
	}
}

func TestLoadCompanionsMissing(t *testing.T) {
	yaml := `citations: []
`
	path := writeTempConfig(t, yaml)
	doc, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	c := LoadCompanions(doc)
	if c != nil {
		t.Errorf("expected nil, got %+v", c)
	}
}

func TestSaveCompanionsNew(t *testing.T) {
	yaml := `citations: []
`
	path := writeTempConfig(t, yaml)
	doc, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	c := &Companions{
		Zed:  "https://github.com/toba/zed-skill.git",
		Brew: "https://github.com/toba/homebrew-skill.git",
	}
	if err := SaveCompanions(doc, c); err != nil {
		t.Fatal(err)
	}

	// Reload and verify.
	doc2, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	c2 := LoadCompanions(doc2)
	if c2 == nil {
		t.Fatal("companions missing after save")
	}
	if c2.Zed != c.Zed {
		t.Errorf("zed = %q, want %q", c2.Zed, c.Zed)
	}
	if c2.Brew != c.Brew {
		t.Errorf("brew = %q, want %q", c2.Brew, c.Brew)
	}

	// Upstream section should be preserved.
	data, err := os.ReadFile(path) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !contains(content, "citations:") {
		t.Error("citations section was lost")
	}
}

func TestSaveCompanionsUpdate(t *testing.T) {
	yaml := `citations: []
companions:
  brew: https://github.com/toba/homebrew-old.git
`
	path := writeTempConfig(t, yaml)
	doc, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	c := &Companions{
		Zed:  "https://github.com/toba/zed-new.git",
		Brew: "https://github.com/toba/homebrew-new.git",
	}
	if err := SaveCompanions(doc, c); err != nil {
		t.Fatal(err)
	}

	doc2, _, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	c2 := LoadCompanions(doc2)
	if c2 == nil {
		t.Fatal("companions missing after update")
	}
	if c2.Zed != "https://github.com/toba/zed-new.git" {
		t.Errorf("zed = %q", c2.Zed)
	}
	if c2.Brew != "https://github.com/toba/homebrew-new.git" {
		t.Errorf("brew = %q", c2.Brew)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
