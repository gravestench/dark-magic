//go:build !ebitengine

package desktop

import (
	"image"
	"image/color"
	"testing"
)

// TestViewportPolicies verifies contain centers an aspect-preserving surface while stretch uses every window pixel;
// input mapping and final presentation both rely on these exact coordinates.
func TestViewportPolicies(t *testing.T) {
	contained, err := calculateViewport(1000, 600, 800, 600, "contain")
	if err != nil {
		t.Fatal(err)
	}

	if contained.x != 100 || contained.y != 0 || contained.width != 800 || contained.height != 600 {
		t.Fatalf("contained viewport = %+v", contained)
	}

	stretched, err := calculateViewport(1000, 600, 800, 600, "stretch")
	if err != nil {
		t.Fatal(err)
	}

	if stretched.x != 0 || stretched.y != 0 || stretched.width != 1000 || stretched.height != 600 {
		t.Fatalf("stretched viewport = %+v", stretched)
	}
}

// TestPaletteQuantizationPreservesAlpha verifies nearest-color selection does not make translucent pixels opaque or
// disturb fully transparent pixels.
func TestPaletteQuantizationPreservesAlpha(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2, 1))
	source.SetRGBA(0, 0, color.RGBA{R: 240, G: 10, B: 10, A: 127})
	source.SetRGBA(1, 0, color.RGBA{R: 240, G: 10, B: 10, A: 0})

	got := quantizeImage(source, color.Palette{color.Black, color.RGBA{R: 255, A: 255}})
	if pixel := got.RGBAAt(0, 0); pixel.R != 255 || pixel.G != 0 || pixel.B != 0 || pixel.A != 127 {
		t.Fatalf("quantized pixel = %#v", pixel)
	}

	if pixel := got.RGBAAt(1, 0); pixel.A != 0 {
		t.Fatalf("transparent pixel alpha = %d", pixel.A)
	}
}
