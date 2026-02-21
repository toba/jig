package nope

import (
	"path/filepath"
	"strings"
)

// hasOperatorToken returns true if the command contains an operator token
// matching the given predicate.
func hasOperatorToken(input string, match func(string) bool) bool {
	cmd := ExtractCommand(input)
	if cmd == "" {
		return false
	}
	for _, t := range ShellTokenize(cmd) {
		if t.Operator && match(t.Value) {
			return true
		}
	}
	return false
}

// CheckPipe returns true if the command contains a pipe operator outside quotes.
func CheckPipe(input string) bool {
	return hasOperatorToken(input, func(v string) bool { return v == "|" })
}

// CheckChained returns true if the command contains &&, ||, or ; outside quotes.
func CheckChained(input string) bool {
	return hasOperatorToken(input, func(v string) bool {
		return v == "&&" || v == "||" || v == ";"
	})
}

// CheckRedirect returns true if the command contains > or >> outside quotes.
func CheckRedirect(input string) bool {
	return hasOperatorToken(input, func(v string) bool { return v == ">" || v == ">>" })
}

// CheckSubshell returns true if the command contains $() or backticks outside single quotes.
func CheckSubshell(input string) bool {
	return hasOperatorToken(input, func(v string) bool { return v == "$(" || v == "`" })
}

// sensitiveExtensions are file extensions that indicate credential files.
var sensitiveExtensions = []string{".pem", ".key", ".p12", ".pfx"}

// sensitiveExactNames are exact basenames that indicate credential files.
var sensitiveExactNames = map[string]bool{
	"credentials.json": true,
	"token.pickle":     true,
	"token.json":       true,
	".netrc":           true,
	".npmrc":           true,
	"id_rsa":           true,
	"id_ed25519":       true,
	"id_ecdsa":         true,
}

// envExemptions are .env file variants that are safe to read.
var envExemptions = map[string]bool{
	".env.example":  true,
	".env.sample":   true,
	".env.template": true,
}

// sensitivePathFragments are path fragments that indicate credential directories.
var sensitivePathFragments = []string{".aws/credentials", ".ssh/"}

// CheckCredentialRead returns true if any token references a sensitive file.
func CheckCredentialRead(input string) bool {
	cmd := ExtractCommand(input)
	if cmd == "" {
		return false
	}
	for _, t := range ShellTokenize(cmd) {
		if t.Operator {
			continue
		}
		if isSensitiveFile(t.Value) {
			return true
		}
	}
	return false
}

func isSensitiveFile(s string) bool {
	base := filepath.Base(s)

	// Check .env and .env.* (but not exemptions)
	if base == ".env" {
		return true
	}
	if strings.HasPrefix(base, ".env.") && !envExemptions[base] {
		return true
	}

	// Check sensitive extensions
	for _, ext := range sensitiveExtensions {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}

	// Check exact names
	if sensitiveExactNames[base] {
		return true
	}

	// Check path fragments
	for _, frag := range sensitivePathFragments {
		if strings.Contains(s, frag) {
			return true
		}
	}

	return false
}

// networkTools are commands that perform network operations.
var networkTools = map[string]bool{
	"curl": true,
	"wget": true,
	"nc":   true,
	"ncat": true,
	"ssh":  true,
	"scp":  true,
	"sftp": true,
}

// CheckNetwork returns true if a network tool is found in command position.
// Command position is the first token or the first token after a pipe/chain operator,
// skipping env var assignments (tokens containing '=').
func CheckNetwork(input string) bool {
	cmd := ExtractCommand(input)
	if cmd == "" {
		return false
	}
	tokens := ShellTokenize(cmd)
	cmdPos := true // start in command position
	for _, t := range tokens {
		if t.Operator {
			if t.Value == "|" || t.Value == "&&" || t.Value == "||" || t.Value == ";" {
				cmdPos = true
			}
			continue
		}
		if cmdPos {
			// Skip env var assignments like FOO=bar
			if strings.Contains(t.Value, "=") && !t.Quoted {
				continue
			}
			base := filepath.Base(t.Value)
			if networkTools[base] {
				return true
			}
			cmdPos = false
		}
	}
	return false
}
