package nope

import "testing"

func TestShellTokenize(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []Token
	}{
		{
			name:  "simple command",
			input: "echo hello",
			expect: []Token{
				{Value: "echo"},
				{Value: "hello"},
			},
		},
		{
			name:  "quoted pipe",
			input: `echo "foo|bar"`,
			expect: []Token{
				{Value: "echo"},
				{Value: "foo|bar", Quoted: true},
			},
		},
		{
			name:  "unquoted pipe",
			input: "echo foo | grep bar",
			expect: []Token{
				{Value: "echo"},
				{Value: "foo"},
				{Value: "|", Operator: true},
				{Value: "grep"},
				{Value: "bar"},
			},
		},
		{
			name:  "subshell in double quotes",
			input: `echo "$(whoami)"`,
			expect: []Token{
				{Value: "echo"},
				{Value: "$(", Operator: true},
				{Value: "whoami)", Quoted: true},
			},
		},
		{
			name:  "subshell in single quotes is literal",
			input: `echo '$(whoami)'`,
			expect: []Token{
				{Value: "echo"},
				{Value: "$(whoami)", Quoted: true},
			},
		},
		{
			name:  "escaped double quote",
			input: `echo "hello \"world\""`,
			expect: []Token{
				{Value: "echo"},
				{Value: `hello "world"`, Quoted: true},
			},
		},
		{
			name:  "env var prefix",
			input: "FOO=bar cmd",
			expect: []Token{
				{Value: "FOO=bar"},
				{Value: "cmd"},
			},
		},
		{
			name:  "double ampersand",
			input: "echo foo && echo bar",
			expect: []Token{
				{Value: "echo"},
				{Value: "foo"},
				{Value: "&&", Operator: true},
				{Value: "echo"},
				{Value: "bar"},
			},
		},
		{
			name:  "semicolon",
			input: "echo foo ; echo bar",
			expect: []Token{
				{Value: "echo"},
				{Value: "foo"},
				{Value: ";", Operator: true},
				{Value: "echo"},
				{Value: "bar"},
			},
		},
		{
			name:  "redirect",
			input: "echo foo > file.txt",
			expect: []Token{
				{Value: "echo"},
				{Value: "foo"},
				{Value: ">", Operator: true},
				{Value: "file.txt"},
			},
		},
		{
			name:  "append redirect",
			input: "echo foo >> file.txt",
			expect: []Token{
				{Value: "echo"},
				{Value: "foo"},
				{Value: ">>", Operator: true},
				{Value: "file.txt"},
			},
		},
		{
			name:  "backtick",
			input: "echo `id`",
			expect: []Token{
				{Value: "echo"},
				{Value: "`", Operator: true},
				{Value: "id"},
				{Value: "`", Operator: true},
			},
		},
		{
			name:  "or operator",
			input: "cmd1 || cmd2",
			expect: []Token{
				{Value: "cmd1"},
				{Value: "||", Operator: true},
				{Value: "cmd2"},
			},
		},
		{
			name:  "dollar without paren is not operator",
			input: "echo $HOME",
			expect: []Token{
				{Value: "echo"},
				{Value: "$HOME"},
			},
		},
		{
			name:  "backslash escape",
			input: `echo hello\ world`,
			expect: []Token{
				{Value: "echo"},
				{Value: "hello world"},
			},
		},
		{
			name:  "backtick inside double quotes",
			input: "echo \"`id`\"",
			expect: []Token{
				{Value: "echo"},
				{Value: "`", Operator: true},
				{Value: "id", Quoted: true},
				{Value: "`", Operator: true},
			},
		},
		{
			name:  "backslash-backslash in double quotes",
			input: `echo "a\\b"`,
			expect: []Token{
				{Value: "echo"},
				{Value: `a\b`, Quoted: true},
			},
		},
		{
			name:  "single ampersand (background)",
			input: "echo foo &",
			expect: []Token{
				{Value: "echo"},
				{Value: "foo"},
				{Value: "&"},
			},
		},
		{
			name:  "trailing backslash",
			input: `echo hello\`,
			expect: []Token{
				{Value: "echo"},
				{Value: "hello"},
			},
		},
		{
			name:  "unterminated single quote",
			input: "echo 'hello",
			expect: []Token{
				{Value: "echo"},
				{Value: "hello", Quoted: true},
			},
		},
		{
			name:  "unterminated double quote",
			input: `echo "hello`,
			expect: []Token{
				{Value: "echo"},
				{Value: "hello", Quoted: true},
			},
		},
		{
			name:   "empty string",
			input:  "",
			expect: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShellTokenize(tt.input)
			if len(got) != len(tt.expect) {
				t.Fatalf("len = %d, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.expect), got, tt.expect)
			}
			for i := range got {
				if got[i] != tt.expect[i] {
					t.Errorf("token[%d] = %+v, want %+v", i, got[i], tt.expect[i])
				}
			}
		})
	}
}
