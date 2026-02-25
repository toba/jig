package scoop

import (
	"encoding/json"
	"strings"
)

// ManifestParams holds the inputs needed to generate a Scoop manifest.
type ManifestParams struct {
	Tool        string // binary name, e.g. "jig"
	Desc        string // one-line description
	Homepage    string // e.g. "https://github.com/toba/jig"
	License     string // e.g. "Apache-2.0"
	Tag         string // e.g. "v1.2.3"
	Repo        string // e.g. "toba/jig"
	SHA256AMD64 string // hex-encoded sha256 for windows amd64 zip
	SHA256ARM64 string // hex-encoded sha256 for windows arm64 zip
}

// manifest is the JSON structure for a Scoop app manifest.
type manifest struct {
	Version      string                        `json:"version"`
	Description  string                        `json:"description"`
	Homepage     string                        `json:"homepage"`
	License      string                        `json:"license"`
	Architecture map[string]manifestArchConfig `json:"architecture"`
	Bin          []string                      `json:"bin"`
	Checkver     manifestCheckver              `json:"checkver"`
	Autoupdate   manifestAutoupdate            `json:"autoupdate"`
}

type manifestArchConfig struct {
	URL  string `json:"url"`
	Hash string `json:"hash"`
}

type manifestAutoArchConfig struct {
	URL string `json:"url"`
}

type manifestCheckver struct {
	Github string `json:"github"`
}

type manifestAutoupdate struct {
	Architecture map[string]manifestAutoArchConfig `json:"architecture"`
}

// GenerateManifest produces the JSON content for a Scoop app manifest.
func GenerateManifest(p ManifestParams) string {
	version := strings.TrimPrefix(p.Tag, "v")
	baseURL := "https://github.com/" + p.Repo + "/releases/download/" + p.Tag + "/"
	autoURL := "https://github.com/" + p.Repo + "/releases/download/v$version/"

	m := manifest{
		Version:     version,
		Description: p.Desc,
		Homepage:    p.Homepage,
		License:     p.License,
		Architecture: map[string]manifestArchConfig{
			"64bit": {
				URL:  baseURL + p.Tool + "_windows_amd64.zip",
				Hash: p.SHA256AMD64,
			},
			"arm64": {
				URL:  baseURL + p.Tool + "_windows_arm64.zip",
				Hash: p.SHA256ARM64,
			},
		},
		Bin: []string{p.Tool + ".exe"},
		Checkver: manifestCheckver{
			Github: "https://github.com/" + p.Repo,
		},
		Autoupdate: manifestAutoupdate{
			Architecture: map[string]manifestAutoArchConfig{
				"64bit": {
					URL: autoURL + p.Tool + "_windows_amd64.zip",
				},
				"arm64": {
					URL: autoURL + p.Tool + "_windows_arm64.zip",
				},
			},
		},
	}

	data, _ := json.MarshalIndent(m, "", "    ")
	return string(data) + "\n"
}
