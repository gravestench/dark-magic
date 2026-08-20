package world

import "testing"

// TestWorldDepthCrossesTileBaseline verifies entities cross the baseline used to split lower and upper tile passes.
func TestWorldDepthCrossesTileBaseline(t *testing.T) {
	wall := TileDepth(LayerUpperWall, 4, 7)
	behind := int(EntityDepth(20, 37)) // local column zero

	inFront := int(EntityDepth(21, 36)) // same tile, interior
	if behind >= wall || wall >= inFront {
		t.Fatalf("depth behind=%d wall=%d front=%d", behind, wall, inFront)
	}
}

// TestWorldDepthPassOffsetsAreStable pins global pass ordering so batched renderers cannot reorder legacy layers.
func TestWorldDepthPassOffsetsAreStable(t *testing.T) {
	floor := TileDepth(LayerFloor, 2, 3)
	shadow := TileDepth(LayerShadow, 2, 3)
	lower := TileDepth(LayerLowerWall, 2, 3)
	upper := TileDepth(LayerUpperWall, 2, 3)

	roof := TileDepth(LayerRoof, 2, 3)
	if lower >= floor || floor >= shadow || shadow >= upper || upper >= roof {
		t.Fatalf("unexpected pass order: %d %d %d %d %d", lower, floor, shadow, upper, roof)
	}
}

// TestEntityBoundaryRuleMatchesLegacyTilePasses protects the exact entity/tile equality behavior at a baseline.
func TestEntityBoundaryRuleMatchesLegacyTilePasses(t *testing.T) {
	wall := TileDepth(LayerUpperWall, 3, 2)
	for _, point := range [][2]float64{{15, 10}, {19, 10}, {15, 14}} {
		if depth := int(EntityDepth(point[0], point[1])); depth >= wall {
			t.Fatalf("boundary entity %v depth=%d should be behind wall=%d", point, depth, wall)
		}
	}

	for _, point := range [][2]float64{{16, 11}, {19, 14}} {
		if depth := int(EntityDepth(point[0], point[1])); depth <= wall {
			t.Fatalf("interior entity %v depth=%d should be in front of wall=%d", point, depth, wall)
		}
	}
}
