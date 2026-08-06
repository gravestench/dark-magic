package raylibRenderer

import (
	"image"
	"image/color"
	"testing"
)

func TestGetAllPixelDataConvertsModelsAndRespectsBounds(t *testing.T) {
	img := image.NewNRGBA(image.Rect(4, 7, 6, 8))
	img.SetNRGBA(4, 7, color.NRGBA{R: 10, G: 20, B: 30, A: 255})
	img.SetNRGBA(5, 7, color.NRGBA{R: 40, G: 50, B: 60, A: 255})
	pixels := getAllPixelData(img)
	if len(pixels) != 2 {
		t.Fatalf("pixel count = %d, want 2", len(pixels))
	}
	if pixels[0] != (color.RGBA{R: 10, G: 20, B: 30, A: 255}) {
		t.Fatalf("first pixel = %#v", pixels[0])
	}
}
