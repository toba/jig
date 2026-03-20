package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadPackages(t *testing.T) {
	yaml := `packages: [brew, scoop]
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	pkgs := LoadPackages(doc)
	if len(pkgs) != 2 || pkgs[0] != "brew" || pkgs[1] != "scoop" {
		t.Errorf("packages = %v, want [brew scoop]", pkgs)
	}
}

func TestLoadPackagesMissing(t *testing.T) {
	yaml := `citations: []
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	pkgs := LoadPackages(doc)
	if pkgs != nil {
		t.Errorf("expected nil, got %v", pkgs)
	}
}

func TestHasPackage(t *testing.T) {
	yaml := `packages: [brew, scoop]
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if !HasPackage(doc, "brew") {
		t.Error("expected HasPackage(brew) = true")
	}
	if !HasPackage(doc, "scoop") {
		t.Error("expected HasPackage(scoop) = true")
	}
	if HasPackage(doc, "zed") {
		t.Error("expected HasPackage(zed) = false")
	}
}

func TestAddPackage(t *testing.T) {
	yaml := `citations: []
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := AddPackage(doc, "brew"); err != nil {
		t.Fatal(err)
	}

	// Reload and verify.
	doc2, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	if !HasPackage(doc2, "brew") {
		t.Error("brew not found after AddPackage")
	}

	// Other sections preserved.
	data, err := os.ReadFile(path) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "citations:") {
		t.Error("citations section was lost")
	}
}

func TestAddPackageIdempotent(t *testing.T) {
	yaml := `packages: [brew]
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := AddPackage(doc, "brew"); err != nil {
		t.Fatal(err)
	}

	doc2, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	pkgs := LoadPackages(doc2)
	if len(pkgs) != 1 {
		t.Errorf("expected 1 package, got %v", pkgs)
	}
}

func TestAddPackageSorted(t *testing.T) {
	yaml := `packages: [scoop]
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := AddPackage(doc, "brew"); err != nil {
		t.Fatal(err)
	}

	doc2, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	pkgs := LoadPackages(doc2)
	if len(pkgs) != 2 || pkgs[0] != "brew" || pkgs[1] != "scoop" {
		t.Errorf("expected [brew scoop], got %v", pkgs)
	}
}

func TestLoadZedExtension(t *testing.T) {
	yaml := `zed_extension: toba/gubby
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := LoadZedExtension(doc)
	if ext != "toba/gubby" {
		t.Errorf("zed_extension = %q, want toba/gubby", ext)
	}
}

func TestLoadZedExtensionMissing(t *testing.T) {
	yaml := `citations: []
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	ext := LoadZedExtension(doc)
	if ext != "" {
		t.Errorf("expected empty, got %q", ext)
	}
}

func TestSaveZedExtension(t *testing.T) {
	yaml := `citations: []
`
	path := writeTempConfig(t, yaml)
	doc, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := SaveZedExtension(doc, "toba/gubby"); err != nil {
		t.Fatal(err)
	}

	doc2, err := LoadDocument(path)
	if err != nil {
		t.Fatal(err)
	}
	ext := LoadZedExtension(doc2)
	if ext != "toba/gubby" {
		t.Errorf("zed_extension = %q after save", ext)
	}

	data, err := os.ReadFile(path) //nolint:gosec // test path
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "citations:") {
		t.Error("citations section was lost")
	}
}
