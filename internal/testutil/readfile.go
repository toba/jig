package testutil

import (
	"os"
	"testing"
)

func ReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // test helper
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
