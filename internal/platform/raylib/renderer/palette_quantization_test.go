package raylibRenderer

import (
	"image/color"
	"testing"
)

func TestPaletteLUTEmitsOnlyTargetColors(t *testing.T) {
	palette := color.Palette{color.RGBA{A: 255}, color.RGBA{R: 255, A: 255}, color.RGBA{G: 255, A: 255}, color.RGBA{B: 255, A: 255}}
	lut := buildPaletteLUT(palette)
	if lut.Bounds().Dx() != paletteLUTWidth || lut.Bounds().Dy() != paletteLUTHeight {
		t.Fatalf("LUT bounds = %v", lut.Bounds())
	}
	allowed := make(map[color.RGBA]bool, len(palette))
	for _, entry := range palette {
		allowed[color.RGBAModel.Convert(entry).(color.RGBA)] = true
	}
	for y := 0; y < lut.Bounds().Dy(); y++ {
		for x := 0; x < lut.Bounds().Dx(); x++ {
			if got := lut.RGBAAt(x, y); !allowed[got] {
				t.Fatalf("LUT[%d,%d] = %#v outside target palette", x, y, got)
			}
		}
	}
}

func TestPaletteLUTPreservesExactCubeEndpoint(t *testing.T) {
	palette := color.Palette{color.Black, color.White, color.RGBA{R: 255, G: 255, A: 255}}
	lut := buildPaletteLUT(palette)
	if got := lut.RGBAAt(paletteLUTWidth-1, 0); got != (color.RGBA{R: 255, G: 255, A: 255}) {
		t.Fatalf("yellow endpoint = %#v", got)
	}
}
