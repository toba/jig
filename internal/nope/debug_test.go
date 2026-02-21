package nope

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugLogging(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "nope.log")

	logger := NewDebugLogger(logPath, dir)
	defer logger.Close()

	rules, err := CompileRules([]RuleDef{
		{Name: "test-rule", Pattern: `psql.*(INSERT|UPDATE)`, Message: "blocked"},
	})
	if err != nil {
		t.Fatal(err)
	}

	input := `{"command":"local_psql -c \"INSERT INTO x\""}`
	logger.Log(map[string]any{"event": "check", "tool": "Bash", "input": input})
	msg := CheckRules(rules, "Bash", input, logger)
	if msg == "" {
		t.Fatal("expected block")
	}
	logger.Log(map[string]any{"event": "result", "tool": "Bash", "result": "block", "message": msg})

	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	log := string(data)

	// Verify JSONL output contains expected fields
	if !strings.Contains(log, `"event":"check"`) {
		t.Errorf("log missing check event, got:\n%s", log)
	}
	if !strings.Contains(log, `"event":"rule"`) {
		t.Errorf("log missing rule event, got:\n%s", log)
	}
	if !strings.Contains(log, `"result":"match"`) {
		t.Errorf("log missing rule match, got:\n%s", log)
	}
	if !strings.Contains(log, `"result":"block"`) {
		t.Errorf("log missing block result, got:\n%s", log)
	}
	if !strings.Contains(log, "local_psql") {
		t.Error("log missing input content")
	}

	// Each line should be valid JSON
	for line := range strings.SplitSeq(strings.TrimSpace(log), "\n") {
		if line == "" {
			continue
		}
		if line[0] != '{' {
			t.Errorf("non-JSON line: %s", line)
		}
	}
}

func TestDebugRelativePath(t *testing.T) {
	root := t.TempDir()

	logger := NewDebugLogger("nope.log", root)
	defer logger.Close()

	logger.Log(map[string]any{"event": "test"})
	logger.Close()

	expected := filepath.Join(root, "nope.log")
	data, err := os.ReadFile(expected)
	if err != nil {
		t.Fatalf("log file not created at %s: %v", expected, err)
	}
	if !strings.Contains(string(data), `"event":"test"`) {
		t.Errorf("log missing test event, got: %s", data)
	}
}

func TestDebugAbsolutePathUnchanged(t *testing.T) {
	dir := t.TempDir()
	absPath := filepath.Join(dir, "abs.log")

	logger := NewDebugLogger(absPath, "/some/other/root")
	defer logger.Close()

	logger.Log(map[string]any{"event": "test"})
	logger.Close()

	data, err := os.ReadFile(absPath)
	if err != nil {
		t.Fatalf("log file not created at %s: %v", absPath, err)
	}
	if !strings.Contains(string(data), `"event":"test"`) {
		t.Errorf("log missing test event, got: %s", data)
	}
}

func TestDebugDisabled(t *testing.T) {
	logger := NewDebugLogger("", "")
	if logger != nil {
		t.Error("logger should be nil when debug path is empty")
	}

	// Should not panic
	logger.Log(map[string]any{"event": "test"})
	logger.Close()
}
