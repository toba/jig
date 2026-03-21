#!/usr/bin/env bash
set -euo pipefail

target="$(realpath "$(brew --prefix jig)/bin/jig")"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

go build -ldflags "-X github.com/toba/jig/cmd.ver=dev" -o "$tmp" .
install -m 755 "$tmp" "$target"

echo "Installed to $target"
