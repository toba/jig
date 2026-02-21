package nope

import (
	"os"
	"strings"
	"testing"
)

func TestExitError(t *testing.T) {
	err := ExitError{Code: 2}
	if err.Error() != "exit 2" {
		t.Errorf("ExitError.Error() = %q, want %q", err.Error(), "exit 2")
	}
}

func TestReadHookInput(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		wantTool string
		wantIn   string
		wantErr  bool
	}{
		{
			name:     "valid JSON",
			stdin:    `{"tool_name":"Bash","tool_input":{"command":"echo hi"}}`,
			wantTool: "Bash",
			wantIn:   `{"command":"echo hi"}`,
		},
		{
			name:     "empty stdin",
			stdin:    "",
			wantTool: "Bash",
			wantIn:   "",
		},
		{
			name:     "null tool_input",
			stdin:    `{"tool_name":"Read","tool_input":null}`,
			wantTool: "Read",
			wantIn:   "",
		},
		{
			name:     "missing tool_name defaults to Bash",
			stdin:    `{"tool_input":{"command":"ls"}}`,
			wantTool: "Bash",
			wantIn:   `{"command":"ls"}`,
		},
		{
			name:    "malformed JSON",
			stdin:   `{not valid json`,
			wantErr: true,
		},
		{
			name:    "truncated JSON",
			stdin:   `{"tool_name":"Bash","tool_input":`,
			wantErr: true,
		},
		{
			name:    "plain text",
			stdin:   `just some random text`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace os.Stdin with a pipe containing test data.
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			_, _ = w.WriteString(tt.stdin)
			w.Close()

			oldStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			toolName, input, err := ReadHookInput()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if toolName != tt.wantTool {
				t.Errorf("toolName = %q, want %q", toolName, tt.wantTool)
			}
			// Normalize whitespace for comparison.
			got := strings.TrimSpace(input)
			want := strings.TrimSpace(tt.wantIn)
			if got != want {
				t.Errorf("input = %q, want %q", got, want)
			}
		})
	}
}
