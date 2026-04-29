package cc

import (
	"os"
	"path/filepath"
	"strings"
)

// Home returns the user's home directory.
func Home() (string, error) {
	return os.UserHomeDir()
}

// JigDir returns ~/.jig.
func JigDir() (string, error) {
	h, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".jig"), nil
}

// ConfigPath returns the path to ~/.jig/cc.yaml.
func ConfigPath() (string, error) {
	d, err := JigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "cc.yaml"), nil
}

// HistoryPath returns the path to ~/.jig/cc-history.yaml.
func HistoryPath() (string, error) {
	d, err := JigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "cc-history.yaml"), nil
}

// AliasesDir returns ~/.jig/cc.
func AliasesDir() (string, error) {
	d, err := JigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "cc"), nil
}

// AliasDir returns ~/.jig/cc/<name>.
func AliasDir(name string) (string, error) {
	d, err := AliasesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, name), nil
}

// ExpandHome expands a leading ~ to the user's home directory.
func ExpandHome(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	h, err := Home()
	if err != nil {
		return p
	}
	if p == "~" {
		return h
	}
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(h, p[2:])
	}
	return p
}
