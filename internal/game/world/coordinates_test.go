package world

import "testing"

func TestCoordinateSpacesRoundTrip(t *testing.T) {
	space := Coordinates{HeightTiles: 25}
	for _, point := range []Point{{}, {1.25, 9.75}, {100, 42}} {
		worldPixel := space.SubtileToWorldPixel(point)
		assertPointNear(t, space.WorldPixelToSubtile(worldPixel), point)
		assertPointNear(t, space.SubtileToTile(space.TileToSubtile(point)), point)
	}
}

func TestCameraScreenRoundTrip(t *testing.T) {
	space := Coordinates{HeightTiles: 25}
	point := Point{31.25, 47.75}
	camera := Point{40, 40}
	anchor := Point{200, 300}
	screen := space.SubtileToScreen(point, camera, anchor)
	assertPointNear(t, space.ScreenToSubtile(screen, camera, anchor), point)
	assertPointNear(t, space.SubtileToScreen(camera, camera, anchor), anchor)
}

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

func assertPointNear(t *testing.T, got, want Point) {
	t.Helper()
	const epsilon = 0.000001
	if absFloat(got.X-want.X) > epsilon || absFloat(got.Y-want.Y) > epsilon {
		t.Fatalf("point = %#v, want %#v", got, want)
	}
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
