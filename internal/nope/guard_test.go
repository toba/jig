package nope

import "testing"

func TestExitError(t *testing.T) {
	err := ExitError{Code: 2}
	if err.Error() != "exit 2" {
		t.Errorf("ExitError.Error() = %q, want %q", err.Error(), "exit 2")
	}
}
