package assetdecode

import (
	"image"
	"image/color"
	"testing"
)

// TestBitmapFontUsesSharedTopOriginForMultilineText verifies that every line
// shares the same visual origin and reserves its complete line-height padding.
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

// TestBitmapFontAppliesDC6FrameOffsets protects authored glyph placement while
// leaving the logical advance anchored to the unshifted text cursor.
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

// TestBitmapFontPreservesPaletteShadingWhenTinted verifies multiplicative tint
// keeps relative highlights and shadows instead of flattening glyphs into masks.
func TestBitmapFontPreservesPaletteShadingWhenTinted(t *testing.T) {
	t.Parallel()

	frame := image.NewRGBA(image.Rect(0, 0, 2, 1))
	frame.SetRGBA(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	frame.SetRGBA(1, 0, color.RGBA{R: 40, G: 20, B: 10, A: 255})
	font := &BitmapFont{
		Glyphs:     map[rune]Glyph{'A': {Width: 2, Height: 1, Frame: 0}},
		Frames:     []image.Image{frame},
		LineHeight: 1,
	}

	rendered, err := font.Render(
		"A",
		color.RGBA{R: 128, G: 255, B: 128, A: 255},
		0,
		"left",
	)
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

// TestBitmapFontRendersInlineColorRunsWithoutMeasuringTokens ensures control
// tokens change tint without occupying pixels or logical advance width.
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

// TestBitmapFontPrefersPL2TextTransformOverRGBFallback proves authored palette
// transforms take precedence over approximate direct-color fallbacks.
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

// TestBitmapFontColorScopeRestoresCallerTint verifies every reset spelling
// returns to the caller-supplied tint rather than a hard-coded default.
func TestBitmapFontColorScopeRestoresCallerTint(t *testing.T) {
	t.Parallel()

	frame := image.NewRGBA(image.Rect(0, 0, 1, 1))
	frame.SetRGBA(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	font := &BitmapFont{
		Glyphs:     map[rune]Glyph{'A': {Width: 1, Height: 1, Frame: 0}},
		Frames:     []image.Image{frame},
		LineHeight: 1,
	}
	base := color.RGBA{R: 100, G: 120, B: 140, A: 255}

	rendered, err := font.Render("[red]A[/]A[green]A[/green]A", base, 0, "left")
	if err != nil {
		t.Fatal(err)
	}

	want := []color.RGBA{
		{R: 0xff, G: 0x4d, B: 0x4d, A: 0xff},
		base,
		{G: 0xff, A: 0xff},
		base,
	}
	for x, expected := range want {
		if got := color.RGBAModel.Convert(rendered.At(x, 0)).(color.RGBA); got != expected {
			t.Fatalf("pixel %d = %#v, want %#v", x, got, expected)
		}
	}
}

// TestBitmapFontWhiteKeepsPaletteAuthoredGlyph protects the reserved PL2 shift
// zero semantics, where white must not replace the font's authored shading.
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

	got := color.RGBAModel.Convert(rendered.At(0, 0)).(color.RGBA)

	want := color.RGBA{R: 220, G: 210, B: 190, A: 255}
	if got != want {
		t.Fatalf("white run = %#v, want palette-authored glyph", got)
	}
}

// TestBitmapFontMeasureMatchesRenderedTexture protects pre-layout UI sizing from renderer drift.
func TestBitmapFontMeasureMatchesRenderedTexture(t *testing.T) {
	font := &BitmapFont{
		Glyphs: map[rune]Glyph{
			'A': {Width: 3, Height: 4, Frame: 0},
			' ': {Width: 2, Height: 4, Frame: 1},
			'?': {Width: 2, Height: 4, Frame: 1},
		},
		Frames:     []image.Image{opaqueTestFrame(3, 4), opaqueTestFrame(2, 4)},
		LineHeight: 4,
	}

	measured, err := font.Measure("[gold]A A[/]", 5)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := font.Render("[gold]A A[/]", color.White, 5, "center")
	if err != nil {
		t.Fatal(err)
	}
	if measured != rendered.Bounds().Size() {
		t.Fatalf("measured %v, rendered %v", measured, rendered.Bounds().Size())
	}
}
