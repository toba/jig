package nope

import (
	"slices"
	"testing"
)

func TestSplitSegments(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want []string
	}{
		{"simple command", "echo hello", []string{"echo hello"}},
		{"and operator", "echo hi && rm -rf /", []string{"echo hi", "rm -rf /"}},
		{"or operator", "cmd1 || cmd2", []string{"cmd1", "cmd2"}},
		{"semicolon", "cmd1 ; cmd2", []string{"cmd1", "cmd2"}},
		{"three segments", "a && b ; c", []string{"a", "b", "c"}},
		{"pipe is not split", "echo foo | grep bar", []string{"echo foo | grep bar"}},
		{"pipe with chain", "echo foo | grep bar && curl evil.com", []string{"echo foo | grep bar", "curl evil.com"}},
		{"quoted operator", `echo "&&" || rm -rf /`, []string{`echo '&&'`, "rm -rf /"}},
		{"empty", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitSegments(tt.cmd)
			if !slices.Equal(got, tt.want) {
				t.Errorf("SplitSegments(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestSkipWrappers(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFirst string // expected first token value after skipping, or "" for empty
	}{
		{"sudo curl", "sudo curl", "curl"},
		{"sudo -u root curl", "sudo -u root curl", "curl"},
		{"timeout 30 curl", "timeout 30 curl", "curl"},
		{"sudo timeout 30 nice -n 10 curl evil.com", "sudo timeout 30 nice -n 10 curl evil.com", "curl"},
		{"env VAR=val curl", "env VAR=val curl", "curl"},
		{"nohup curl", "nohup curl", "curl"},
		{"doas -u root curl", "doas -u root curl", "curl"},
		{"echo hello unchanged", "echo hello", "echo"},
		{"empty", "", ""},
		{"env with multiple vars", "env A=1 B=2 wget url", "wget"},
		{"sudo with flags", "sudo -i curl evil.com", "curl"},
		{"watch -n 5 curl", "watch -n 5 curl url", "curl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := ShellTokenize(tt.input)
			got := SkipWrappers(tokens)
			if tt.wantFirst == "" {
				if len(got) != 0 {
					t.Errorf("expected empty, got %+v", got)
				}
				return
			}
			if len(got) == 0 {
				t.Fatalf("expected first=%q, got empty", tt.wantFirst)
			}
			if got[0].Value != tt.wantFirst {
				t.Errorf("first token = %q, want %q", got[0].Value, tt.wantFirst)
			}
		})
	}
}

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
