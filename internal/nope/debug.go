package nope

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DebugLogger writes JSONL debug entries to a file. Nil-safe: calling Log on
// a nil receiver is a no-op.
type DebugLogger struct {
	f *os.File
}

// NewDebugLogger opens a debug log file. Relative paths are resolved against root.
// Returns nil if path is empty.
func NewDebugLogger(path, root string) *DebugLogger {
	if path == "" {
		return nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) //nolint:gosec // debug log path from config
	if err != nil {
		fmt.Fprintf(os.Stderr, "nope: debug log: %v\n", err)
		return nil
	}
	return &DebugLogger{f: f}
}

// Log writes a JSONL entry to the debug log.
func (d *DebugLogger) Log(fields map[string]any) {
	if d == nil {
		return
	}
	fields["ts"] = time.Now().Format(time.RFC3339Nano)
	data, err := json.Marshal(fields)
	if err != nil {
		return
	}
	d.f.Write(data)       //nolint:errcheck // debug log, best-effort
	d.f.WriteString("\n") //nolint:errcheck // debug log, best-effort
}

// Close closes the debug log file.
func (d *DebugLogger) Close() {
	if d == nil {
		return
	}
	d.f.Close() //nolint:errcheck // debug log, best-effort
}
