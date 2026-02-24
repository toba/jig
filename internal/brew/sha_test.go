package brew

import (
	"errors"
	"slices"
	"testing"
)

func TestResolveSHA256Sidecar(t *testing.T) {
	orig := ghRelease
	defer func() { ghRelease = orig }()

	ghRelease = func(args ...string) ([]byte, error) {
		// Match the sidecar download pattern.
		if slices.Contains(args, "tool_darwin_arm64.tar.gz.sha256") {
			return []byte("abc123def456  tool_darwin_arm64.tar.gz\n"), nil
		}
		return nil, errors.New("not found")
	}

	sha, err := ResolveSHA256("org/tool", "v1.0.0", "tool_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if sha != "abc123def456" {
		t.Errorf("SHA = %q, want abc123def456", sha)
	}
}

func TestResolveSHA256Checksums(t *testing.T) {
	orig := ghRelease
	defer func() { ghRelease = orig }()

	calls := 0
	ghRelease = func(args ...string) ([]byte, error) {
		calls++
		for _, a := range args {
			if a == "tool_darwin_arm64.tar.gz.sha256" {
				return nil, errors.New("not found")
			}
			if a == "checksums.txt" {
				return []byte("deadbeef01  tool_linux_amd64.tar.gz\nfeedface02  tool_darwin_arm64.tar.gz\n"), nil
			}
		}
		return nil, errors.New("not found")
	}

	sha, err := ResolveSHA256("org/tool", "v1.0.0", "tool_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if sha != "feedface02" {
		t.Errorf("SHA = %q, want feedface02", sha)
	}
}

func TestResolveSHA256AllFail(t *testing.T) {
	orig := ghRelease
	defer func() { ghRelease = orig }()

	ghRelease = func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	}

	_, err := ResolveSHA256("org/tool", "v1.0.0", "tool_darwin_arm64.tar.gz")
	if err == nil {
		t.Error("expected error when all strategies fail")
	}
}
