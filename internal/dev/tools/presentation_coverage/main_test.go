package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/assets/catalog"
)

// TestWritePresentationCoverageEmitsIndentedJSON protects the command's machine-readable fields as well as the
// human-readable indentation and trailing newline expected by presentation-coverage consumers.
func TestWritePresentationCoverageEmitsIndentedJSON(t *testing.T) {
	source := validCoverageSource()

	var output bytes.Buffer

	if err := writePresentationCoverage(source, &output); err != nil {
		t.Fatal(err)
	}

	assertCoverageReport(t, output.Bytes())
}

// TestWritePresentationCoverageReturnsCatalogReadError verifies that catalog loading remains the first workflow
// phase, which keeps failure ordering and filesystem error identity stable for command callers.
func TestWritePresentationCoverageReturnsCatalogReadError(t *testing.T) {
	err := writePresentationCoverage(fstest.MapFS{}, &bytes.Buffer{})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("write presentation coverage error = %v, want fs.ErrNotExist", err)
	}

	assertPathError(t, err, "manifests/asset-catalog.v1.json")
}

// TestWritePresentationCoverageReturnsCatalogDecodeError ensures malformed input remains a raw JSON error rather than
// being hidden by later fixture or coverage work.
func TestWritePresentationCoverageReturnsCatalogDecodeError(t *testing.T) {
	source := fstest.MapFS{
		"manifests/asset-catalog.v1.json": {Data: []byte(`{"version":`)},
	}

	err := writePresentationCoverage(source, &bytes.Buffer{})

	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatalf("write presentation coverage error = %T(%v), want *json.SyntaxError", err, err)
	}
}

// TestWritePresentationCoverageReturnsCoverageError verifies that coverage construction failures pass through
// unchanged after both inputs load successfully, preserving the command's existing diagnostics.
func TestWritePresentationCoverageReturnsCoverageError(t *testing.T) {
	source := validCoverageSource()
	delete(source, "manifests/presentation.v1.json")

	err := writePresentationCoverage(source, &bytes.Buffer{})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("write presentation coverage error = %v, want fs.ErrNotExist", err)
	}

	assertPathError(t, err, "manifests/presentation.v1.json")
}

// TestWritePresentationCoverageReturnsOutputError confirms that a downstream write failure remains observable after
// coverage succeeds, rather than being mistaken for a successful report.
func TestWritePresentationCoverageReturnsOutputError(t *testing.T) {
	wantErr := errors.New("output rejected")

	err := writePresentationCoverage(validCoverageSource(), rejectingWriter{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("write presentation coverage error = %v, want %v", err, wantErr)
	}
}

// validCoverageSource returns the smallest complete filesystem accepted by the coverage workflow. Each caller owns
// the returned map, so failure-path tests can remove files without leaking state into other tests.
func validCoverageSource() fstest.MapFS {
	return fstest.MapFS{
		"manifests/asset-catalog.v1.json": {
			Data: []byte(`{"version":1,"assets":[{"id":"panel","path":"data/ui/panel.dc6"}]}`),
		},
		"manifests/asset-fixture.v1.json": {
			Data: []byte(`{"version":1,"manifest_version":1,"assets":[{"id":"panel","path":"data/ui/panel.dc6"}]}`),
		},
		"manifests/presentation.v1.json": {Data: []byte(`{"sheet":"data/ui/panel.dc6"}`)},
		"boot.lua":                       {Data: []byte(`local panel = "data/ui/panel.dc6"`)},
		"lua/screen.lua":                 {Data: []byte(`return {}`)},
	}
}

// assertCoverageReport checks both semantic classification and the formatting bytes that distinguish this command
// from a compact JSON encoder. Keeping those assertions together makes the output contract visible in one place.
func assertCoverageReport(t *testing.T, data []byte) {
	t.Helper()

	if !bytes.HasPrefix(data, []byte("{\n  \"version\": 1,")) {
		t.Fatalf("coverage report is not indented:\n%s", data)
	}

	if !bytes.HasSuffix(data, []byte("\n")) {
		t.Fatalf("coverage report has no trailing newline:\n%s", data)
	}

	var coverage assetcatalog.Coverage
	if err := json.Unmarshal(data, &coverage); err != nil {
		t.Fatalf("decode coverage report: %v", err)
	}

	wantPaths := []string{"data/ui/panel.dc6"}

	if coverage.Version != 1 {
		t.Errorf("coverage version = %d, want 1", coverage.Version)
	}

	if !reflect.DeepEqual(coverage.PresentationPaths, wantPaths) {
		t.Errorf("presentation paths = %#v, want %#v", coverage.PresentationPaths, wantPaths)
	}

	if !reflect.DeepEqual(coverage.VerifiedPresentation, wantPaths) {
		t.Errorf("verified presentation = %#v, want %#v", coverage.VerifiedPresentation, wantPaths)
	}

	if len(coverage.Fingerprint) != 64 {
		t.Errorf("fingerprint length = %d, want 64", len(coverage.Fingerprint))
	}
}

// assertPathError verifies which workflow input failed without depending on platform-specific PathError formatting.
// The helper preserves useful test diagnostics while keeping each scenario focused on ordering.
func assertPathError(t *testing.T, err error, wantPath string) {
	t.Helper()

	var pathError *fs.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("error = %T(%v), want *fs.PathError", err, err)
	}

	if pathError.Path != wantPath {
		t.Errorf("error path = %q, want %q", pathError.Path, wantPath)
	}
}

// rejectingWriter returns its configured sentinel on every write so tests can prove output errors retain identity.
// It intentionally accepts no bytes, avoiding partial-output behavior that belongs to encoding/json itself.
type rejectingWriter struct {
	err error
}

// Write rejects the encoded report immediately, providing a deterministic failure at the command's output boundary.
func (writer rejectingWriter) Write(_ []byte) (int, error) {
	return 0, writer.err
}
