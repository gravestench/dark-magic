package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/gravestench/dark-magic/internal/assets/catalog"
	"github.com/gravestench/dark-magic/internal/content"
)

// main binds the embedded content and process streams to report generation. Keeping exit policy here lets the
// report workflow return original filesystem, JSON, catalog, and output errors unchanged.
func main() {
	if err := writePresentationCoverage(content.D2Legacy(), os.Stdout); err != nil {
		fatal(err)
	}
}

// writePresentationCoverage loads the catalog inputs, derives coverage, and emits one indented JSON document. The
// sequential phases preserve both first-failure ordering and the command's existing output contract.
func writePresentationCoverage(source fs.FS, output io.Writer) error {
	manifest, fixture, err := readCoverageInputs(source)
	if err != nil {
		return err
	}

	coverage, err := assetcatalog.BuildCoverage(source, manifest, fixture)
	if err != nil {
		return err
	}

	return writeCoverageReport(output, coverage)
}

// readCoverageInputs reads the catalog before its fixture so a broken catalog remains the first error reported. It
// deliberately leaves validation to BuildCoverage, matching the command's established responsibility boundary.
func readCoverageInputs(source fs.FS) (assetcatalog.Manifest, assetcatalog.Fixture, error) {
	manifest, err := readManifest[assetcatalog.Manifest](source, "manifests/asset-catalog.v1.json")
	if err != nil {
		return assetcatalog.Manifest{}, assetcatalog.Fixture{}, err
	}

	fixture, err := readManifest[assetcatalog.Fixture](source, "manifests/asset-fixture.v1.json")
	if err != nil {
		return assetcatalog.Manifest{}, assetcatalog.Fixture{}, err
	}

	return manifest, fixture, nil
}

// readManifest decodes one embedded contract without wrapping or validating failures. Returning the original error
// preserves the exact filesystem and JSON diagnostics that the command exposes on standard error.
func readManifest[T any](source fs.FS, name string) (T, error) {
	var value T

	data, err := fs.ReadFile(source, name)
	if err != nil {
		return value, err
	}

	if err := json.Unmarshal(data, &value); err != nil {
		return value, err
	}

	return value, nil
}

// writeCoverageReport uses an encoder so successful output retains its trailing newline. Any writer failure is
// returned unchanged, allowing main to preserve the command's existing error text and exit status.
func writeCoverageReport(output io.Writer, coverage assetcatalog.Coverage) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")

	return encoder.Encode(coverage)
}

// fatal reports an unrecoverable command error and terminates with status 1. Centralizing the process boundary keeps
// lower-level helpers testable and prevents them from unexpectedly bypassing caller cleanup.
func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
