package nope

import "strings"

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
