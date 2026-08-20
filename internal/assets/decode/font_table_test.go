package assetdecode

import (
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

// TestFontTableAndMeasuredRendering proves that decoded metrics drive fallback
// advances and wrapping, preserving the integration between Woo! records and layout.
func TestFontTableAndMeasuredRendering(t *testing.T) {
	table := make([]byte, fontHeaderBytes+2*fontGlyphBytes)
	copy(table, "Woo!\x01")
	putTestGlyph(table, fontHeaderBytes, 'A', 3, 4, 0)
	putTestGlyph(table, fontHeaderBytes+fontGlyphBytes, '?', 2, 4, 1)

	glyphs, err := FontTable(table)
	if err != nil {
		t.Fatal(err)
	}

	font := &BitmapFont{
		Glyphs:     glyphs,
		Frames:     []image.Image{opaqueTestFrame(3, 4), opaqueTestFrame(2, 4)},
		LineHeight: 4,
	}

	rendered, err := font.Render("AXA", color.RGBA{R: 200, A: 255}, 6, "center")
	if err != nil {
		t.Fatal(err)
	}

	if rendered.Bounds().Dx() != 6 || rendered.Bounds().Dy() != 8 {
		t.Fatalf("bounds = %v", rendered.Bounds())
	}
}

// TestFontTableRejectsTruncatedGlyph ensures incomplete fixed-width records fail
// before their partial metrics can affect rendering or frame lookup.
func TestFontTableRejectsTruncatedGlyph(t *testing.T) {
	data := append([]byte("Woo!\x01\x00\x00\x00\x00\x00\x00\x00"), 1)
	if _, err := FontTable(data); err == nil {
		t.Fatal("expected truncated glyph error")
	}
}

// putTestGlyph writes one complete metric record so tests expose authored values
// without duplicating binary offsets in scenario setup.
func putTestGlyph(
	table []byte,
	offset int,
	code rune,
	width byte,
	height byte,
	frame uint16,
) {
	binary.LittleEndian.PutUint16(table[offset:offset+2], uint16(code))
	table[offset+3] = width
	table[offset+4] = height
	binary.LittleEndian.PutUint16(table[offset+8:offset+10], frame)
}

// opaqueTestFrame creates a fully opaque glyph with stable dimensions, keeping
// metric and wrapping assertions independent of palette details.
func opaqueTestFrame(width, height int) image.Image {
	frame := image.NewRGBA(image.Rect(0, 0, width, height))
	for index := range frame.Pix {
		frame.Pix[index] = 0xff
	}

	return frame
}
