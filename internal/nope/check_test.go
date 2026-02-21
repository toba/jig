package nope

import (
	"strings"
	"testing"
)

func loadTestRules(t *testing.T) []CompiledRule {
	t.Helper()
	cfg, err := LoadConfig("testdata/nope.yaml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	rules, err := CompileRules(cfg.Rules)
	if err != nil {
		t.Fatalf("CompileRules: %v", err)
	}
	return rules
}

func TestCheckRulesBlocked(t *testing.T) {
	rules := loadTestRules(t)

	tests := []struct {
		name  string
		input string
		want  string // substring of expected block message
	}{
		// git checkout/switch
		{
			name:  "git checkout",
			input: `{"command":"git checkout feature-branch"}`,
			want:  "git checkout/switch",
		},
		{
			name:  "git switch",
			input: `{"command":"git switch main"}`,
			want:  "git checkout/switch",
		},
		// git commit/push
		{
			name:  "git commit",
			input: `{"command":"git commit -m 'fix'"}`,
			want:  "git commit/push",
		},
		{
			name:  "git push",
			input: `{"command":"git push origin main"}`,
			want:  "git commit/push",
		},
		// dropdb
		{
			name:  "dropdb",
			input: `{"command":"dropdb pacer"}`,
			want:  "dropdb is not allowed",
		},
		// prod migrations
		{
			name:  "prod-query migrations",
			input: `{"command":"./scripts/prod-query.sh -f migrations/001.sql"}`,
			want:  "running migrations against production",
		},
		// psql apply migrations
		{
			name:  "psql -f migrations",
			input: `{"command":"psql -f internal/database/migrations/001.sql"}`,
			want:  "applying migrations directly",
		},
		// psql DML/DDL
		{
			name:  "psql UPDATE",
			input: `{"command":"psql -c 'UPDATE users SET active=true'"}`,
			want:  "DML/DDL via psql",
		},
		{
			name:  "psql DROP",
			input: `{"command":"psql -c 'DROP TABLE users'"}`,
			want:  "DML/DDL via psql",
		},
		{
			name:  "psql SET search_path",
			input: `{"command":"psql -c 'SET search_path TO public'"}`,
			want:  "DML/DDL via psql",
		},
		// pgmigrate
		{
			name:  "pgmigrate",
			input: `{"command":"pgmigrate apply"}`,
			want:  "pgmigrate not allowed",
		},
		// prod-query DDL/DML
		{
			name:  "prod-query DELETE",
			input: `{"command":"./scripts/prod-query.sh -c 'DELETE FROM users'"}`,
			want:  "DDL/DML against production",
		},
		// beans archive
		{
			name:  "beans archive",
			input: `{"command":"beans archive"}`,
			want:  "beans archive not allowed",
		},
		// bq mutations
		{
			name:  "bq mk",
			input: `{"command":"bq mk --dataset myproject:mydataset"}`,
			want:  "bq infrastructure mutations",
		},
		{
			name:  "bq rm",
			input: `{"command":"bq rm myproject:mydataset.mytable"}`,
			want:  "bq infrastructure mutations",
		},
		// gcloud mutations
		{
			name:  "gcloud run deploy",
			input: `{"command":"gcloud run deploy my-service --image gcr.io/foo/bar"}`,
			want:  "gcloud infrastructure mutations",
		},
		{
			name:  "gcloud sql create",
			input: `{"command":"gcloud sql create my-instance"}`,
			want:  "gcloud infrastructure mutations",
		},
		// tofu
		{
			name:  "tofu apply",
			input: `{"command":"tofu apply"}`,
			want:  "tofu apply/destroy/import",
		},
		{
			name:  "tofu destroy",
			input: `{"command":"tofu destroy"}`,
			want:  "tofu apply/destroy/import",
		},
		// curl bearer
		{
			name:  "curl with bearer token",
			input: `{"command":"curl -H 'Authorization: Bearer sk-123' https://api.example.com"}`,
			want:  "bearer tokens not allowed",
		},
		// curl PMS APIs
		{
			name:  "curl guesty",
			input: `{"command":"curl https://open-api.guesty.com/v1/listings"}`,
			want:  "direct curl to PMS APIs",
		},
		{
			name:  "curl hostaway",
			input: `{"command":"curl https://api.hostaway.com/v1/listings"}`,
			want:  "direct curl to PMS APIs",
		},
		// psql oauth_tokens
		{
			name:  "psql oauth_tokens",
			input: `{"command":"psql -c 'SELECT * FROM oauth_tokens'"}`,
			want:  "oauth_tokens access not allowed",
		},
		// deploy.sh
		{
			name:  "deploy.sh",
			input: `{"command":"./scripts/deploy.sh prod"}`,
			want:  "deploy.sh not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := CheckRules(rules, "Bash", tt.input, nil)
			if msg == "" {
				t.Fatal("expected block, got allow")
			}
			if !strings.Contains(msg, tt.want) {
				t.Errorf("message = %q, want substring %q", msg, tt.want)
			}
		})
	}
}

func TestCheckRulesAllowed(t *testing.T) {
	rules := loadTestRules(t)

	tests := []struct {
		name  string
		input string
	}{
		{"git status", `{"command":"git status"}`},
		{"git diff", `{"command":"git diff"}`},
		{"git log", `{"command":"git log --oneline -5"}`},
		{"go test", `{"command":"go test ./..."}`},
		{"go build", `{"command":"go build ./..."}`},
		{"single line command", `{"command":"echo hello"}`},
		{"psql SELECT", `{"command":"psql -c 'SELECT * FROM users'"}`},
		{"bq query", `{"command":"bq query 'SELECT 1'"}`},
		{"bq load data", `{"command":"bq load mydataset.mytable data.csv"}`},
		{"gcloud list", `{"command":"gcloud run services list"}`},
		{"tofu plan", `{"command":"tofu plan"}`},
		{"beans list", `{"command":"beans list --json"}`},
		{"empty input", ``},
		{"non-command json", `{"file_path":"/tmp/foo.txt"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := CheckRules(rules, "Bash", tt.input, nil)
			if msg != "" {
				t.Errorf("expected allow, got block: %s", msg)
			}
		})
	}
}

func TestCheckRulesToolFiltering(t *testing.T) {
	tests := []struct {
		name     string
		rules    []RuleDef
		toolName string
		input    string
		wantMsg  string // "" means allow
	}{
		{
			name:     "default tools scoped to Bash — blocks Bash",
			rules:    []RuleDef{{Name: "r", Pattern: "secret", Message: "blocked"}},
			toolName: "Bash",
			input:    `{"command":"echo secret"}`,
			wantMsg:  "blocked",
		},
		{
			name:     "default tools scoped to Bash — skips Write",
			rules:    []RuleDef{{Name: "r", Pattern: "secret", Message: "blocked"}},
			toolName: "Write",
			input:    `{"file_path":"secret.txt"}`,
			wantMsg:  "",
		},
		{
			name:     "rule scoped to Write — blocks Write",
			rules:    []RuleDef{{Name: "r", Pattern: `\.env`, Tools: []string{"Write"}, Message: "no env"}},
			toolName: "Write",
			input:    `{"file_path":".env"}`,
			wantMsg:  "no env",
		},
		{
			name:     "rule scoped to Write — skips Bash",
			rules:    []RuleDef{{Name: "r", Pattern: `\.env`, Tools: []string{"Write"}, Message: "no env"}},
			toolName: "Bash",
			input:    `{"command":"cat .env"}`,
			wantMsg:  "",
		},
		{
			name:     "rule scoped to multiple tools",
			rules:    []RuleDef{{Name: "r", Pattern: `\.env`, Tools: []string{"Write", "Edit"}, Message: "no env"}},
			toolName: "Edit",
			input:    `{"file_path":".env"}`,
			wantMsg:  "no env",
		},
		{
			name:     "wildcard * matches any tool",
			rules:    []RuleDef{{Name: "r", Pattern: "password", Tools: []string{"*"}, Message: "no passwords"}},
			toolName: "Read",
			input:    `{"file_path":"password.txt"}`,
			wantMsg:  "no passwords",
		},
		{
			name:     "wildcard * matches Bash too",
			rules:    []RuleDef{{Name: "r", Pattern: "password", Tools: []string{"*"}, Message: "no passwords"}},
			toolName: "Bash",
			input:    `{"command":"cat password.txt"}`,
			wantMsg:  "no passwords",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := CompileRules(tt.rules)
			if err != nil {
				t.Fatalf("CompileRules: %v", err)
			}
			msg := CheckRules(compiled, tt.toolName, tt.input, nil)
			if msg != tt.wantMsg {
				t.Errorf("CheckRules = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestCheckRulesMultilinePatternMatch(t *testing.T) {
	// Patterns with .* must match across newlines (e.g. shell line continuations).
	tests := []struct {
		name    string
		pattern string
		input   string
		wantHit bool
	}{
		{
			name:    "psql DDL across newline",
			pattern: `psql.*(UPDATE|INSERT|DELETE|ALTER|DROP|CREATE|TRUNCATE|SET\s+search_path)`,
			input:   `{"command":"prod_psql -c \\\n        \"SET search_path TO core;\n         SELECT 1\""}`,
			wantHit: true,
		},
		{
			name:    "psql DDL single line still matches",
			pattern: `psql.*(UPDATE|INSERT|DELETE|ALTER|DROP|CREATE|TRUNCATE|SET\s+search_path)`,
			input:   `{"command":"psql -c 'SET search_path TO core'"}`,
			wantHit: true,
		},
		{
			name:    "no match when keyword absent",
			pattern: `psql.*(UPDATE|INSERT|DELETE|ALTER|DROP|CREATE|TRUNCATE|SET\s+search_path)`,
			input:   `{"command":"prod_psql -c \\\n        \"SELECT 1\""}`,
			wantHit: false,
		},
		{
			name:    "curl bearer across newline",
			pattern: `curl.*Authorization.*Bearer`,
			input:   `{"command":"curl \\\n  -H 'Authorization: Bearer sk-123' \\\n  https://api.example.com"}`,
			wantHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := CompileRules([]RuleDef{
				{Name: "test", Pattern: tt.pattern, Message: "blocked"},
			})
			if err != nil {
				t.Fatalf("CompileRules: %v", err)
			}
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

func TestCheckRulesNonBashToolFromTestdata(t *testing.T) {
	rules := loadTestRules(t)

	// The testdata nope.yaml has a Write-scoped rule for .env files.
	// It should block Write tool but not Bash tool.
	envInput := `{"file_path":"/tmp/.env"}`

	msg := CheckRules(rules, "Write", envInput, nil)
	if msg == "" {
		t.Error("expected block for Write to .env, got allow")
	}

	msg = CheckRules(rules, "Bash", envInput, nil)
	if msg != "" {
		t.Errorf("expected allow for Bash with .env file_path, got block: %s", msg)
	}
}
