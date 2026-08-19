package assetdecode

import (
	"testing"
	"testing/fstest"
)

// TestDCCRejectsTruncatedData verifies that codec failures retain their error
// path instead of exposing a partially decoded animation.
func TestDCCRejectsTruncatedData(t *testing.T) {
	_, err := DCC(fstest.MapFS{
		"bad.dcc": &fstest.MapFile{Data: []byte{0x74}},
	}, "bad.dcc", "")
	if err == nil {
		t.Fatal("expected truncated DCC error")
	}
}
