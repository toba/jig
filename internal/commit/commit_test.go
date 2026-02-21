package commit

import "testing"

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
