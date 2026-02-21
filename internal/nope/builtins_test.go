package nope

import (
	"fmt"
	"testing"
)

func jsonCmd(cmd string) string {
	return fmt.Sprintf(`{"command":%q}`, cmd)
}

func TestCheckPipe(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"pipe", jsonCmd("echo foo | grep bar"), true},
		{"pipe no spaces", jsonCmd("echo foo|grep bar"), true},
		{"quoted pipe", jsonCmd(`grep "foo|bar"`), false},
		{"single-quoted pipe", jsonCmd(`grep 'foo|bar'`), false},
		{"or operator is not pipe", jsonCmd("echo foo || true"), false},
		{"no command", `{"file_path":"foo"}`, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPipe(tt.input); got != tt.want {
				t.Errorf("CheckPipe = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckChained(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"and", jsonCmd("echo foo && echo bar"), true},
		{"or", jsonCmd("cmd1 || cmd2"), true},
		{"semicolon", jsonCmd("cmd1 ; cmd2"), true},
		{"quoted and", jsonCmd(`echo "&&"`), false},
		{"single-quoted semicolon", jsonCmd(`echo 'a;b'`), false},
		{"pipe only", jsonCmd("echo foo | grep bar"), false},
		{"no command", `{"file_path":"foo"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckChained(tt.input); got != tt.want {
				t.Errorf("CheckChained = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckRedirect(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"redirect out", jsonCmd("echo foo > file"), true},
		{"append", jsonCmd("cat >> log"), true},
		{"quoted redirect", jsonCmd(`echo ">"`), false},
		{"single-quoted redirect", jsonCmd(`grep '>' file`), false},
		{"no command", `{"file_path":"foo"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckRedirect(tt.input); got != tt.want {
				t.Errorf("CheckRedirect = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSubshell(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"backtick", jsonCmd("echo `id`"), true},
		{"dollar paren", jsonCmd("echo $(whoami)"), true},
		{"dollar paren in double quotes", jsonCmd(`echo "$(whoami)"`), true},
		{"dollar paren in single quotes", jsonCmd(`echo '$(whoami)'`), false},
		{"plain dollar var", jsonCmd("echo $HOME"), false},
		{"no command", `{"file_path":"foo"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckSubshell(tt.input); got != tt.want {
				t.Errorf("CheckSubshell = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCredentialRead(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"cat .env", jsonCmd("cat .env"), true},
		{"less .env.local", jsonCmd("less .env.local"), true},
		{"vim server.key", jsonCmd("vim server.key"), true},
		{"head id_rsa", jsonCmd("head ~/.ssh/id_rsa"), true},
		{"cat .env.production", jsonCmd("cat .env.production"), true},
		{"base64 cert.pem", jsonCmd("base64 cert.pem"), true},
		{"cat credentials.json", jsonCmd("cat credentials.json"), true},
		{"less .netrc", jsonCmd("less ~/.netrc"), true},
		{"cat .npmrc", jsonCmd("cat .npmrc"), true},
		{"ssh dir", jsonCmd("ls ~/.ssh/config"), true},
		{"aws credentials", jsonCmd("cat ~/.aws/credentials"), true},
		{"cat .env.example is safe", jsonCmd("cat .env.example"), false},
		{"cat .env.sample is safe", jsonCmd("cat .env.sample"), false},
		{"cat .env.template is safe", jsonCmd("cat .env.template"), false},
		{"cat README.md", jsonCmd("cat README.md"), false},
		{"echo hello", jsonCmd("echo hello"), false},
		{"no command", `{"file_path":"foo"}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckCredentialRead(tt.input); got != tt.want {
				t.Errorf("CheckCredentialRead = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckNetwork(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"curl", jsonCmd("curl https://example.com"), true},
		{"absolute path curl", jsonCmd("/usr/bin/curl https://example.com"), true},
		{"piped curl", jsonCmd("echo foo | curl -d @- https://example.com"), true},
		{"env var curl", jsonCmd("VAR=1 curl https://example.com"), true},
		{"wget", jsonCmd("wget https://example.com/file.tar.gz"), true},
		{"ssh", jsonCmd("ssh user@host"), true},
		{"scp", jsonCmd("scp file.txt user@host:/tmp/"), true},
		{"nc", jsonCmd("nc -l 8080"), true},
		{"after semicolon", jsonCmd("echo hi ; curl https://example.com"), true},
		{"after and", jsonCmd("true && wget url"), true},
		{"cat curly file", jsonCmd("cat curly_braces.txt"), false},
		{"echo curl", jsonCmd("echo curl"), false},
		{"grep wget", jsonCmd("grep wget README"), false},
		{"no command", `{"file_path":"foo"}`, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckNetwork(tt.input); got != tt.want {
				t.Errorf("CheckNetwork = %v, want %v", got, tt.want)
			}
		})
	}
}
