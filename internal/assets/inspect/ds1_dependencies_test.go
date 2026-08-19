package assetinspect

import (
	"slices"
	"testing"
)

// TestResolveDS1TilePathsConvertsLegacyTG1Declarations protects the compatibility
// normalization and first-declaration ordering expected by archive-backed callers.
func TestResolveDS1TilePathsConvertsLegacyTG1Declarations(t *testing.T) {
	t.Parallel()

	got := resolveDS1TilePaths([]string{
		`\D2\Data\Global\Tiles\Act1\Barracks\basewall.tg1`,
		`\d2\data\global\tiles\act1\barracks\floor.TG1`,
		`data/global/tiles/Act1/Barracks/basewall.tg1`,
	})
	want := []string{
		"Data/Global/Tiles/Act1/Barracks/basewall.dt1",
		"data/global/tiles/act1/barracks/floor.dt1",
	}

	if !slices.Equal(got, want) {
		t.Fatalf("resolved paths = %#v, want %#v", got, want)
	}
}
