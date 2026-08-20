package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestCommandOptionsHasRequiredPaths pins validation to both paths so partial invocations remain usage errors.
func TestCommandOptionsHasRequiredPaths(t *testing.T) {
	tests := []struct {
		name      string
		options   commandOptions
		wantValid bool
	}{
		{name: "both paths", options: commandOptions{sourcePath: "source", assetPath: "asset"}, wantValid: true},
		{name: "missing source", options: commandOptions{assetPath: "asset"}},
		{name: "missing asset", options: commandOptions{sourcePath: "source"}},
		{name: "both missing", options: commandOptions{}},
	}

	for _, test := range tests {
		if got := test.options.hasRequiredPaths(); got != test.wantValid {
			t.Errorf("%s: hasRequiredPaths() = %t, want %t", test.name, got, test.wantValid)
		}
	}
}

// TestInspectAssetWritesExactJSONReport guards the report format and newline consumed by command-line clients.
func TestInspectAssetWritesExactJSONReport(t *testing.T) {
	const (
		assetName = "sample.bin"
		assetData = "asset bytes"
	)

	sourcePath := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourcePath, assetName), []byte(assetData), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	options := commandOptions{sourcePath: sourcePath, assetPath: assetName}
	if err := inspectAsset(options, &output); err != nil {
		t.Fatal(err)
	}

	digest := sha256.Sum256([]byte(assetData))

	want := fmt.Sprintf(
		"{\"path\":%q,\"type\":\"bin\",\"bytes\":%d,\"sha256\":\"%x\"}\n",
		assetName,
		len(assetData),
		digest,
	)
	if got := output.String(); got != want {
		t.Fatalf("report = %q, want %q", got, want)
	}
}

// TestInspectAssetWritesReportBeforePreviewFailure preserves the observable phase order for failed preview requests.
func TestInspectAssetWritesReportBeforePreviewFailure(t *testing.T) {
	const (
		assetName = "sample.bin"
		wantError = `PNG preview is not supported for "bin" assets`
	)

	sourcePath := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourcePath, assetName), []byte("asset bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer

	options := commandOptions{
		sourcePath:  sourcePath,
		assetPath:   assetName,
		previewPath: filepath.Join(t.TempDir(), "preview.png"),
	}

	err := inspectAsset(options, &output)
	if err == nil {
		t.Fatal("inspectAsset() error = nil, want unsupported preview error")
	}

	if err.Error() != wantError {
		t.Fatalf("inspectAsset() error = %q, want %q", err, wantError)
	}

	if output.Len() == 0 {
		t.Fatal("inspectAsset() wrote no report before preview failure")
	}
}

// TestCommandOptionsUsesTexturedDS1Preview pins the compatibility rule for case-insensitive DS1 extensions.
func TestCommandOptionsUsesTexturedDS1Preview(t *testing.T) {
	tests := []struct {
		name         string
		options      commandOptions
		wantTextured bool
	}{
		{
			name:         "DS1 with tile dependencies",
			options:      commandOptions{assetPath: "map.DS1", dt1Paths: "floor.dt1"},
			wantTextured: true,
		},
		{
			name:    "DS1 without tile dependencies",
			options: commandOptions{assetPath: "map.ds1"},
		},
		{
			name:    "other asset with tile dependencies",
			options: commandOptions{assetPath: "sprite.dc6", dt1Paths: "floor.dt1"},
		},
	}

	for _, test := range tests {
		if got := test.options.usesTexturedDS1Preview(); got != test.wantTextured {
			t.Errorf("%s: usesTexturedDS1Preview() = %t, want %t", test.name, got, test.wantTextured)
		}
	}
}
