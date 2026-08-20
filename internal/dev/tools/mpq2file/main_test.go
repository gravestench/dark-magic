package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// failingWriter returns a stable injected error so tests can verify write failures are not swallowed.
type failingWriter struct {
	err error
}

// Write implements io.Writer and deliberately fails before accepting any bytes.
func (writer failingWriter) Write(_ []byte) (int, error) {
	return 0, writer.err
}

// TestOptionsHasRequiredPaths locks the validation matrix that distinguishes usage errors from executable work.
func TestOptionsHasRequiredPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options options
		want    bool
	}{
		{
			name: "source and asset",
			options: options{
				sourcePath: "assets.mpq",
				assetPath:  "data/item.txt",
			},
			want: true,
		},
		{
			name: "missing source",
			options: options{
				assetPath: "data/item.txt",
			},
		},
		{
			name: "missing asset",
			options: options{
				sourcePath: "assets.mpq",
			},
		},
		{
			name: "missing both",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := test.options.hasRequiredPaths(); got != test.want {
				t.Fatalf("hasRequiredPaths() = %t, want %t", got, test.want)
			}
		})
	}
}

// TestWriteAssetUsesStdoutWithoutAPath verifies omitted output paths preserve the asset bytes without filesystem work.
func TestWriteAssetUsesStdoutWithoutAPath(t *testing.T) {
	t.Parallel()

	data := []byte("asset bytes")

	var stdout bytes.Buffer

	if err := writeAsset("", data, &stdout); err != nil {
		t.Fatalf("writeAsset() error = %v", err)
	}

	if !bytes.Equal(stdout.Bytes(), data) {
		t.Fatalf("stdout bytes = %q, want %q", stdout.Bytes(), data)
	}
}

// TestWriteAssetCreatesOutputFile verifies explicit destinations receive unchanged bytes and the established mode.
func TestWriteAssetCreatesOutputFile(t *testing.T) {
	t.Parallel()

	data := []byte("asset bytes")
	outputPath := filepath.Join(t.TempDir(), "asset.bin")

	if err := writeAsset(outputPath, data, nil); err != nil {
		t.Fatalf("writeAsset() error = %v", err)
	}

	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	if !bytes.Equal(written, data) {
		t.Fatalf("output bytes = %q, want %q", written, data)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat output: %v", err)
	}

	if got, want := info.Mode().Perm(), os.FileMode(0o644); got != want {
		t.Fatalf("output mode = %v, want %v", got, want)
	}
}

// TestWriteAssetReturnsStdoutError ensures write failures reach fatal unchanged instead of being mistaken for success.
func TestWriteAssetReturnsStdoutError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("stdout unavailable")

	err := writeAsset("", []byte("asset bytes"), failingWriter{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("writeAsset() error = %v, want %v", err, wantErr)
	}
}
