package assetdecode

import (
	"encoding/binary"
	"image/color"
	"testing"
	"testing/fstest"

	cof "github.com/gravestench/cof"
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

func TestDCCRejectsTruncatedData(t *testing.T) {
	_, err := DCC(fstest.MapFS{"bad.dcc": &fstest.MapFile{Data: []byte{0x74}}}, "bad.dcc", "")
	if err == nil {
		t.Fatal("expected truncated DCC error")
	}
}

func TestCOFReadsCompositionMetadata(t *testing.T) {
	input := cof.New()
	input.NumberOfDirections = 1
	input.FramesPerDirection = 1
	input.NumberOfLayers = 1
	input.Speed = 128
	input.CofLayers = []cof.CofLayer{{Type: 0, Selectable: true, WeaponClass: cof.WeaponClassFromString("hth")}}
	input.AnimationFrames = []cof.FrameEvent{1}
	input.Priority = [][][]cof.CompositeType{{{0}}}
	decoded, err := COF(fstest.MapFS{"unit.cof": &fstest.MapFile{Data: cof.Marshal(input)}}, "unit.cof")
	if err != nil {
		t.Fatal(err)
	}
	if decoded.NumberOfLayers != 1 || decoded.FramesPerDirection != 1 || decoded.Priority[0][0][0] != 0 {
		t.Fatalf("COF = %#v", decoded)
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
