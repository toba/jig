package commit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMatchesGitignorePattern(t *testing.T) {
	shouldMatch := []string{
		"debug.log",
		"tmp/data.tmp",
		".cache",
		"app.pyc",
		"mod.pyo",
		"main.o",
		"lib.a",
		"lib.so",
		"lib.dylib",
		".env",
		"config.env.local",
		".DS_Store",
		"file.swp",
		"file.swo",
		"node_modules/foo",
		"__pycache__/mod",
		".venv/bin",
		"venv/lib",
		".idea/workspace.xml",
		"dist/bundle.js",
		"build/output",
		"coverage/lcov",
		"project.coverage",
		"credentials.json",
		"secrets.yaml",
		"server.key",
		"cert.pem",
		"keystore.p12",
		"DerivedData/Build",
		"project.xcuserstate",
		"xcuserdata/jason",
		"file.moved-aside",
		"Pods/AFNetworking",
	}

	for _, f := range shouldMatch {
		if !matchesGitignorePattern(f) {
			t.Errorf("expected %q to match a gitignore pattern", f)
		}
	}

	shouldNotMatch := []string{
		"main.go",
		"README.md",
		"internal/config/config.go",
		"go.mod",
		".gitignore",
		"cmd/commit.go",
		"Makefile",
		"scripts/test.sh",
		"docs/guide.md",
		"envoy.yaml",
		"keychain.go",
	}

	for _, f := range shouldNotMatch {
		if matchesGitignorePattern(f) {
			t.Errorf("expected %q NOT to match a gitignore pattern", f)
		}
	}
}

// setupGitRepo creates a temp dir with an initialized git repo and
// changes the working directory to it. Returns a cleanup function.
func setupGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	run("git", "init")
	run("git", "config", "user.email", "test@test.com")
	run("git", "config", "user.name", "Test")

	// Create an initial commit so we have a HEAD.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "-A")
	run("git", "commit", "-m", "initial commit")

	// Change to the temp repo directory.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	return dir
}

func TestStageAll(t *testing.T) {
	dir := setupGitRepo(t)

	// Create a new file to be staged.
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := StageAll()
	if err != nil {
		t.Fatalf("StageAll() error: %v", err)
	}

	if !strings.Contains(status, "new.txt") {
		t.Errorf("StageAll() status = %q, expected it to contain 'new.txt'", status)
	}
}

func TestStageAllClean(t *testing.T) {
	setupGitRepo(t)

	// No changes — should return empty.
	status, err := StageAll()
	if err != nil {
		t.Fatalf("StageAll() error: %v", err)
	}
	if status != "" {
		t.Errorf("StageAll() on clean repo = %q, want empty", status)
	}
}

func TestDiff(t *testing.T) {
	dir := setupGitRepo(t)

	// Stage a change.
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "file.txt")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	diff, err := Diff()
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}

	if !strings.Contains(diff, "file.txt") {
		t.Errorf("Diff() = %q, expected it to contain 'file.txt'", diff)
	}
	if !strings.Contains(diff, "+content") {
		t.Errorf("Diff() = %q, expected it to contain '+content'", diff)
	}
}

func TestDiffEmpty(t *testing.T) {
	setupGitRepo(t)

	diff, err := Diff()
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if diff != "" {
		t.Errorf("Diff() on clean repo = %q, want empty", diff)
	}
}

func TestLatestTag(t *testing.T) {
	t.Run("no tags", func(t *testing.T) {
		setupGitRepo(t)

		tag, err := LatestTag()
		if err != nil {
			t.Fatalf("LatestTag() error: %v", err)
		}
		if tag != "" {
			t.Errorf("LatestTag() = %q, want empty", tag)
		}
	})

	t.Run("with tags", func(t *testing.T) {
		dir := setupGitRepo(t)

		for _, v := range []string{"v0.1.0", "v0.2.0", "v1.0.0"} {
			cmd := exec.Command("git", "tag", v)
			cmd.Dir = dir
			if err := cmd.Run(); err != nil {
				t.Fatalf("git tag %s: %v", v, err)
			}
		}

		tag, err := LatestTag()
		if err != nil {
			t.Fatalf("LatestTag() error: %v", err)
		}
		if tag != "v1.0.0" {
			t.Errorf("LatestTag() = %q, want %q", tag, "v1.0.0")
		}
	})
}

func TestRecentCommits(t *testing.T) {
	t.Run("empty tag returns recent commits", func(t *testing.T) {
		setupGitRepo(t)

		log, err := RecentCommits("")
		if err != nil {
			t.Fatalf("RecentCommits(\"\") error: %v", err)
		}
		if !strings.Contains(log, "initial commit") {
			t.Errorf("RecentCommits(\"\") = %q, expected 'initial commit'", log)
		}
	})

	t.Run("with tag and new commits", func(t *testing.T) {
		dir := setupGitRepo(t)

		// Tag the initial commit.
		cmd := exec.Command("git", "tag", "v1.0.0")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		// Make a new commit after the tag.
		if err := os.WriteFile(filepath.Join(dir, "after-tag.txt"), []byte("data\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, args := range [][]string{
			{"git", "add", "-A"},
			{"git", "commit", "-m", "post-tag commit"},
		} {
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"GIT_AUTHOR_NAME=Test",
				"GIT_AUTHOR_EMAIL=test@test.com",
				"GIT_COMMITTER_NAME=Test",
				"GIT_COMMITTER_EMAIL=test@test.com",
			)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("command %v: %v\n%s", args, err, out)
			}
		}

		log, err := RecentCommits("v1.0.0")
		if err != nil {
			t.Fatalf("RecentCommits(\"v1.0.0\") error: %v", err)
		}
		if !strings.Contains(log, "post-tag commit") {
			t.Errorf("RecentCommits(\"v1.0.0\") = %q, expected 'post-tag commit'", log)
		}
		if strings.Contains(log, "initial commit") {
			t.Errorf("RecentCommits(\"v1.0.0\") should not contain 'initial commit'")
		}
	})

	t.Run("with tag but no new commits falls back to recent", func(t *testing.T) {
		dir := setupGitRepo(t)

		// Tag HEAD — no commits after tag.
		cmd := exec.Command("git", "tag", "v1.0.0")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		log, err := RecentCommits("v1.0.0")
		if err != nil {
			t.Fatalf("RecentCommits(\"v1.0.0\") error: %v", err)
		}
		if !strings.Contains(log, "initial commit") {
			t.Errorf("RecentCommits(\"v1.0.0\") = %q, expected fallback to recent commits", log)
		}
	})
}

func TestHasStagedChanges(t *testing.T) {
	t.Run("no staged changes", func(t *testing.T) {
		setupGitRepo(t)

		has, err := HasStagedChanges()
		if err != nil {
			t.Fatalf("HasStagedChanges() error: %v", err)
		}
		if has {
			t.Error("HasStagedChanges() = true, want false on clean repo")
		}
	})

	t.Run("with staged changes", func(t *testing.T) {
		dir := setupGitRepo(t)

		if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("data\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "add", "new.txt")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}

		has, err := HasStagedChanges()
		if err != nil {
			t.Fatalf("HasStagedChanges() error: %v", err)
		}
		if !has {
			t.Error("HasStagedChanges() = false, want true with staged file")
		}
	})
}

func TestCommit(t *testing.T) {
	dir := setupGitRepo(t)

	// Stage a change.
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	if err := Commit("test commit message"); err != nil {
		t.Fatalf("Commit() error: %v", err)
	}

	// Verify the commit was made.
	out, err := exec.Command("git", "log", "--oneline", "-1").Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "test commit message") {
		t.Errorf("git log = %q, expected 'test commit message'", string(out))
	}
}

func TestCommitNoChanges(t *testing.T) {
	setupGitRepo(t)

	// Commit with nothing staged should fail.
	err := Commit("empty commit")
	if err == nil {
		t.Fatal("Commit() with nothing staged should return an error")
	}
	if !strings.Contains(err.Error(), "git commit") {
		t.Errorf("Commit() error = %q, expected it to contain 'git commit'", err.Error())
	}
}

func TestTag(t *testing.T) {
	setupGitRepo(t)

	if err := Tag("v2.0.0"); err != nil {
		t.Fatalf("Tag() error: %v", err)
	}

	// Verify the tag was created.
	out, err := exec.Command("git", "tag", "-l", "v2.0.0").Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "v2.0.0" {
		t.Errorf("git tag output = %q, want 'v2.0.0'", strings.TrimSpace(string(out)))
	}
}

func TestTagDuplicate(t *testing.T) {
	setupGitRepo(t)

	if err := Tag("v1.0.0"); err != nil {
		t.Fatal(err)
	}

	// Creating the same tag again should fail.
	err := Tag("v1.0.0")
	if err == nil {
		t.Fatal("Tag() with duplicate tag should return an error")
	}
	if !strings.Contains(err.Error(), "v1.0.0") {
		t.Errorf("Tag() error = %q, expected it to contain 'v1.0.0'", err.Error())
	}
}

func TestStatus(t *testing.T) {
	t.Run("clean repo", func(t *testing.T) {
		setupGitRepo(t)

		status, err := Status()
		if err != nil {
			t.Fatalf("Status() error: %v", err)
		}
		if status != "" {
			t.Errorf("Status() = %q, want empty for clean repo", status)
		}
	})

	t.Run("dirty repo", func(t *testing.T) {
		dir := setupGitRepo(t)

		if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("data\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		status, err := Status()
		if err != nil {
			t.Fatalf("Status() error: %v", err)
		}
		if !strings.Contains(status, "untracked.txt") {
			t.Errorf("Status() = %q, expected 'untracked.txt'", status)
		}
	})
}

func TestGitignoreCandidates(t *testing.T) {
	t.Run("no candidates", func(t *testing.T) {
		dir := setupGitRepo(t)

		// Create an untracked file that does NOT match gitignore patterns.
		if err := os.WriteFile(filepath.Join(dir, "normal.go"), []byte("package main\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		candidates, err := GitignoreCandidates()
		if err != nil {
			t.Fatalf("GitignoreCandidates() error: %v", err)
		}
		if len(candidates) != 0 {
			t.Errorf("GitignoreCandidates() = %v, want empty", candidates)
		}
	})

	t.Run("with candidates", func(t *testing.T) {
		dir := setupGitRepo(t)

		// Create untracked files that match gitignore patterns.
		if err := os.WriteFile(filepath.Join(dir, "debug.log"), []byte("log\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=x\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		candidates, err := GitignoreCandidates()
		if err != nil {
			t.Fatalf("GitignoreCandidates() error: %v", err)
		}
		if len(candidates) != 2 {
			t.Errorf("GitignoreCandidates() = %v, want 2 candidates", candidates)
		}
	})

	t.Run("clean repo", func(t *testing.T) {
		setupGitRepo(t)

		candidates, err := GitignoreCandidates()
		if err != nil {
			t.Fatalf("GitignoreCandidates() error: %v", err)
		}
		if candidates != nil {
			t.Errorf("GitignoreCandidates() = %v, want nil", candidates)
		}
	})
}

// setupGitRepoWithRemote creates a local repo with a bare remote.
// Returns the working dir and a run helper.
func setupGitRepoWithRemote(t *testing.T) (dir string, run func(args ...string)) {
	t.Helper()

	bare := t.TempDir()
	dir = t.TempDir()

	runIn := func(d string, args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = d
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	// Create bare remote.
	runIn(bare, "git", "init", "--bare")

	// Create working repo.
	runIn(dir, "git", "init")
	runIn(dir, "git", "config", "user.email", "test@test.com")
	runIn(dir, "git", "config", "user.name", "Test")
	runIn(dir, "git", "remote", "add", "origin", bare)

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runIn(dir, "git", "add", "-A")
	runIn(dir, "git", "commit", "-m", "initial commit")
	runIn(dir, "git", "push", "-u", "origin", "HEAD")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	return dir, func(args ...string) { runIn(dir, args...) }
}

func TestUnpushedVersionTags(t *testing.T) {
	t.Run("no unpushed tags", func(t *testing.T) {
		setupGitRepoWithRemote(t)

		tags, err := unpushedVersionTags()
		if err != nil {
			t.Fatalf("unpushedVersionTags() error: %v", err)
		}
		if len(tags) != 0 {
			t.Errorf("unpushedVersionTags() = %v, want empty", tags)
		}
	})

	t.Run("all local tags unpushed", func(t *testing.T) {
		_, run := setupGitRepoWithRemote(t)

		run("git", "tag", "v1.0.0")
		run("git", "tag", "v0.9.0")
		run("git", "tag", "v1.1.0")

		tags, err := unpushedVersionTags()
		if err != nil {
			t.Fatalf("unpushedVersionTags() error: %v", err)
		}

		// Should be sorted by version: v0.9.0, v1.0.0, v1.1.0.
		want := []string{"v0.9.0", "v1.0.0", "v1.1.0"}
		if len(tags) != len(want) {
			t.Fatalf("unpushedVersionTags() = %v, want %v", tags, want)
		}
		for i, tag := range tags {
			if tag != want[i] {
				t.Errorf("tags[%d] = %q, want %q", i, tag, want[i])
			}
		}
	})

	t.Run("some tags already pushed", func(t *testing.T) {
		_, run := setupGitRepoWithRemote(t)

		run("git", "tag", "v1.0.0")
		run("git", "push", "origin", "v1.0.0")
		run("git", "tag", "v1.1.0")
		run("git", "tag", "v1.2.0")

		tags, err := unpushedVersionTags()
		if err != nil {
			t.Fatalf("unpushedVersionTags() error: %v", err)
		}

		want := []string{"v1.1.0", "v1.2.0"}
		if len(tags) != len(want) {
			t.Fatalf("unpushedVersionTags() = %v, want %v", tags, want)
		}
		for i, tag := range tags {
			if tag != want[i] {
				t.Errorf("tags[%d] = %q, want %q", i, tag, want[i])
			}
		}
	})

	t.Run("ignores non-version tags", func(t *testing.T) {
		_, run := setupGitRepoWithRemote(t)

		run("git", "tag", "release-1")
		run("git", "tag", "v2.0.0")

		tags, err := unpushedVersionTags()
		if err != nil {
			t.Fatalf("unpushedVersionTags() error: %v", err)
		}

		// Only v* tags.
		if len(tags) != 1 || tags[0] != "v2.0.0" {
			t.Errorf("unpushedVersionTags() = %v, want [v2.0.0]", tags)
		}
	})
}

func TestPushOrdersTags(t *testing.T) {
	_, run := setupGitRepoWithRemote(t)

	// Create multiple unpushed tags.
	run("git", "tag", "v1.0.0")
	run("git", "tag", "v0.5.0")
	run("git", "tag", "v1.1.0")

	if err := Push(); err != nil {
		t.Fatalf("Push() error: %v", err)
	}

	// Verify all tags made it to the remote.
	out, err := exec.Command("git", "ls-remote", "--tags", "origin").Output()
	if err != nil {
		t.Fatalf("ls-remote: %v", err)
	}
	remote := string(out)
	for _, tag := range []string{"v0.5.0", "v1.0.0", "v1.1.0"} {
		if !strings.Contains(remote, "refs/tags/"+tag) {
			t.Errorf("remote missing tag %s", tag)
		}
	}
}

func TestGitignorePatternsCompiled(t *testing.T) {
	// Verify all patterns in gitignorePatterns are non-nil (compiled successfully).
	for i, re := range gitignorePatterns {
		if re == nil {
			t.Errorf("gitignorePatterns[%d] is nil", i)
		}
	}
}

func TestMatchesGitignorePatternEdgeCases(t *testing.T) {
	tests := []struct {
		path  string
		match bool
	}{
		// Nested paths
		{"some/deep/dir/debug.log", true},
		{"project/node_modules/package/index.js", true},
		{"src/__pycache__/module.pyc", true},
		{"app/DerivedData/Build/Products", true},

		// Partial name matches that should NOT match
		{"catalog.go", false},
		{"logger.go", false},
		{"env_config.go", false},
		{"building.go", false},

		// Empty path
		{"", false},
	}

	for _, tt := range tests {
		got := matchesGitignorePattern(tt.path)
		if got != tt.match {
			t.Errorf("matchesGitignorePattern(%q) = %v, want %v", tt.path, got, tt.match)
		}
	}
}
