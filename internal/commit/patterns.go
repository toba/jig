package commit

import "regexp"

// gitignorePatterns are regex patterns that suggest an untracked file should be
// listed in .gitignore rather than committed. Extracted from the duplicated
// GITIGNORE_PATTERNS arrays across ~18 repos' commit scripts.
var gitignorePatterns = []*regexp.Regexp{
	// Build artifacts
	regexp.MustCompile(`\.log$`),
	regexp.MustCompile(`\.tmp$`),
	regexp.MustCompile(`\.cache$`),
	regexp.MustCompile(`\.o$`),
	regexp.MustCompile(`\.a$`),
	regexp.MustCompile(`\.so$`),
	regexp.MustCompile(`\.dylib$`),

	// Python
	regexp.MustCompile(`\.pyc$`),
	regexp.MustCompile(`\.pyo$`),
	regexp.MustCompile(`__pycache__/`),
	regexp.MustCompile(`\.venv/`),
	regexp.MustCompile(`venv/`),

	// Node
	regexp.MustCompile(`node_modules/`),

	// Editor/OS
	regexp.MustCompile(`\.env$`),
	regexp.MustCompile(`\.env\.local$`),
	regexp.MustCompile(`\.DS_Store$`),
	regexp.MustCompile(`\.swp$`),
	regexp.MustCompile(`\.swo$`),
	regexp.MustCompile(`\.idea/`),

	// Build dirs
	regexp.MustCompile(`dist/`),
	regexp.MustCompile(`build/`),
	regexp.MustCompile(`coverage/`),
	regexp.MustCompile(`\.coverage$`),

	// Secrets
	regexp.MustCompile(`credentials\.`),
	regexp.MustCompile(`secrets\.`),
	regexp.MustCompile(`\.key$`),
	regexp.MustCompile(`\.pem$`),
	regexp.MustCompile(`\.p12$`),

	// iOS/macOS (from Swift project scripts)
	regexp.MustCompile(`DerivedData/`),
	regexp.MustCompile(`\.xcuserstate$`),
	regexp.MustCompile(`xcuserdata/`),
	regexp.MustCompile(`\.moved-aside$`),
	regexp.MustCompile(`Pods/`),
}
