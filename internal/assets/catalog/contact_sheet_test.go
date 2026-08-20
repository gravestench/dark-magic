package assetcatalog

import (
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/assets/decode"
)

// TestDC6ContactSheetRejectsNil ensures callers receive a controlled error instead of a decoder dereference panic.
func TestDC6ContactSheetRejectsNil(t *testing.T) {
	if _, err := DC6ContactSheet(nil); err == nil {
		t.Fatal("expected nil DC6 to fail")
	}
}

// TestReadPalette protects the BGR-to-RGBA conversion used before contact-sheet rendering. A channel-order regression
// would produce structurally valid diagnostics whose colors convey the wrong asset information.
func TestReadPalette(t *testing.T) {
	data := make([]byte, 256*3)
	data[3], data[4], data[5] = 10, 20, 30

	palette, err := assetdecode.Palette(fstest.MapFS{"pal.dat": {Data: data}}, "pal.dat")
	if err != nil {
		t.Fatal(err)
	}

	r, g, b, a := palette[1].RGBA()
	if r>>8 != 30 || g>>8 != 20 || b>>8 != 10 || a>>8 != 255 {
		t.Fatalf("unexpected color: %d %d %d %d", r>>8, g>>8, b>>8, a>>8)
	}
}
