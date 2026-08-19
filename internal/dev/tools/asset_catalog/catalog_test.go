package main

import (
	"os"
	"path/filepath"
	"testing"

	assetcatalog "github.com/gravestench/dark-magic/internal/assets/catalog"
)

// TestDefaultManifestComesFromEmbeddedD2Legacy protects the command's no-flag catalog source and its curated asset
// count, making accidental fallback or embedded-data drift visible.
func TestDefaultManifestComesFromEmbeddedD2Legacy(t *testing.T) {
	manifest, err := loadManifest("")
	if err != nil {
		t.Fatal(err)
	}

	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}

	if manifest.Version != 1 || len(manifest.Assets) != 113 {
		t.Fatalf("default manifest = version %d with %d assets", manifest.Version, len(manifest.Assets))
	}
}

// TestLoadManifestReadsExplicitFile verifies an explicit manifest replaces, rather than augments, the embedded source
// while retaining the same JSON decoding contract.
func TestLoadManifestReadsExplicitFile(t *testing.T) {
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifestPath, []byte(`{"version":7,"assets":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	manifest, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	if manifest.Version != 7 || len(manifest.Assets) != 0 {
		t.Fatalf("explicit manifest = version %d with %d assets", manifest.Version, len(manifest.Assets))
	}
}

// TestContactSheetWriterReturnsPortableRelativePath locks down both artifact ownership and the slash-separated path
// stored in report JSON, regardless of the host path separator.
func TestContactSheetWriterReturnsPortableRelativePath(t *testing.T) {
	sheetsDirectory := t.TempDir()
	writeSheet := contactSheetWriter(sheetsDirectory)
	wantData := []byte("sheet bytes")

	relativePath, err := writeSheet("inventory.png", wantData)
	if err != nil {
		t.Fatal(err)
	}

	if relativePath != "contact-sheets/inventory.png" {
		t.Fatalf("relative sheet path = %q", relativePath)
	}

	data, err := os.ReadFile(filepath.Join(sheetsDirectory, "inventory.png"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != string(wantData) {
		t.Fatalf("sheet data = %q, want %q", data, wantData)
	}
}

// TestFoundHypothesisCountLeavesResultOrderingIrrelevant verifies the summary counts only successful observations and
// does not depend on successes being contiguous in the deterministic report.
func TestFoundHypothesisCountLeavesResultOrderingIrrelevant(t *testing.T) {
	report := assetcatalog.Report{Results: []assetcatalog.Result{
		{Found: true},
		{Found: false},
		{Found: true},
	}}

	if got := foundHypothesisCount(report); got != 2 {
		t.Fatalf("foundHypothesisCount() = %d, want 2", got)
	}
}
