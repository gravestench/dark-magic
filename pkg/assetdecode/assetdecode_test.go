package assetdecode

import (
	"encoding/binary"
	"image/color"
	"testing"
	"testing/fstest"
)

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
	got := color.RGBAModel.Convert(frame.ToImageRGBA().At(0, 0)).(color.RGBA)
	if got != (color.RGBA{R: 10, G: 20, B: 30, A: 0xff}) {
		t.Fatalf("pixel = %#v", got)
	}
	if _, err := Frame(asset, 1, 0); err == nil {
		t.Fatal("expected direction bounds error")
	}
}

func TestPaletteRejectsTruncatedData(t *testing.T) {
	_, err := Palette(fstest.MapFS{"bad.dat": &fstest.MapFile{Data: make([]byte, 12)}}, "bad.dat")
	if err == nil {
		t.Fatal("expected truncated palette error")
	}
}

func onePixelDC6(index byte) []byte {
	data := make([]byte, 16+8+4+32+3+3)
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(data[offset:offset+4], value) }
	put(0, 6)
	put(16, 1)
	put(20, 1)
	put(24, 28)
	put(32, 1)
	put(36, 1)
	put(56, 3)
	data[60], data[61], data[62] = 1, index, 0x80
	return data
}
