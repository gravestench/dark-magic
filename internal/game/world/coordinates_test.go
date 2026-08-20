package world

import "testing"

// TestCoordinateSpacesRoundTrip pins tile/subtile and subtile/pixel inverses so adapters can safely compose them.
func TestCoordinateSpacesRoundTrip(t *testing.T) {
	space := Coordinates{HeightTiles: 25}
	for _, point := range []Point{{}, {1.25, 9.75}, {100, 42}} {
		worldPixel := space.SubtileToWorldPixel(point)
		assertPointNear(t, space.WorldPixelToSubtile(worldPixel), point)
		assertPointNear(t, space.SubtileToTile(space.TileToSubtile(point)), point)
	}
}

// TestCameraScreenRoundTrip ensures viewport anchoring does not lose world-pixel coordinates.
func TestCameraScreenRoundTrip(t *testing.T) {
	space := Coordinates{HeightTiles: 25}
	point := Point{31.25, 47.75}
	camera := Point{40, 40}
	anchor := Point{200, 300}
	screen := space.SubtileToScreen(point, camera, anchor)
	assertPointNear(t, space.ScreenToSubtile(screen, camera, anchor), point)
	assertPointNear(t, space.SubtileToScreen(camera, camera, anchor), anchor)
}

// TestCollisionCellUsesSubtileCenters protects the half-subtile boundary shared by movement and direct collision reads.
func TestCollisionCellUsesSubtileCenters(t *testing.T) {
	for _, test := range []struct {
		value float64
		want  int
	}{{0.49, 0}, {0.5, 1}, {7.49, 7}, {7.5, 8}} {
		if got := CollisionCell(test.value); got != test.want {
			t.Fatalf("CollisionCell(%v) = %d, want %d", test.value, got, test.want)
		}
	}
}

// TestSubtileCentersCoverExactlyOneDT1FloorDiamond pins the projection span used to align collision overlays with DT1.
func TestSubtileCentersCoverExactlyOneDT1FloorDiamond(t *testing.T) {
	space := Coordinates{HeightTiles: 1}
	top := space.SubtileToWorldPixel(Point{X: 0, Y: 0})
	bottom := space.SubtileToWorldPixel(Point{X: SubtilesPerTile - 1, Y: SubtilesPerTile - 1})

	halfSubtileHeight := float64(TilePixelHeight) / (2 * SubtilesPerTile)
	if got := top.Y - halfSubtileHeight; got != PreviewMargin {
		t.Fatalf("first collision diamond begins at y=%v, want tile top %d", got, PreviewMargin)
	}

	if got := bottom.Y + halfSubtileHeight; got != PreviewMargin+TilePixelHeight {
		t.Fatalf("last collision diamond ends at y=%v, want tile bottom %d", got, PreviewMargin+TilePixelHeight)
	}
}

// assertPointNear compares projected floats with enough tolerance for inverse arithmetic but catches visible drift.
func assertPointNear(t *testing.T, got, want Point) {
	t.Helper()

	const epsilon = 0.000001
	if absFloat(got.X-want.X) > epsilon || absFloat(got.Y-want.Y) > epsilon {
		t.Fatalf("point = %#v, want %#v", got, want)
	}
}

// absFloat keeps tolerance assertions independent from sign while avoiding a testing dependency on extra helpers.
func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}

	return value
}
