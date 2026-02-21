package zed

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// ExtensionParams holds the inputs for generating extension files.
type ExtensionParams struct {
	ExtID     string   // e.g. "gozer"
	ExtName   string   // e.g. "Gozer"
	Version   string   // e.g. "0.14.0" (no v prefix)
	Desc      string   // extension description
	Org       string   // e.g. "toba"
	ExtRepo   string   // e.g. "gozer" (repo name only)
	LSPRepo   string   // e.g. "toba/go-template-lsp" (owner/name)
	LSPName   string   // e.g. "go-template-lsp" (binary name)
	Languages []string // e.g. ["Go Text Template", "Go HTML Template"]
}

// GenerateExtensionToml produces the extension.toml content.
func GenerateExtensionToml(p ExtensionParams) string {
	langs := make([]string, len(p.Languages))
	for i, l := range p.Languages {
		langs[i] = fmt.Sprintf("%q", l)
	}

	return fmt.Sprintf(`id = "%s"
name = "%s"
version = "%s"
schema_version = 1
authors = ["Jason Abbott"]
description = "%s"
repository = "https://github.com/%s/%s"

[language_servers.%s]
name = "%s"
languages = [%s]
`, p.ExtID, p.ExtName, p.Version, p.Desc,
		p.Org, p.ExtRepo,
		p.ExtID, p.ExtName,
		strings.Join(langs, ", "))
}

// GenerateCargoToml produces the Cargo.toml content.
func GenerateCargoToml(p ExtensionParams) string {
	return fmt.Sprintf(`[package]
name = "%s"
version = "%s"
edition = "2021"
license = "MIT"

[lib]
crate-type = ["cdylib"]

[dependencies]
zed_extension_api = "0.7.0"
`, p.LSPName, p.Version)
}

// GenerateLibRs produces the src/lib.rs content.
func GenerateLibRs(p ExtensionParams) string {
	structName := pascalCase(p.ExtID) + "Extension"

	return fmt.Sprintf(`use std::fs;
use zed_extension_api::{self as zed, LanguageServerId, Result, Worktree};

const GITHUB_REPO: &str = "%s";
const BINARY_NAME: &str = "%s";

struct %s {
    cached_binary_path: Option<String>,
}

impl %s {
    fn language_server_binary_path(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<String> {
        if let Some(path) = &self.cached_binary_path {
            if fs::metadata(path).is_ok() {
                return Ok(path.clone());
            }
        }

        // Check for binary in worktree root (dev extensions)
        let dev_binary_path = format!("{}/{}", worktree.root_path(), BINARY_NAME);
        if fs::metadata(&dev_binary_path).is_ok() {
            self.cached_binary_path = Some(dev_binary_path.clone());
            return Ok(dev_binary_path);
        }

        zed::set_language_server_installation_status(
            language_server_id,
            &zed::LanguageServerInstallationStatus::CheckingForUpdate,
        );

        let release = zed::latest_github_release(
            GITHUB_REPO,
            zed::GithubReleaseOptions {
                require_assets: true,
                pre_release: false,
            },
        )?;

        let (platform, arch) = zed::current_platform();
        let asset_name = format!(
            "%s_{os}_{arch}.{ext}",
            os = match platform {
                zed::Os::Mac => "darwin",
                zed::Os::Linux => "linux",
                zed::Os::Windows => "windows",
            },
            arch = match arch {
                zed::Architecture::Aarch64 => "arm64",
                zed::Architecture::X8664 => "amd64",
                zed::Architecture::X86 => "386",
            },
            ext = match platform {
                zed::Os::Windows => "zip",
                _ => "tar.gz",
            }
        );

        let asset = release
            .assets
            .iter()
            .find(|a| a.name == asset_name)
            .ok_or_else(|| format!("no asset found matching {}", asset_name))?;

        let version_dir = format!("{}-{}", BINARY_NAME, release.version);
        let binary_path = format!(
            "{}/{}{}",
            version_dir,
            BINARY_NAME,
            match platform {
                zed::Os::Windows => ".exe",
                _ => "",
            }
        );

        if fs::metadata(&binary_path).is_err() {
            zed::set_language_server_installation_status(
                language_server_id,
                &zed::LanguageServerInstallationStatus::Downloading,
            );

            zed::download_file(
                &asset.download_url,
                &version_dir,
                match platform {
                    zed::Os::Windows => zed::DownloadedFileType::Zip,
                    _ => zed::DownloadedFileType::GzipTar,
                },
            )
            .map_err(|e| format!("failed to download file: {e}"))?;

            zed::make_file_executable(&binary_path)?;
        }

        self.cached_binary_path = Some(binary_path.clone());
        Ok(binary_path)
    }
}

impl zed::Extension for %s {
    fn new() -> Self {
        Self {
            cached_binary_path: None,
        }
    }

    fn language_server_command(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &Worktree,
    ) -> Result<zed::Command> {
        let binary_path =
            self.language_server_binary_path(language_server_id, worktree)?;

        Ok(zed::Command {
            command: binary_path,
            args: vec![],
            env: worktree.shell_env(),
        })
    }
}

zed::register_extension!(%s);
`, p.LSPRepo, p.LSPName,
		structName, structName,
		p.LSPName,
		structName, structName)
}

// GenerateBumpVersionScript returns the bump-version.sh content (identical across all extensions).
func GenerateBumpVersionScript() string {
	return `#!/usr/bin/env bash
# Bump version in extension.toml, Cargo.toml, and Cargo.lock.
# Usage: scripts/bump-version.sh <version>
# Example: scripts/bump-version.sh 0.14.0
set -euo pipefail

VERSION="${1:?Usage: bump-version.sh <version>}"
VERSION="${VERSION#v}" # strip leading v if present

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

sed -i '' -e "s/^version = \".*\"/version = \"$VERSION\"/" "$REPO_DIR/extension.toml"

# Update only the [package] version, not dependency versions
sed -i '' -e "/^\[package\]/,/^\[/{s/^version = \".*\"/version = \"$VERSION\"/;}" "$REPO_DIR/Cargo.toml"

# Regenerate lockfile if cargo is available
if command -v cargo &>/dev/null; then
  (cd "$REPO_DIR" && cargo generate-lockfile)
fi

echo "Bumped to $VERSION"
`
}

// GenerateBumpVersionWorkflow returns the bump-version.yml workflow content.
func GenerateBumpVersionWorkflow() string {
	return `name: Bump Version

on:
  repository_dispatch:
    types: [bump-version]

permissions:
  contents: write

jobs:
  bump:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: dtolnay/rust-toolchain@stable

      - name: Bump version
        run: |
          VERSION="${{ github.event.client_payload.version }}"
          if [ -z "$VERSION" ]; then
            echo "::error::No version in client_payload"
            exit 1
          fi
          bash scripts/bump-version.sh "$VERSION"

      - name: Commit, tag, and push
        run: |
          VERSION="${{ github.event.client_payload.version }}"
          VERSION="${VERSION#v}"
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add extension.toml Cargo.toml Cargo.lock
          if git diff --cached --quiet; then
            echo "No changes"
            exit 0
          fi
          git commit -m "bump version to $VERSION"
          git tag "v$VERSION"
          git push
          git push origin "v$VERSION"
`
}

// GenerateLicense returns an MIT license with the current year.
func GenerateLicense() string {
	year := time.Now().Year()
	return fmt.Sprintf(`MIT License

Copyright (c) %d Toba

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`, year)
}

// GenerateReadme produces a short README for the extension repo.
func GenerateReadme(p ExtensionParams) string {
	return fmt.Sprintf(`# %s

A Zed editor extension: %s

Powered by [%s](https://github.com/%s).

## Installation

1. Open Zed
2. Go to Extensions (Cmd+Shift+X)
3. Search for "%s"
4. Click Install

The extension automatically downloads the LSP binary for your platform.

**As a Dev Extension:**

1. Clone this repository
2. In Zed, open the command palette (Cmd+Shift+P)
3. Run "zed: install dev extension"
4. Select this directory

## Building

`+"```bash\ncargo build --target wasm32-wasip1\n```"+`

## License

MIT License — see [LICENSE](LICENSE) for details.
`, p.ExtName, p.Desc,
		p.LSPName, p.LSPRepo,
		p.ExtName)
}

// pascalCase converts a hyphenated or lowercase string to PascalCase.
// e.g. "gozer" -> "Gozer", "my-ext" -> "MyExt"
func pascalCase(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '-' || r == '_' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
