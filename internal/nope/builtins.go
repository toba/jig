package nope

import (
	"path/filepath"
	"slices"
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

// CheckExfiltration returns true if a command exfiltrates sensitive files over
// the network. It detects: curl/wget uploading sensitive files, scp of sensitive
// files, bash /dev/tcp or /dev/udp socket writes, and piped credential access
// to network tools.
func CheckExfiltration(input string) bool {
	cmd := ExtractCommand(input)
	if cmd == "" {
		return false
	}
	tokens := ShellTokenize(cmd)

	// Split into segments at pipe/chain operators.
	type segment struct {
		tokens []Token
		kind   string // "pipe", "chain", or "first"
	}
	var segments []segment
	var cur []Token
	var curKind string = "first"

	for _, t := range tokens {
		if t.Operator && (t.Value == "|" || t.Value == "&&" || t.Value == "||" || t.Value == ";") {
			segments = append(segments, segment{cur, curKind})
			if t.Value == "|" {
				curKind = "pipe"
			} else {
				curKind = "chain"
			}
			cur = nil
			continue
		}
		cur = append(cur, t)
	}
	segments = append(segments, segment{cur, curKind})

	for i, seg := range segments {
		unwrapped := SkipWrappers(seg.tokens)
		if len(unwrapped) == 0 {
			continue
		}
		base := filepath.Base(unwrapped[0].Value)

		// (a) curl uploading sensitive files
		if base == "curl" {
			if checkCurlExfil(unwrapped[1:]) {
				return true
			}
		}

		// (b) wget --post-file with sensitive file
		if base == "wget" {
			if checkWgetExfil(unwrapped[1:]) {
				return true
			}
		}

		// (c) scp of sensitive files
		if base == "scp" {
			if checkScpExfil(unwrapped[1:]) {
				return true
			}
		}

		// (d) /dev/tcp or /dev/udp tokens
		for _, t := range seg.tokens {
			if !t.Operator && (strings.Contains(t.Value, "/dev/tcp/") || strings.Contains(t.Value, "/dev/udp/")) {
				return true
			}
		}

		// (e) Piped credential access to network tool:
		// Previous segment has a sensitive file, and this segment (after pipe)
		// starts with a network tool.
		if seg.kind == "pipe" && i > 0 && networkTools[base] {
			prev := segments[i-1]
			for _, t := range prev.tokens {
				if !t.Operator && isSensitiveFile(t.Value) {
					return true
				}
			}
		}
	}

	return false
}

// checkCurlExfil checks curl arguments for sensitive file uploads.
func checkCurlExfil(args []Token) bool {
	for i, t := range args {
		if t.Operator {
			continue
		}
		v := t.Value

		// --data @file, --data-binary @file, --data-raw @file, --data-urlencode @file, -d @file
		if (v == "-d" || v == "--data" || v == "--data-binary" || v == "--data-raw" || v == "--data-urlencode") &&
			i+1 < len(args) && !args[i+1].Operator {
			next := args[i+1].Value
			if strings.HasPrefix(next, "@") && isSensitiveFile(next[1:]) {
				return true
			}
		}
		// -d@file (combined form)
		if strings.HasPrefix(v, "-d@") && isSensitiveFile(v[3:]) {
			return true
		}
		// --data=@file etc.
		for _, prefix := range []string{"--data=@", "--data-binary=@", "--data-raw=@", "--data-urlencode=@"} {
			if strings.HasPrefix(v, prefix) && isSensitiveFile(v[len(prefix):]) {
				return true
			}
		}

		// -F key=@file, --form key=@file
		if (v == "-F" || v == "--form") && i+1 < len(args) && !args[i+1].Operator {
			next := args[i+1].Value
			if _, after, ok := strings.Cut(next, "=@"); ok {
				if isSensitiveFile(after) {
					return true
				}
			}
		}

		// --upload-file file, -T file
		if (v == "--upload-file" || v == "-T") && i+1 < len(args) && !args[i+1].Operator {
			if isSensitiveFile(args[i+1].Value) {
				return true
			}
		}
	}
	return false
}

// checkWgetExfil checks wget arguments for sensitive file uploads.
func checkWgetExfil(args []Token) bool {
	for i, t := range args {
		if t.Operator {
			continue
		}
		v := t.Value

		// --post-file=<file>
		if strings.HasPrefix(v, "--post-file=") && isSensitiveFile(v[len("--post-file="):]) {
			return true
		}
		// --post-file <file>
		if v == "--post-file" && i+1 < len(args) && !args[i+1].Operator {
			if isSensitiveFile(args[i+1].Value) {
				return true
			}
		}
	}
	return false
}

// checkScpExfil checks scp arguments for sensitive file transfers.
// Flags with values (-P, -i, -F, -o, etc.) are skipped. Remaining non-flag
// tokens are checked; the last one is the destination (user@host:...) and is
// skipped, while earlier ones are source files.
func checkScpExfil(args []Token) bool {
	// scp flags that take a following argument value
	scpValueFlags := map[string]bool{
		"-P": true, "-i": true, "-F": true, "-o": true,
		"-c": true, "-l": true, "-S": true, "-J": true,
	}

	var sources []string
	skip := false
	for _, t := range args {
		if t.Operator {
			continue
		}
		if skip {
			skip = false
			continue
		}
		if scpValueFlags[t.Value] {
			skip = true
			continue
		}
		// Skip flag-only args (e.g. -r, -v, -q)
		if strings.HasPrefix(t.Value, "-") {
			continue
		}
		sources = append(sources, t.Value)
	}
	// Last non-flag token is the destination; check everything before it
	if len(sources) < 2 {
		return false
	}
	return slices.ContainsFunc(sources[:len(sources)-1], isSensitiveFile)
}

// CheckNetwork returns true if a network tool is found in command position.
// Command position is the first token or the first token after a pipe/chain operator,
// after stripping wrapper commands (sudo, timeout, etc.) and env var assignments.
func CheckNetwork(input string) bool {
	cmd := ExtractCommand(input)
	if cmd == "" {
		return false
	}
	tokens := ShellTokenize(cmd)

	// Split into segments at pipe/chain operators and check each segment.
	var segment []Token
	check := func() bool {
		unwrapped := SkipWrappers(segment)
		if len(unwrapped) > 0 && !unwrapped[0].Operator {
			base := filepath.Base(unwrapped[0].Value)
			if networkTools[base] {
				return true
			}
		}
		return false
	}

	for _, t := range tokens {
		if t.Operator && (t.Value == "|" || t.Value == "&&" || t.Value == "||" || t.Value == ";") {
			if check() {
				return true
			}
			segment = segment[:0]
			continue
		}
		segment = append(segment, t)
	}
	return check()
}
