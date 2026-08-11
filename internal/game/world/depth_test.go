package world

import "testing"

func TestWorldDepthCrossesTileBaseline(t *testing.T) {
	wall := TileDepth(LayerUpperWall, 4, 7)
	behind := int(EntityDepth(25, 34))
	inFront := int(EntityDepth(30, 35))
	if !(behind < wall && wall < inFront) {
		t.Fatalf("depth behind=%d wall=%d front=%d", behind, wall, inFront)
	}
}

func TestWorldDepthPassOffsetsAreStable(t *testing.T) {
	shadow := TileDepth(LayerShadow, 2, 3)
	lower := TileDepth(LayerLowerWall, 2, 3)
	upper := TileDepth(LayerUpperWall, 2, 3)
	roof := TileDepth(LayerRoof, 2, 3)
	if !(shadow < lower && lower < upper && upper < roof) {
		t.Fatalf("unexpected pass order: %d %d %d %d", shadow, lower, upper, roof)
	}
}
