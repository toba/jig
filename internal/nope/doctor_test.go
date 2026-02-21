package nope

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTobaYAML(t *testing.T, dir string, content []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".jig.yaml"), content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunDoctor(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, dir string)
		wantExit int
	}{
		{
			name:     "no config file",
			setup:    func(t *testing.T, dir string) {},
			wantExit: 1,
		},
		{
			name: "no nope section",
			setup: func(t *testing.T, dir string) {
				setupTobaYAML(t, dir, []byte("citations:\n  sources: []\n"))
			},
			wantExit: 1,
		},
		{
			name: "valid config with no rules",
			setup: func(t *testing.T, dir string) {
				setupTobaYAML(t, dir, []byte("nope:\n  rules: []\n"))
			},
			wantExit: 0,
		},
		{
			name: "valid config with bad pattern",
			setup: func(t *testing.T, dir string) {
				cfg := "nope:\n  rules:\n    - name: bad\n      pattern: '[invalid'\n      message: nope\n"
				setupTobaYAML(t, dir, []byte(cfg))
			},
			wantExit: 1,
		},
		{
			name: "valid config with good rules",
			setup: func(t *testing.T, dir string) {
				cfg := "nope:\n  rules:\n    - name: test\n      pattern: 'rm -rf'\n      message: no\n"
				setupTobaYAML(t, dir, []byte(cfg))
			},
			wantExit: 0,
		},
		{
			name: "starter config from init",
			setup: func(t *testing.T, dir string) {
				origDir, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
				RunInit()
				if err := os.Chdir(origDir); err != nil {
					t.Fatal(err)
				}
			},
			wantExit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tc.setup(t, dir)

			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Chdir(origDir); err != nil {
					t.Logf("warning: could not restore dir: %v", err)
				}
			}()

			got := RunDoctor()
			if got != tc.wantExit {
				t.Errorf("RunDoctor() = %d, want %d", got, tc.wantExit)
			}
		})
	}
}
