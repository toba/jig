// Package constants defines shared string constants used across multiple
// internal packages to avoid raw-string duplication and circular imports.
package constants

const (
	// ConfigFileName is the name of the project config file.
	ConfigFileName = ".jig.yaml"

	// DefaultBranch is the fallback branch name when none is specified.
	DefaultBranch = "main"

	// GoreleaserYAML is the primary goreleaser config filename.
	GoreleaserYAML = ".goreleaser.yaml"

	// GoreleaserYML is the alternate goreleaser config filename.
	GoreleaserYML = ".goreleaser.yml"
)
