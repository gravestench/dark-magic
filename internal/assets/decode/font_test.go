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

func TestBitmapFontUsesSharedTopOriginForMultilineText(t *testing.T) {
	t.Parallel()

	frame := image.NewAlpha(image.Rect(0, 0, 2, 4))
	for index := range frame.Pix {
		frame.Pix[index] = 0xff
	}
	font := &BitmapFont{
		Glyphs:     map[rune]Glyph{'A': {Width: 2, Height: 2, Frame: 0}},
		Frames:     []image.Image{frame},
		LineHeight: 6,
	}
	rendered, err := font.Render("A\nA", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	if rendered.Bounds() != image.Rect(0, 0, 2, 12) {
		t.Fatalf("bounds = %v", rendered.Bounds())
	}
	for _, point := range []image.Point{{0, 0}, {0, 3}, {0, 6}, {0, 9}} {
		_, _, _, alpha := rendered.At(point.X, point.Y).RGBA()
		if alpha == 0 {
			t.Errorf("shared-origin pixel %v transparent", point)
		}
	}
	for _, point := range []image.Point{{0, 4}, {0, 5}, {0, 10}, {0, 11}} {
		_, _, _, alpha := rendered.At(point.X, point.Y).RGBA()
		if alpha != 0 {
			t.Errorf("line padding pixel %v opaque", point)
		}
	}
}

func TestBitmapFontAppliesDC6FrameOffsets(t *testing.T) {
	t.Parallel()

	frame := image.NewAlpha(image.Rect(0, 0, 1, 1))
	frame.SetAlpha(0, 0, color.Alpha{A: 255})
	font := &BitmapFont{
		Glyphs:       map[rune]Glyph{'A': {Width: 3, Height: 1, Frame: 0}},
		Frames:       []image.Image{frame},
		FrameOffsets: []image.Point{{X: 1, Y: 2}},
		LineHeight:   4,
	}
	rendered, err := font.Render("A", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, alpha := rendered.At(1, 2).RGBA(); alpha == 0 {
		t.Fatal("offset glyph pixel is transparent")
	}
	if _, _, _, alpha := rendered.At(0, 0).RGBA(); alpha != 0 {
		t.Fatal("unoffset origin is opaque")
	}
}

func TestBitmapFontPreservesPaletteShadingWhenTinted(t *testing.T) {
	t.Parallel()

	frame := image.NewRGBA(image.Rect(0, 0, 2, 1))
	frame.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	frame.SetRGBA(1, 0, color.RGBA{R: 40, G: 20, B: 10, A: 255})
	font := &BitmapFont{Glyphs: map[rune]Glyph{'A': {Width: 2, Height: 1, Frame: 0}}, Frames: []image.Image{frame}, LineHeight: 1}
	rendered, err := font.Render("A", color.RGBA{R: 128, G: 255, B: 128, A: 255}, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	bright := color.RGBAModel.Convert(rendered.At(0, 0)).(color.RGBA)
	dark := color.RGBAModel.Convert(rendered.At(1, 0)).(color.RGBA)
	if bright.R <= dark.R || bright.G <= dark.G || bright.B <= dark.B {
		t.Fatalf("palette shading flattened: bright=%#v dark=%#v", bright, dark)
	}
	if bright.R >= 200 || bright.G != 100 || bright.B >= 50 {
		t.Fatalf("tint was not multiplicative: %#v", bright)
	}
}

func TestBitmapFontRendersInlineColorRunsWithoutMeasuringTokens(t *testing.T) {
	t.Parallel()

	frame := image.NewRGBA(image.Rect(0, 0, 1, 1))
	frame.SetRGBA(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	font := &BitmapFont{
		Glyphs:     map[rune]Glyph{'A': {Width: 1, Height: 1, Frame: 0}},
		Frames:     []image.Image{frame},
		LineHeight: 1,
	}
	rendered, err := font.Render("[red]A[blue]A", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	if rendered.Bounds().Dx() != 2 {
		t.Fatalf("tokenized text width = %d, want 2", rendered.Bounds().Dx())
	}
	red := color.RGBAModel.Convert(rendered.At(0, 0)).(color.RGBA)
	blue := color.RGBAModel.Convert(rendered.At(1, 0)).(color.RGBA)
	if red != (color.RGBA{R: 0xff, G: 0x4d, B: 0x4d, A: 0xff}) {
		t.Fatalf("red run = %#v", red)
	}
	if blue != (color.RGBA{R: 0x69, G: 0x69, B: 0xff, A: 0xff}) {
		t.Fatalf("blue run = %#v", blue)
	}
}

func TestBitmapFontPrefersPL2TextTransformOverRGBFallback(t *testing.T) {
	t.Parallel()

	base := image.NewRGBA(image.Rect(0, 0, 1, 1))
	base.SetRGBA(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	transformed := image.NewRGBA(image.Rect(0, 0, 1, 1))
	transformed.SetRGBA(0, 0, color.RGBA{R: 31, G: 7, B: 63, A: 255})
	font := &BitmapFont{
		Glyphs:     map[rune]Glyph{'A': {Width: 1, Height: 1, Frame: 0}},
		Frames:     []image.Image{base},
		TextFrames: map[int][]image.Image{1: {transformed}},
		LineHeight: 1,
	}
	rendered, err := font.Render("[red]A", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	got := color.RGBAModel.Convert(rendered.At(0, 0)).(color.RGBA)
	if got != (color.RGBA{R: 31, G: 7, B: 63, A: 255}) {
		t.Fatalf("PL2-transformed glyph = %#v", got)
	}
}

func TestBitmapFontWhiteKeepsPaletteAuthoredGlyph(t *testing.T) {
	t.Parallel()

	base := image.NewRGBA(image.Rect(0, 0, 1, 1))
	base.SetRGBA(0, 0, color.RGBA{R: 220, G: 210, B: 190, A: 255})
	blackShift := image.NewRGBA(image.Rect(0, 0, 1, 1))
	blackShift.SetRGBA(0, 0, color.RGBA{A: 255})
	font := &BitmapFont{
		Glyphs:     map[rune]Glyph{'A': {Width: 1, Height: 1, Frame: 0}},
		Frames:     []image.Image{base},
		TextFrames: map[int][]image.Image{0: {blackShift}},
		LineHeight: 1,
	}
	rendered, err := font.Render("[white]A", color.White, 0, "left")
	if err != nil {
		t.Fatal(err)
	}
	if got := color.RGBAModel.Convert(rendered.At(0, 0)).(color.RGBA); got != (color.RGBA{R: 220, G: 210, B: 190, A: 255}) {
		t.Fatalf("white run = %#v, want palette-authored glyph", got)
	}
}
