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

func TestDrawIntoTightChunksPreservesMapCoordinatesAcrossBoundary(t *testing.T) {
	chunks := make(map[[2]int]*draftChunk)
	source := image.NewRGBA(image.Rect(0, 0, 4, 1))
	for x := 0; x < 4; x++ {
		source.SetRGBA(x, 0, color.RGBA{R: uint8(10 + x), A: 255})
	}
	drawIntoTightChunks(chunks, 4, 8, 4, image.Rect(2, 1, 6, 2), source)
	left, right := chunks[[2]int{0, 0}], chunks[[2]int{1, 0}]
	if left.bounds != image.Rect(2, 1, 4, 2) || right.bounds != image.Rect(4, 1, 6, 2) {
		t.Fatalf("unexpected bounds: left=%v right=%v", left.bounds, right.bounds)
	}
	if left.pixels.RGBAAt(0, 0).R != 10 || right.pixels.RGBAAt(1, 0).R != 13 {
		t.Fatalf("pixels were not preserved across chunk boundary")
	}
}

func TestTrimTransparentMarginsPreservesWorldOrigin(t *testing.T) {
	pixels := image.NewRGBA(image.Rect(0, 0, 6, 5))
	pixels.SetRGBA(2, 1, color.RGBA{R: 42, A: 255})
	pixels.SetRGBA(4, 3, color.RGBA{R: 84, A: 255})
	trimmed, origin, ok := trimTransparentMargins(&draftChunk{bounds: image.Rect(100, 200, 106, 205), pixels: pixels})
	if !ok || origin != image.Pt(102, 201) || trimmed.Bounds() != image.Rect(0, 0, 3, 3) {
		t.Fatalf("trim result: ok=%v origin=%v bounds=%v", ok, origin, trimmed.Bounds())
	}
	if trimmed.RGBAAt(0, 0).R != 42 || trimmed.RGBAAt(2, 2).R != 84 {
		t.Fatalf("trimmed pixels do not match their original world positions")
	}
}
