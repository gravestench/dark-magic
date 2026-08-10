package raylibRenderer

import (
	"image"
	"image/color"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
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

func TestContiguousRGBAUsesDecodedBuffer(t *testing.T) {
	img := image.NewRGBA(image.Rect(4, 7, 6, 9))
	img.SetRGBA(4, 7, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	pixels, ok := contiguousRGBA(img)
	if !ok {
		t.Fatal("contiguous RGBA image did not use direct upload path")
	}
	if len(pixels) != 16 || pixels[0] != 10 || pixels[3] != 255 {
		t.Fatalf("pixels = %v", pixels)
	}
	pixels[1] = 99
	if got := img.RGBAAt(4, 7).G; got != 99 {
		t.Fatalf("direct buffer did not alias source: green = %d", got)
	}
}

func TestContiguousRGBARejectsPaddedSubimage(t *testing.T) {
	parent := image.NewRGBA(image.Rect(0, 0, 4, 4))
	subimage := parent.SubImage(image.Rect(1, 1, 3, 3))
	if _, ok := contiguousRGBA(subimage); ok {
		t.Fatal("padded subimage unexpectedly used direct upload path")
	}
}

func TestIntersectClipPreservesParentAndChildBounds(t *testing.T) {
	parent := rl.NewRectangle(10, 10, 100, 80)
	child := rl.NewRectangle(50, 0, 100, 40)
	got := intersectClip(&parent, &child)
	want := rl.NewRectangle(50, 10, 60, 30)
	if *got != want {
		t.Fatalf("intersection = %#v, want %#v", *got, want)
	}
}

func TestLeafCullingUsesLogicalViewport(t *testing.T) {
	node := &node{image: image.NewRGBA(image.Rect(0, 0, 20, 10)), origin: rl.Vector2{X: .5, Y: .5}, local: rl.MatrixIdentity(), world: rl.MatrixIdentity()}
	node.SetPosition(110, 50)
	node.UpdateWorldMatrix(rl.MatrixIdentity(), false)
	if !node.outside(100, 100) {
		t.Fatal("offscreen leaf was not culled")
	}
	node.SetPosition(95, 50)
	node.UpdateWorldMatrix(rl.MatrixIdentity(), false)
	if node.outside(100, 100) {
		t.Fatal("partially visible leaf was culled")
	}
}

func BenchmarkTexturePixelPreparation(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	b.Run("direct-rgba", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			pixels, ok := contiguousRGBA(img)
			if !ok || len(pixels) == 0 {
				b.Fatal("direct path unavailable")
			}
		}
	})
	b.Run("generic-conversion", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			if pixels := getAllPixelData(img); len(pixels) == 0 {
				b.Fatal("conversion returned no pixels")
			}
		}
	})
}
