package brew

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleFormula = `class Todo < Formula
  desc "Issue tracker for you, your team, and your coding agents"
  homepage "https://github.com/toba/todo"
  url "https://github.com/toba/todo/releases/download/v0.14.2/todo_darwin_arm64.tar.gz"
  version "0.14.2"
  sha256 "d7274cc62c978ac6aa6f085f0138b6612634cce8b1475bdcc22db7c6de1bb914"
  license "Apache-2.0"

  depends_on :macos
  depends_on arch: :arm64

  def install
    bin.install "todo"
  end

  test do
    assert_match "todo", shell_output("#{bin}/todo version")
  end
end
`

func TestReadFormula(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todo.rb")
	if err := os.WriteFile(path, []byte(sampleFormula), 0o644); err != nil {
		t.Fatal(err)
	}

	f, err := ReadFormula(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.URL != "https://github.com/toba/todo/releases/download/v0.14.2/todo_darwin_arm64.tar.gz" {
		t.Errorf("URL = %q", f.URL)
	}
	if f.Version != "0.14.2" {
		t.Errorf("Version = %q", f.Version)
	}
	if f.SHA256 != "d7274cc62c978ac6aa6f085f0138b6612634cce8b1475bdcc22db7c6de1bb914" {
		t.Errorf("SHA256 = %q", f.SHA256)
	}
}

func TestFormulaUpdate(t *testing.T) {
	f := &Formula{
		Content: sampleFormula,
		URL:     "https://github.com/toba/todo/releases/download/v0.14.2/todo_darwin_arm64.tar.gz",
		Version: "0.14.2",
		SHA256:  "d7274cc62c978ac6aa6f085f0138b6612634cce8b1475bdcc22db7c6de1bb914",
	}

	updated := f.Update(
		"https://github.com/toba/todo/releases/download/v0.15.0/todo_darwin_arm64.tar.gz",
		"0.15.0",
		"aabbccdd",
	)

	got := &Formula{Content: updated}
	got.URL = extractQuoted(updated, "url")
	got.Version = extractQuoted(updated, "version")
	got.SHA256 = extractQuoted(updated, "sha256")

	if got.URL != "https://github.com/toba/todo/releases/download/v0.15.0/todo_darwin_arm64.tar.gz" {
		t.Errorf("updated URL = %q", got.URL)
	}
	if got.Version != "0.15.0" {
		t.Errorf("updated Version = %q", got.Version)
	}
	if got.SHA256 != "aabbccdd" {
		t.Errorf("updated SHA256 = %q", got.SHA256)
	}

	// Verify non-formula content is preserved.
	if !strings.Contains(updated, `desc "Issue tracker`) {
		t.Error("desc line was lost")
	}
	if !strings.Contains(updated, `bin.install "todo"`) {
		t.Error("install block was lost")
	}
	if !strings.Contains(updated, `depends_on :macos`) {
		t.Error("depends_on was lost")
	}
}

func TestReadFormulaMissingURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.rb")
	if err := os.WriteFile(path, []byte("class Bad < Formula\nend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ReadFormula(path)
	if err == nil {
		t.Error("expected error for formula without url")
	}
}

func TestGenerateFormula(t *testing.T) {
	got := GenerateFormula(FormulaParams{
		Tool:    "todo",
		Desc:    "Issue tracker for you, your team, and your coding agents",
		Repo:    "toba/todo",
		Tag:     "v0.14.2",
		Asset:   "todo_darwin_arm64.tar.gz",
		SHA256:  "d7274cc62c978ac6aa6f085f0138b6612634cce8b1475bdcc22db7c6de1bb914",
		License: "Apache-2.0",
	})

	checks := []string{
		`class Todo < Formula`,
		`desc "Issue tracker for you, your team, and your coding agents"`,
		`homepage "https://github.com/toba/todo"`,
		`url "https://github.com/toba/todo/releases/download/v0.14.2/todo_darwin_arm64.tar.gz"`,
		`version "0.14.2"`,
		`sha256 "d7274cc62c978ac6aa6f085f0138b6612634cce8b1475bdcc22db7c6de1bb914"`,
		`license "Apache-2.0"`,
		`bin.install "todo"`,
		`assert_match "todo"`,
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("formula missing %q", want)
		}
	}
}

func TestFormulaClassName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"todo", "Todo"},
		{"go-bigq", "GoBigq"},
		{"my_tool", "MyTool"},
		{"skill", "Skill"},
	}
	for _, tt := range tests {
		got := formulaClassName(tt.in)
		if got != tt.want {
			t.Errorf("formulaClassName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
