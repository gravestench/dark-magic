package assetdecode

import (
	"encoding/binary"
	"image"
	"image/color"
	"testing"
)

func TestFontTableAndMeasuredRendering(t *testing.T) {
	table := make([]byte, fontHeaderBytes+2*fontGlyphBytes)
	copy(table, "Woo!\x01")
	putGlyph := func(offset int, code rune, width, height byte, frame uint16) {
		binary.LittleEndian.PutUint16(table[offset:offset+2], uint16(code))
		table[offset+3], table[offset+4] = width, height
		binary.LittleEndian.PutUint16(table[offset+8:offset+10], frame)
	}
	putGlyph(fontHeaderBytes, 'A', 3, 4, 0)
	putGlyph(fontHeaderBytes+fontGlyphBytes, '?', 2, 4, 1)
	glyphs, err := FontTable(table)
	if err != nil {
		t.Fatal(err)
	}
	frame := func(width int) image.Image {
		result := image.NewRGBA(image.Rect(0, 0, width, 4))
		for index := range result.Pix {
			result.Pix[index] = 0xff
		}
		return result
	}
	font := &BitmapFont{Glyphs: glyphs, Frames: []image.Image{frame(3), frame(2)}, LineHeight: 4}
	rendered, err := font.Render("AXA", color.RGBA{R: 200, A: 255}, 6, "center")
	if err != nil {
		t.Fatal(err)
	}
	if rendered.Bounds().Dx() != 6 || rendered.Bounds().Dy() != 8 {
		t.Fatalf("bounds = %v", rendered.Bounds())
	}
}

func TestFontTableRejectsTruncatedGlyph(t *testing.T) {
	data := append([]byte("Woo!\x01\x00\x00\x00\x00\x00\x00\x00"), 1)
	if _, err := FontTable(data); err == nil {
		t.Fatal("expected truncated glyph error")
	}
}
