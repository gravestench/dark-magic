package main

import (
	"image"
	"image/color"
	"testing"
)

// TestNormalizedChromeRepeatsWithoutTileOutlines verifies that repeatable axes
// share exact edge pixels without flattening the tile's authored interior.
func TestNormalizedChromeRepeatsWithoutTileOutlines(t *testing.T) {
	const size = 16
	for _, name := range []string{"panel_fill", "panel_top", "panel_left", "button_idle", "tab_selected"} {
		frame := image.NewNRGBA(image.Rect(0, 0, size, size))
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				frame.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 4), G: uint8(y * 4), A: 255})
			}
		}
		normalizeChromeSeams(frame, name)
		for y := 0; y < size; y++ {
			left := color.NRGBAModel.Convert(frame.At(0, y)).(color.NRGBA)
			right := color.NRGBAModel.Convert(frame.At(size-1, y)).(color.NRGBA)
			if name != "panel_left" && left != right {
				t.Fatalf("%s horizontal seam at row %d: %#v != %#v", name, y, left, right)
			}
		}
		for x := 0; x < size; x++ {
			top := color.NRGBAModel.Convert(frame.At(x, 0)).(color.NRGBA)
			bottom := color.NRGBAModel.Convert(frame.At(x, size-1)).(color.NRGBA)
			if name != "panel_top" && top != bottom {
				t.Fatalf("%s vertical seam at column %d: %#v != %#v", name, x, top, bottom)
			}
		}
	}
}

// TestSmallPixelFontsUseOneOpaquePaletteIndex ensures palette conversion can
// never turn antialiased edge coverage into muddy fringe colors.
func TestSmallPixelFontsUseOneOpaquePaletteIndex(t *testing.T) {
	palette := editorPalette()
	for _, spec := range editorFontSpecs {
		if !spec.PixelArt {
			continue
		}
		glyph := rasterizePixelGlyph(palette, spec, 'M')
		seen := make(map[byte]bool)
		for _, index := range glyph.Pixels {
			if index != 0 {
				seen[index] = true
			}
		}
		if len(seen) != 1 {
			t.Fatalf("%s opaque palette indexes = %v, want exactly one", spec.Name, seen)
		}
		if glyph.Width != spec.CellWidth || glyph.Height != spec.LineHeight {
			t.Fatalf("%s glyph size = %dx%d", spec.Name, glyph.Width, glyph.Height)
		}
	}
}

func TestEveryCompositionComponentHasVariants(t *testing.T) {
	for _, spec := range componentSpecs {
		if count := componentVariantCount(spec); count < 2 {
			t.Errorf("%s variant count = %d, want at least two", spec.Name, count)
		}
	}
}

func TestCompositionSpritesKeepNativeBoundsAndRepeatableEdges(t *testing.T) {
	for _, spec := range componentSpecs {
		for variant := 0; variant < componentVariantCount(spec); variant++ {
			frame := drawNativeComponent(spec, variant)
			if got := frame.Bounds().Size(); got != (image.Pt(spec.Width, spec.Height)) {
				t.Fatalf("%s[%d] bounds = %v, want %dx%d", spec.Name, variant, got, spec.Width, spec.Height)
			}
			if spec.RepeatAxis == "x" || spec.RepeatAxis == "xy" {
				for y := 0; y < spec.Height; y++ {
					if frame.NRGBAAt(0, y) != frame.NRGBAAt(spec.Width-1, y) {
						t.Fatalf("%s[%d] horizontal seam at y=%d", spec.Name, variant, y)
					}
				}
			}
			if spec.RepeatAxis == "y" || spec.RepeatAxis == "xy" {
				for x := 0; x < spec.Width; x++ {
					if frame.NRGBAAt(x, 0) != frame.NRGBAAt(x, spec.Height-1) {
						t.Fatalf("%s[%d] vertical seam at x=%d", spec.Name, variant, x)
					}
				}
			}
		}
	}
}

func TestPanelChromeDoesNotBakeInThePanelFill(t *testing.T) {
	corner := drawNativeComponent(componentSpecs[0], 0)
	if alpha := corner.NRGBAAt(24, 24).A; alpha != 0 {
		t.Fatalf("panel corner inner field alpha = %d, want transparent border-only chrome", alpha)
	}
	fill := drawNativeComponent(componentSpecs[4], 0)
	if alpha := fill.NRGBAAt(16, 16).A; alpha != 255 {
		t.Fatalf("panel fill alpha = %d, want opaque tiled fill", alpha)
	}
}
