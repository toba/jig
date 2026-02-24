package github

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// mockGH replaces ghExec for testing and restores it on cleanup.
func mockGH(t *testing.T, fn func(args ...string) ([]byte, error)) {
	t.Helper()
	orig := ghExec
	ghExec = fn
	t.Cleanup(func() { ghExec = orig })
}

func TestCommitNormalize(t *testing.T) {
	c := Commit{
		SHA: "abc123",
		RawCommit: rawCommit{
			Message: "Fix bug\n\nDetailed description",
			RawAuthor: rawDate{
				Name: "Test User",
				Date: time.Date(2026, 2, 18, 22, 8, 27, 0, time.UTC),
			},
		},
		RawAuthor: &rawAuthor{Login: "testuser"},
	}
	c.Normalize()

	if c.Message != "Fix bug" {
		t.Errorf("message = %q, want %q", c.Message, "Fix bug")
	}
	if c.Author != "testuser" {
		t.Errorf("author = %q, want %q", c.Author, "testuser")
	}
	if !c.Date.Equal(time.Date(2026, 2, 18, 22, 8, 27, 0, time.UTC)) {
		t.Errorf("date = %v, want 2026-02-18T22:08:27Z", c.Date)
	}
}

func TestCommitNormalizeFallbackAuthor(t *testing.T) {
	c := Commit{
		RawCommit: rawCommit{
			Message:   "Single line",
			RawAuthor: rawDate{Name: "Fallback Name"},
		},
	}
	c.Normalize()

	if c.Author != "Fallback Name" {
		t.Errorf("author = %q, want %q", c.Author, "Fallback Name")
	}
	if c.Message != "Single line" {
		t.Errorf("message = %q, want %q", c.Message, "Single line")
	}
}

func TestCommitNormalizeNilRawAuthor(t *testing.T) {
	c := Commit{
		RawCommit: rawCommit{
			Message:   "msg",
			RawAuthor: rawDate{Name: "Name"},
		},
		RawAuthor: nil,
	}
	c.Normalize()
	if c.Author != "Name" {
		t.Errorf("author = %q, want %q", c.Author, "Name")
	}
}

func TestCommitNormalizeEmptyLogin(t *testing.T) {
	c := Commit{
		RawCommit: rawCommit{
			Message:   "msg",
			RawAuthor: rawDate{Name: "Name"},
		},
		RawAuthor: &rawAuthor{Login: ""},
	}
	c.Normalize()
	if c.Author != "Name" {
		t.Errorf("author = %q, want %q", c.Author, "Name")
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"hello\nworld", "hello"},
		{"", ""},
		{"first\nsecond\nthird", "first"},
		{"\n", ""},
		{"no newline", "no newline"},
	}
	for _, tt := range tests {
		if got := firstLine(tt.in); got != tt.want {
			t.Errorf("firstLine(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestGetCommits_Success(t *testing.T) {
	payload := []map[string]any{
		{
			"sha": "abc123",
			"commit": map[string]any{
				"message": "First commit\n\nBody",
				"author": map[string]any{
					"name": "Author",
					"date": "2026-01-15T10:00:00Z",
				},
			},
			"author": map[string]any{
				"login": "octocat",
			},
		},
		{
			"sha": "def456",
			"commit": map[string]any{
				"message": "Second commit",
				"author": map[string]any{
					"name": "Author2",
					"date": "2026-01-16T10:00:00Z",
				},
			},
		},
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	commits, err := c.GetCommits("owner/repo", "main", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].SHA != "abc123" {
		t.Errorf("expected SHA abc123, got %q", commits[0].SHA)
	}
	if commits[0].Message != "First commit" {
		t.Errorf("expected message 'First commit', got %q", commits[0].Message)
	}
	if commits[0].Author != "octocat" {
		t.Errorf("expected author 'octocat', got %q", commits[0].Author)
	}
	if commits[1].Author != "Author2" {
		t.Errorf("expected author 'Author2', got %q", commits[1].Author)
	}
}

func TestGetCommits_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("gh api: Not Found")
	})

	c := NewClient()
	_, err := c.GetCommits("owner/repo", "main", 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetCommits_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("not json"), nil
	})

	c := NewClient()
	_, err := c.GetCommits("owner/repo", "main", 10)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetCommitsSince_Success(t *testing.T) {
	payload := []map[string]any{
		{
			"sha": "abc123",
			"commit": map[string]any{
				"message": "Recent commit",
				"author": map[string]any{
					"name": "Author",
					"date": "2026-02-01T10:00:00Z",
				},
			},
		},
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	commits, err := c.GetCommitsSince("owner/repo", "main", "2026-01-01T00:00:00Z", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
}

func TestGetCommitsSince_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("gh api error")
	})

	c := NewClient()
	_, err := c.GetCommitsSince("owner/repo", "main", "2026-01-01T00:00:00Z", 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetCommitsSince_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("{bad}"), nil
	})

	c := NewClient()
	_, err := c.GetCommitsSince("owner/repo", "main", "2026-01-01T00:00:00Z", 10)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCompare_Success(t *testing.T) {
	payload := map[string]any{
		"status":        "ahead",
		"ahead_by":      2,
		"total_commits": 2,
		"commits": []map[string]any{
			{
				"sha": "aaa",
				"commit": map[string]any{
					"message": "commit 1",
					"author":  map[string]any{"name": "A", "date": "2026-01-15T10:00:00Z"},
				},
			},
			{
				"sha": "bbb",
				"commit": map[string]any{
					"message": "commit 2",
					"author":  map[string]any{"name": "B", "date": "2026-01-16T10:00:00Z"},
				},
			},
		},
		"files": []map[string]any{
			{"filename": "file.go", "status": "modified"},
		},
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	resp, err := c.Compare("owner/repo", "v1.0", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "ahead" {
		t.Errorf("expected status 'ahead', got %q", resp.Status)
	}
	if resp.AheadBy != 2 {
		t.Errorf("expected ahead_by 2, got %d", resp.AheadBy)
	}
	if len(resp.Commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(resp.Commits))
	}
	if len(resp.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(resp.Files))
	}
	if resp.Commits[0].Message != "commit 1" {
		t.Errorf("expected commit message 'commit 1', got %q", resp.Commits[0].Message)
	}
}

func TestCompare_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	})

	c := NewClient()
	_, err := c.Compare("owner/repo", "v1.0", "main")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompare_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("bad"), nil
	})

	c := NewClient()
	_, err := c.Compare("owner/repo", "v1.0", "main")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetCommitDetail_Success(t *testing.T) {
	payload := map[string]any{
		"sha": "abc123",
		"commit": map[string]any{
			"message": "Detailed commit\n\nWith body",
			"author": map[string]any{
				"name": "Author",
				"date": "2026-01-15T10:00:00Z",
			},
		},
		"author": map[string]any{"login": "dev"},
		"files": []map[string]any{
			{"filename": "main.go", "status": "modified"},
		},
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	commit, err := c.GetCommitDetail("owner/repo", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commit.SHA != "abc123" {
		t.Errorf("expected SHA abc123, got %q", commit.SHA)
	}
	if commit.Message != "Detailed commit" {
		t.Errorf("expected message 'Detailed commit', got %q", commit.Message)
	}
	if commit.Author != "dev" {
		t.Errorf("expected author 'dev', got %q", commit.Author)
	}
	if len(commit.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(commit.Files))
	}
}

func TestGetCommitDetail_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	})

	c := NewClient()
	_, err := c.GetCommitDetail("owner/repo", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetCommitDetail_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("bad"), nil
	})

	c := NewClient()
	_, err := c.GetCommitDetail("owner/repo", "abc123")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetHeadSHA_Success(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("abc123def456\n"), nil
	})

	c := NewClient()
	sha, err := c.GetHeadSHA("owner/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != "abc123def456" {
		t.Errorf("expected 'abc123def456', got %q", sha)
	}
}

func TestGetHeadSHA_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	})

	c := NewClient()
	_, err := c.GetHeadSHA("owner/repo", "main")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRepo_Success(t *testing.T) {
	payload := map[string]any{
		"default_branch": "main",
		"description":    "A test repo",
		"language":       "Go",
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	info, err := c.GetRepo("owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.DefaultBranch != "main" {
		t.Errorf("expected default branch 'main', got %q", info.DefaultBranch)
	}
	if info.Description != "A test repo" {
		t.Errorf("expected description 'A test repo', got %q", info.Description)
	}
	if info.Language != "Go" {
		t.Errorf("expected language 'Go', got %q", info.Language)
	}
}

func TestGetRepo_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	})

	c := NewClient()
	_, err := c.GetRepo("owner/repo")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRepo_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("bad"), nil
	})

	c := NewClient()
	_, err := c.GetRepo("owner/repo")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetTree_Success(t *testing.T) {
	payload := map[string]any{
		"sha":       "tree123",
		"truncated": false,
		"tree": []map[string]any{
			{"path": "README.md", "type": "blob"},
			{"path": "src", "type": "tree"},
			{"path": "src/main.go", "type": "blob"},
		},
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	resp, err := c.GetTree("owner/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SHA != "tree123" {
		t.Errorf("expected SHA 'tree123', got %q", resp.SHA)
	}
	if len(resp.Tree) != 3 {
		t.Errorf("expected 3 tree entries, got %d", len(resp.Tree))
	}
	if resp.Tree[0].Path != "README.md" {
		t.Errorf("expected path 'README.md', got %q", resp.Tree[0].Path)
	}
	if resp.Tree[0].Type != "blob" {
		t.Errorf("expected type 'blob', got %q", resp.Tree[0].Type)
	}
	if resp.Truncated {
		t.Error("expected truncated=false")
	}
}

func TestGetTree_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	})

	c := NewClient()
	_, err := c.GetTree("owner/repo", "main")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetTree_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("bad"), nil
	})

	c := NewClient()
	_, err := c.GetTree("owner/repo", "main")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGetLicense_Success(t *testing.T) {
	payload := map[string]any{
		"license": map[string]any{
			"key":     "mit",
			"name":    "MIT License",
			"spdx_id": "MIT",
		},
	}
	data, _ := json.Marshal(payload)
	mockGH(t, func(args ...string) ([]byte, error) {
		return data, nil
	})

	c := NewClient()
	info, err := c.GetLicense("owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.License.Key != "mit" {
		t.Errorf("expected key 'mit', got %q", info.License.Key)
	}
	if info.License.Name != "MIT License" {
		t.Errorf("expected name 'MIT License', got %q", info.License.Name)
	}
	if info.License.SPDXID != "MIT" {
		t.Errorf("expected SPDX ID 'MIT', got %q", info.License.SPDXID)
	}
}

func TestGetLicense_Error(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	})

	c := NewClient()
	_, err := c.GetLicense("owner/repo")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetLicense_InvalidJSON(t *testing.T) {
	mockGH(t, func(args ...string) ([]byte, error) {
		return []byte("bad"), nil
	})

	c := NewClient()
	_, err := c.GetLicense("owner/repo")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGH_ArgsPassedCorrectly(t *testing.T) {
	var capturedArgs []string
	mockGH(t, func(args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte("{}"), nil
	})

	c := NewClient()
	_, _ = c.GetRepo("owner/repo")

	if len(capturedArgs) != 2 {
		t.Fatalf("expected 2 args, got %d: %v", len(capturedArgs), capturedArgs)
	}
	if capturedArgs[0] != "api" {
		t.Errorf("expected first arg 'api', got %q", capturedArgs[0])
	}
	if capturedArgs[1] != "repos/owner/repo" {
		t.Errorf("expected second arg 'repos/owner/repo', got %q", capturedArgs[1])
	}
}

func TestGetHeadSHA_ArgsIncludeJQ(t *testing.T) {
	var capturedArgs []string
	mockGH(t, func(args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte("sha123\n"), nil
	})

	c := NewClient()
	_, _ = c.GetHeadSHA("owner/repo", "main")

	if len(capturedArgs) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(capturedArgs), capturedArgs)
	}
	if capturedArgs[2] != "--jq" {
		t.Errorf("expected arg '--jq', got %q", capturedArgs[2])
	}
	if capturedArgs[3] != ".sha" {
		t.Errorf("expected arg '.sha', got %q", capturedArgs[3])
	}
}
