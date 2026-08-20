package assetdecode

import (
	"testing"
	"testing/fstest"
)

// TestPaletteRejectsTruncatedData verifies that a short BGR table fails before
// callers can receive a partially initialized 256-color palette.
func TestPaletteRejectsTruncatedData(t *testing.T) {
	_, err := Palette(fstest.MapFS{
		"bad.dat": &fstest.MapFile{Data: make([]byte, 12)},
	}, "bad.dat")
	if err == nil {
		t.Fatal("expected truncated palette error")
	}
}
