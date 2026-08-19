package assetinspect

import (
	"testing"
	"testing/fstest"
)

// TestInspect verifies that unsupported formats still receive stable metadata;
// this fallback lets callers inventory assets before format support is added.
func TestInspect(t *testing.T) {
	source := fstest.MapFS{"data/example.bin": {Data: []byte("dark magic")}}

	report, err := Inspect(source, "data/example.bin")
	if err != nil {
		t.Fatal(err)
	}

	if report.Type != "bin" || report.Bytes != 10 {
		t.Fatalf("unexpected report: %+v", report)
	}

	if report.SHA256 != "cd75a14415e1556f823b6b21b9247f7c93f7b05fedfca667c305311c7af5d334" {
		t.Fatalf("unexpected digest: %s", report.SHA256)
	}
}

// TestInspectReportsKnownFormatDecodeErrors ensures recognized extensions do
// not silently fall back to metadata when their decoder rejects malformed data.
func TestInspectReportsKnownFormatDecodeErrors(t *testing.T) {
	source := fstest.MapFS{"broken.dc6": {Data: []byte("not a dc6")}}

	if _, err := Inspect(source, "broken.dc6"); err == nil {
		t.Fatal("expected malformed DC6 data to fail decoding")
	}
}
