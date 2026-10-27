package mapgen

import (
	"context"
	"math"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/world"
)

// TestDenOfEvilOneRoomReal exercises the production recipe and owned 1.14d
// decoding. CI skips because no protected assets are distributed with this test.
func TestDenOfEvilOneRoomReal(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to the owned Expansion 1.14d MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Error(err)
		}
	})
	runtime, err := NewRuntime(t.Context(), source, recordstore.New(source))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	for difficulty := 0; difficulty < 3; difficulty++ {
		seen := make(map[[2]int]bool)
		// Exercise every eligible orientation/variant through Lua selection. The
		// fixed seed also remains part of this integration acceptance check.
		for attempt := 0; attempt <= 128 && len(seen) < 8; attempt++ {
			seed := attempt - 1
			if attempt == 0 {
				seed = 42
			}
			zone, err := runtime.Generate(t.Context(), "maze", float64(8), float64(seed), float64(difficulty))
			if err != nil {
				t.Fatal(err)
			}
			if len(zone.Stamps()) != 1 || len(zone.Rooms()) != 1 || zone.Bounds().Width != 24 || zone.Bounds().Height != 24 {
				t.Fatalf("Den recipe is not a single 24x24 room: %#v", zone.Definition())
			}
			stamp := zone.Stamps()[0]
			choice := [2]int{stamp.PresetDef, stamp.Variant}
			if seen[choice] {
				continue
			}
			seen[choice] = true
			materializer, err := world.NewMaterializer(source, zone)
			if err != nil {
				t.Fatal(err)
			}
			if err := materializer.Step(t.Context()); err != nil {
				t.Fatal(err)
			}
			decoded, err := materializer.Result()
			if err != nil {
				t.Fatal(err)
			}
			var markers []world.SpecialTile
			for _, tile := range decoded.SpecialTiles {
				if tile.Orientation == 10 && tile.MainIndex == 0 && tile.SubIndex == 21 {
					markers = append(markers, tile)
				}
			}
			if len(markers) != 1 {
				t.Fatalf("cave-up full-identity marker count=%d; facts=%#v", len(markers), decoded.SpecialTiles)
			}
			marker := markers[0]
			x, y := float64(marker.X*world.SubtilesPerTile+2), float64(marker.Y*world.SubtilesPerTile+2)
			px, py, found := decoded.OpenPointNearSubtileForRadius(x, y, 1)
			if !found || math.Hypot(px-x, py-y) > 10 {
				t.Fatalf("no player-sized point near cave-up marker: (%v,%v) found=%v", px, py, found)
			}
		}
		if len(seen) != 8 {
			t.Fatalf("difficulty=%d: selected %d of eight owned eligible chambers", difficulty, len(seen))
		}
	}
}
