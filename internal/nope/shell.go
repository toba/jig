package nope

import (
	"path/filepath"
	"strings"
)

// SplitSegments splits a command string on chain operators (&&, ||, ;)
// into independent command segments. Pipe (|) is NOT a split point since
// piped commands form a single pipeline. Returns one trimmed string per segment.
func SplitSegments(cmd string) []string {
	tokens := ShellTokenize(cmd)
	var segments []string
	var cur []string

	for _, t := range tokens {
		if t.Operator && (t.Value == "&&" || t.Value == "||" || t.Value == ";") {
			if len(cur) > 0 {
				segments = append(segments, strings.Join(cur, " "))
				cur = cur[:0]
			}
			continue
		}
		// Reconstruct the token for the segment string.
		if t.Operator {
			cur = append(cur, t.Value)
		} else if t.Quoted {
			// Re-quote so the reconstructed segment is valid shell.
			cur = append(cur, "'"+strings.ReplaceAll(t.Value, "'", "'\\''")+"'")
		} else {
			cur = append(cur, t.Value)
		}
	}
	if len(cur) > 0 {
		segments = append(segments, strings.Join(cur, " "))
	}
	return segments
}

// wrapperDef describes a command wrapper that precedes the real command.
type wrapperDef struct {
	argFlags       map[string]bool // flags that consume the next token (e.g. -u)
	positionalArgs int             // non-flag args before the real command (e.g. timeout takes 1 for duration)
	skipEnvVars    bool            // skip VAR=val tokens (for env command)
}

// wrappers maps command basenames to their wrapper definitions.
var wrappers = map[string]wrapperDef{
	"sudo":       {argFlags: map[string]bool{"-u": true, "-g": true, "-C": true, "-D": true, "-R": true, "-T": true}},
	"doas":       {argFlags: map[string]bool{"-u": true}},
	"timeout":    {positionalArgs: 1},
	"env":        {skipEnvVars: true},
	"nice":       {argFlags: map[string]bool{"-n": true}},
	"nohup":      {},
	"time":       {},
	"watch":      {argFlags: map[string]bool{"-n": true, "-d": true, "-t": true}},
	"caffeinate": {argFlags: map[string]bool{"-w": true}},
	"ionice":     {argFlags: map[string]bool{"-c": true, "-n": true, "-p": true}},
	"strace":     {argFlags: map[string]bool{"-e": true, "-o": true, "-p": true, "-s": true, "-P": true}},
	"xargs":      {argFlags: map[string]bool{"-I": true, "-L": true, "-n": true, "-P": true, "-s": true}},
}

// SkipWrappers strips wrapper command tokens and their flags/args from the
// front of a token slice, returning tokens starting at the real command.
// It recurses to handle chained wrappers like "sudo timeout 30 curl".
func SkipWrappers(tokens []Token) []Token {
	// Skip leading operators (defensive)
	for len(tokens) > 0 && tokens[0].Operator {
		tokens = tokens[1:]
	}

	// Skip env var assignments (FOO=bar)
	for len(tokens) > 0 && !tokens[0].Operator && !tokens[0].Quoted && strings.Contains(tokens[0].Value, "=") {
		tokens = tokens[1:]
	}

	if len(tokens) == 0 {
		return tokens
	}

	base := filepath.Base(tokens[0].Value)
	w, ok := wrappers[base]
	if !ok {
		return tokens
	}

	// Consume the wrapper command itself
	tokens = tokens[1:]

	// Consume wrapper flags and their arguments
	for len(tokens) > 0 && !tokens[0].Operator {
		v := tokens[0].Value
		if !strings.HasPrefix(v, "-") {
			break
		}
		if w.argFlags[v] {
			// Flag consumes the next token
			tokens = tokens[1:]
			if len(tokens) > 0 && !tokens[0].Operator {
				tokens = tokens[1:]
			}
		} else {
			// Simple flag (no argument)
			tokens = tokens[1:]
		}
	}

	// Consume positional args (e.g. duration for timeout)
	for i := 0; i < w.positionalArgs && len(tokens) > 0 && !tokens[0].Operator; i++ {
		tokens = tokens[1:]
	}

	// Skip env vars for env command
	if w.skipEnvVars {
		for len(tokens) > 0 && !tokens[0].Operator && !tokens[0].Quoted && strings.Contains(tokens[0].Value, "=") {
			tokens = tokens[1:]
		}
	}

	// Recurse to handle chained wrappers
	return SkipWrappers(tokens)
}

// Token represents a shell token with quoting and operator metadata.
type Token struct {
	Value    string
	Quoted   bool // was any part of this token quoted
	Operator bool // is this an operator (|, &&, ||, ;, >, >>, $(, `)
}

// ShellTokenize splits a shell command into tokens respecting quoting rules.
// Single quotes: everything literal until closing '.
// Double quotes: \" and \\ are escapes; $( and ` remain operators (bash expands them).
// Backslash (unquoted): escapes next character.
// Unterminated quotes: treat rest as quoted (safe default).
func ShellTokenize(cmd string) []Token {
	var tokens []Token
	var cur []byte
	curQuoted := false

	flush := func() {
		if len(cur) > 0 {
			tokens = append(tokens, Token{Value: string(cur), Quoted: curQuoted})
			cur = cur[:0]
			curQuoted = false
		}
	}

	emitOp := func(op string) {
		flush()
		tokens = append(tokens, Token{Value: op, Operator: true})
	}

	i := 0
	for i < len(cmd) {
		ch := cmd[i]

		switch ch {
		case '\'': // single quote — everything literal until closing '
			i++
			curQuoted = true
			for i < len(cmd) && cmd[i] != '\'' {
				cur = append(cur, cmd[i])
				i++
			}
			if i < len(cmd) {
				i++ // skip closing '
			}

		case '"': // double quote — escapes \" and \\; $( and ` are operators
			i++
			curQuoted = true
			for i < len(cmd) && cmd[i] != '"' {
				if cmd[i] == '\\' && i+1 < len(cmd) && (cmd[i+1] == '"' || cmd[i+1] == '\\') {
					cur = append(cur, cmd[i+1])
					i += 2
					continue
				}
				if cmd[i] == '$' && i+1 < len(cmd) && cmd[i+1] == '(' {
					flush()
					emitOp("$(")
					i += 2
					continue
				}
				if cmd[i] == '`' {
					flush()
					emitOp("`")
					i++
					continue
				}
				cur = append(cur, cmd[i])
				i++
			}
			if i < len(cmd) {
				i++ // skip closing "
			}

		case '\\': // backslash escape
			if i+1 < len(cmd) {
				cur = append(cur, cmd[i+1])
				i += 2
			} else {
				i++
			}

		case '|':
			if i+1 < len(cmd) && cmd[i+1] == '|' {
				emitOp("||")
				i += 2
			} else {
				emitOp("|")
				i++
			}

		case '&':
			if i+1 < len(cmd) && cmd[i+1] == '&' {
				emitOp("&&")
				i += 2
			} else {
				cur = append(cur, ch)
				i++
			}

		case ';':
			emitOp(";")
			i++

		case '>':
			if i+1 < len(cmd) && cmd[i+1] == '>' {
				emitOp(">>")
				i += 2
			} else {
				emitOp(">")
				i++
			}

		case '$':
			if i+1 < len(cmd) && cmd[i+1] == '(' {
				emitOp("$(")
				i += 2
			} else {
				cur = append(cur, ch)
				i++
			}

		case '`':
			emitOp("`")
			i++

		case ' ', '\t':
			flush()
			i++

		default:
			cur = append(cur, ch)
			i++
		}
	}
	flush()
	return tokens
}
