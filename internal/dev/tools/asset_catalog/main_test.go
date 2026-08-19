package main

import (
	"path/filepath"
	"testing"
)

// TestExpandHostPathsRequiresMPQDirectory locks down the command's earliest validation failure so later path work
// cannot obscure the required installation input.
func TestExpandHostPathsRequiresMPQDirectory(t *testing.T) {
	_, err := (commandOptions{}).expandHostPaths()
	if err == nil || err.Error() != "-mpq-dir or MPQ_DIRECTORY is required" {
		t.Fatalf("expandHostPaths() error = %v", err)
	}
}

// TestExpandHostPathsExpandsConfiguredPaths verifies every active host-path role uses the shared expansion policy,
// preventing one optional artifact from interpreting environment aliases differently.
func TestExpandHostPathsExpandsConfiguredPaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ASSET_CATALOG_TEST_ROOT", root)

	options := commandOptions{
		mpqDirectory:    "$ASSET_CATALOG_TEST_ROOT/mpqs",
		manifestPath:    "$ASSET_CATALOG_TEST_ROOT/manifests/catalog.json",
		listfilePath:    "$ASSET_CATALOG_TEST_ROOT/listfiles/community.txt",
		outputDirectory: "$ASSET_CATALOG_TEST_ROOT/reports",
		noSheets:        true,
		writeFixture:    "$ASSET_CATALOG_TEST_ROOT/fixtures/observed.json",
	}

	expanded, err := options.expandHostPaths()
	if err != nil {
		t.Fatal(err)
	}

	if want := filepath.Join(root, "mpqs"); expanded.mpqDirectory != want {
		t.Fatalf("mpq directory = %q, want %q", expanded.mpqDirectory, want)
	}

	if want := filepath.Join(root, "manifests", "catalog.json"); expanded.manifestPath != want {
		t.Fatalf("manifest path = %q, want %q", expanded.manifestPath, want)
	}

	if want := filepath.Join(root, "listfiles", "community.txt"); expanded.listfilePath != want {
		t.Fatalf("listfile path = %q, want %q", expanded.listfilePath, want)
	}

	if want := filepath.Join(root, "reports"); expanded.outputDirectory != want {
		t.Fatalf("output directory = %q, want %q", expanded.outputDirectory, want)
	}

	if want := filepath.Join(root, "fixtures", "observed.json"); expanded.writeFixture != want {
		t.Fatalf("write fixture path = %q, want %q", expanded.writeFixture, want)
	}

	if !expanded.noSheets {
		t.Fatal("no-sheets option was not preserved")
	}
}

// TestExpandHostPathsRejectsFixtureModeConflict preserves the mutual-exclusion contract after both paths have passed
// host expansion, ensuring neither mode wins silently.
func TestExpandHostPathsRejectsFixtureModeConflict(t *testing.T) {
	options := commandOptions{
		mpqDirectory:    t.TempDir(),
		outputDirectory: t.TempDir(),
		writeFixture:    "write.json",
		fixturePath:     "verify.json",
	}

	_, err := options.expandHostPaths()
	if err == nil || err.Error() != "-write-fixture and -fixture are mutually exclusive" {
		t.Fatalf("expandHostPaths() error = %v", err)
	}
}
