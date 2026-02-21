package brew

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	urlRe     = regexp.MustCompile(`(?m)^(\s*url\s+)"[^"]*"`)
	versionRe = regexp.MustCompile(`(?m)^(\s*version\s+)"[^"]*"`)
	sha256Re  = regexp.MustCompile(`(?m)^(\s*sha256\s+)"[^"]*"`)
)

// Formula holds the parsed fields from a Homebrew formula file.
type Formula struct {
	Path    string // filesystem path to the .rb file
	Content string // raw file content
	URL     string // current url value
	Version string // current version value
	SHA256  string // current sha256 value
}

// ReadFormula reads and parses a .rb formula file.
func ReadFormula(path string) (*Formula, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading formula: %w", err)
	}
	content := string(data)
	f := &Formula{
		Path:    path,
		Content: content,
		URL:     extractQuoted(content, `url`),
		Version: extractQuoted(content, `version`),
		SHA256:  extractQuoted(content, `sha256`),
	}
	if f.URL == "" {
		return nil, fmt.Errorf("no url field found in %s", path)
	}
	return f, nil
}

// Update rewrites the formula with new version, URL, and SHA256 values.
// It preserves all other content (class name, desc, depends_on, install, test, caveats).
func (f *Formula) Update(newURL, newVersion, newSHA256 string) string {
	s := f.Content
	s = urlRe.ReplaceAllString(s, `${1}"`+newURL+`"`)
	s = versionRe.ReplaceAllString(s, `${1}"`+newVersion+`"`)
	s = sha256Re.ReplaceAllString(s, `${1}"`+newSHA256+`"`)
	return s
}

// WriteFormula writes updated content back to the formula file.
func WriteFormula(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

// FormulaParams holds the inputs needed to generate a new formula from scratch.
type FormulaParams struct {
	Tool    string // binary name, e.g. "todo"
	Desc    string // one-line description
	Repo    string // e.g. "toba/todo"
	Tag     string // e.g. "v0.14.2"
	Asset   string // e.g. "todo_darwin_arm64.tar.gz"
	SHA256  string // hex-encoded sha256
	License string // e.g. "Apache-2.0"
}

// GenerateFormula produces the full .rb content for a Homebrew formula.
func GenerateFormula(p FormulaParams) string {
	version := strings.TrimPrefix(p.Tag, "v")
	className := formulaClassName(p.Tool)
	url := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", p.Repo, p.Tag, p.Asset)
	return fmt.Sprintf(`class %s < Formula
  desc "%s"
  homepage "https://github.com/%s"
  url "%s"
  version "%s"
  sha256 "%s"
  license "%s"

  depends_on :macos
  depends_on arch: :arm64

  def install
    bin.install "%s"
  end

  test do
    assert_match "%s", shell_output("#{bin}/%s version")
  end
end
`, className, p.Desc, p.Repo, url, version, p.SHA256, p.License, p.Tool, p.Tool, p.Tool)
}

// formulaClassName converts a tool name to a Ruby class name.
// e.g. "todo" → "Todo", "go-bigq" → "GoBigq"
func formulaClassName(tool string) string {
	var result []byte
	upper := true
	for i := range len(tool) {
		c := tool[i]
		if c == '-' || c == '_' {
			upper = true
			continue
		}
		if upper {
			if c >= 'a' && c <= 'z' {
				c -= 32
			}
			upper = false
		}
		result = append(result, c)
	}
	return string(result)
}

// extractQuoted pulls the quoted value from a line like:  url "https://..."
func extractQuoted(content, field string) string {
	re := regexp.MustCompile(`(?m)^\s*` + field + `\s+"([^"]*)"`)
	m := re.FindStringSubmatch(content)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
