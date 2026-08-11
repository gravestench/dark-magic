package maprender

import (
	"image"
	"image/color"
	"testing"
)

func TestDrawIntoChunksAllocatesOnlyIntersectedChunks(t *testing.T) {
	chunks := make(map[[2]int]*image.RGBA)
	source := image.NewRGBA(image.Rect(0, 0, 40, 30))
	for index := range source.Pix {
		source.Pix[index] = 255
	}
	drawIntoChunks(chunks, 64, 256, 256, image.Rect(50, 55, 90, 85), source)
	if len(chunks) != 4 {
		t.Fatalf("allocated %d chunks, want 4", len(chunks))
	}
	for _, key := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
		if chunks[key] == nil {
			t.Fatalf("missing chunk %v", key)
		}
	}
}

func TestDrawIntoChunksPreservesPixelsAcrossBoundary(t *testing.T) {
	chunks := make(map[[2]int]*image.RGBA)
	source := image.NewRGBA(image.Rect(0, 0, 4, 1))
	for x := 0; x < 4; x++ {
		source.SetRGBA(x, 0, color.RGBA{R: uint8(10 + x), A: 255})
	}
	drawIntoChunks(chunks, 4, 8, 4, image.Rect(2, 0, 6, 1), source)
	if got := chunks[[2]int{0, 0}].RGBAAt(2, 0).R; got != 10 {
		t.Fatalf("left first pixel = %d", got)
	}
	if got := chunks[[2]int{1, 0}].RGBAAt(1, 0).R; got != 13 {
		t.Fatalf("right last pixel = %d", got)
	}
}

func TestDrawIntoChunksClipsOutsideCanvas(t *testing.T) {
	chunks := make(map[[2]int]*image.RGBA)
	source := image.NewRGBA(image.Rect(0, 0, 20, 20))
	drawIntoChunks(chunks, 16, 32, 32, image.Rect(-10, -10, 10, 10), source)
	if len(chunks) != 1 || chunks[[2]int{0, 0}] == nil {
		t.Fatalf("chunks = %#v", chunks)
	}
}
