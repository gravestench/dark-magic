package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestStandardMPQPriority protects the newest-to-oldest archive order because it determines duplicate asset lookup.
// An exact comparison also catches accidental omission of a recognized Diablo II archive.
func TestStandardMPQPriority(t *testing.T) {
	expected := []string{
		"patch_d2.mpq", "d2exp.mpq", "d2data.mpq", "d2char.mpq",
		"d2music.mpq", "d2sfx.mpq", "d2speech.mpq", "d2video.mpq",
		"d2xmusic.mpq", "d2xtalk.mpq", "d2xvideo.mpq",
	}

	if !reflect.DeepEqual(standardMPQPriority, expected) {
		t.Fatalf("standardMPQPriority = %v, want %v", standardMPQPriority, expected)
	}
}

// TestResolvePlayableBIKReturnsDirectFile verifies that file mode neither opens nor extracts the supplied path.
// A callable no-op cleanup keeps ownership uniform for main without touching the original file.
func TestResolvePlayableBIKReturnsDirectFile(t *testing.T) {
	options := commandOptions{
		fileName:   "movie.bik",
		sourceName: "ignored-source",
		assetName:  "ignored-asset",
	}

	fileName, cleanup, err := resolvePlayableBIK(options)
	if err != nil {
		t.Fatalf("resolvePlayableBIK() error = %v", err)
	}

	cleanup()

	if fileName != options.fileName {
		t.Fatalf("file name = %q, want %q", fileName, options.fileName)
	}
}

// TestExtractBIKCopiesDirectoryAssetAndCleansUp exercises the complete temporary-file ownership contract.
// Registering cleanup before assertions ensures a failed test cannot leak the extracted host file.
func TestExtractBIKCopiesDirectoryAssetAndCleansUp(t *testing.T) {
	sourceDirectory := t.TempDir()
	assetName := "data/local/video/intro.bik"
	expected := []byte("BIK fixture bytes")
	writeSourceAsset(t, sourceDirectory, assetName, expected)

	fileName, cleanup, err := extractBIK(sourceDirectory, assetName)
	if err != nil {
		t.Fatalf("extractBIK() error = %v", err)
	}

	t.Cleanup(cleanup)

	if !strings.HasSuffix(fileName, ".bik") {
		t.Fatalf("temporary file = %q, want .bik suffix", fileName)
	}

	actual, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatalf("read temporary file: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("temporary contents = %q, want %q", actual, expected)
	}

	cleanup()

	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Fatalf("stat cleaned temporary file error = %v, want not-exist", err)
	}
}

// TestExtractBIKReportsMissingAsset preserves the contextual read error and safe cleanup returned after extraction
// fails.
// Calling cleanup also verifies callers do not need a separate ownership branch for failed resolutions.
func TestExtractBIKReportsMissingAsset(t *testing.T) {
	const assetName = "data/local/video/missing.bik"

	_, cleanup, err := extractBIK(t.TempDir(), assetName)
	if err == nil {
		t.Fatal("extractBIK() error = nil, want missing asset error")
	}

	cleanup()

	if !strings.HasPrefix(err.Error(), `reading "`+assetName+`": `) {
		t.Fatalf("error = %q, want contextual asset read prefix", err)
	}
}

// writeSourceAsset creates a slash-delimited io/fs fixture beneath a host directory.
// The helper keeps directory construction and file ownership inside t.TempDir while preserving exact fixture bytes.
func writeSourceAsset(t *testing.T, sourceDirectory, assetName string, data []byte) {
	t.Helper()

	fileName := filepath.Join(sourceDirectory, filepath.FromSlash(assetName))
	if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
		t.Fatalf("create source directories: %v", err)
	}

	if err := os.WriteFile(fileName, data, 0o600); err != nil {
		t.Fatalf("write source asset: %v", err)
	}

	// Confirm the fixture is addressable through io/fs before exercising the production source adapter.
	if _, err := fs.Stat(os.DirFS(sourceDirectory), assetName); err != nil {
		t.Fatalf("stat source fixture: %v", err)
	}
}
