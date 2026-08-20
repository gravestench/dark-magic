package assetdecode

import (
	"image/color"
	"testing"
	"testing/fstest"
)

// TestPaletteAndDC6Frame covers the complete palette-to-frame path and verifies
// that BGR channel order and direction bounds survive codec integration.
func TestPaletteAndDC6Frame(t *testing.T) {
	palette := make([]byte, 256*3)
	palette[3], palette[4], palette[5] = 10, 20, 30
	source := fstest.MapFS{
		"unit/pal.dat": &fstest.MapFile{Data: palette},
		"unit/one.dc6": &fstest.MapFile{Data: onePixelDC6(1)},
	}

	asset, err := DC6(source, "unit/one.dc6", "unit/pal.dat")
	if err != nil {
		t.Fatal(err)
	}

	frame, err := Frame(asset, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := FrameImage(asset, frame)
	if err != nil {
		t.Fatal(err)
	}

	got := color.RGBAModel.Convert(decoded.At(0, 0)).(color.RGBA)
	if got != (color.RGBA{R: 30, G: 20, B: 10, A: 0xff}) {
		t.Fatalf("pixel = %#v", got)
	}

	if _, err := Frame(asset, 1, 0); err == nil {
		t.Fatal("expected direction bounds error")
	}
}

// TestDC6UsesRandomAccessWhenAvailable proves that DC6 decoding succeeds even
// when the fixture forbids the sequential-read fallback.
func TestDC6UsesRandomAccessWhenAvailable(t *testing.T) {
	asset, err := DC6(readerAtFS{"one.dc6": onePixelDC6(1)}, "one.dc6", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(asset.Directions) != 1 || len(asset.Directions[0].Frames) != 1 {
		t.Fatalf("unexpected decoded shape: %#v", asset.Directions)
	}
}
