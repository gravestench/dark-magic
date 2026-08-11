package maprender

import (
	"image"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/world"
)

func TestCollisionDiamondIncludesItsProjectedCenter(t *testing.T) {
	shade, visible := collisionColor(world.Flags{BlockWalk: true})
	if !visible {
		t.Fatal("blocked cell has no diagnostic color")
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 64, 32))
	fillDiamond(pixels, 32, 16, 16, 8, shade)
	red, _, _, alpha := pixels.At(32, 16).RGBA()
	if red == 0 || alpha == 0 {
		t.Fatal("projected collision center is transparent")
	}
}

func TestCollisionRegionUsesAuthoritativeMapCanvasCoordinates(t *testing.T) {
	mapData := &world.Map{WidthTiles: 80, HeightTiles: 80, WidthSubtiles: 400, HeightSubtiles: 400}
	pixels, bounds := CollisionRegionImage(mapData, image.Rect(190, 190, 211, 211))
	if pixels.Bounds().Dx() != bounds.Dx() || pixels.Bounds().Dy() != bounds.Dy() {
		t.Fatalf("region pixels=%v bounds=%v", pixels.Bounds(), bounds)
	}
	x, y := mapData.SubtileToPixel(200, 200)
	if !image.Pt(int(x), int(y)).In(bounds) {
		t.Fatalf("hero map point (%v,%v) lies outside collision region %v", x, y, bounds)
	}
	fullWidth := (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelWidth/2 + world.PreviewMargin*2
	if bounds.Dx() >= fullWidth/2 {
		t.Fatalf("bounded collision diagnostic width = %d, full canvas = %d", bounds.Dx(), fullWidth)
	}
}

func TestTileRegionIsBoundedAndContainsGeometry(t *testing.T) {
	mapData := &world.Map{WidthTiles: 100, HeightTiles: 100, WidthSubtiles: 500, HeightSubtiles: 500}
	imageData, bounds := TileRegionImage(mapData, image.Rect(45, 45, 56, 56))
	if imageData.Bounds().Dx() >= 10000 || imageData.Bounds().Dy() >= 10000 {
		t.Fatalf("diagnostic unexpectedly allocated full map: %v", imageData.Bounds())
	}
	if bounds.Empty() {
		t.Fatal("map-canvas bounds are empty")
	}
	visible := false
	for offset := 3; offset < len(imageData.Pix); offset += 4 {
		if imageData.Pix[offset] != 0 {
			visible = true
			break
		}
	}
	if !visible {
		t.Fatal("tile diagnostic contains no visible geometry")
	}
}
