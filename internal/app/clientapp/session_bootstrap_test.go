package clientapp

import (
	"reflect"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// TestSessionInitialDataUsesSelectedDifficulty protects the offline and hosted
// game rules from silently reverting to Normal, which would also change identity.
func TestSessionInitialDataUsesSelectedDifficulty(t *testing.T) {
	t.Parallel()

	for _, difficulty := range []int{0, 1, 2} {
		app := &application{options: Options{GameDifficulty: difficulty}}

		rules, ok := app.sessionInitialData()["d2legacy.game_rules"].(map[string]any)
		if !ok {
			t.Fatal("session initial data has no d2legacy.game_rules map")
		}

		if rules["difficulty"] != difficulty {
			t.Errorf("game_rules difficulty = %v, want %d", rules["difficulty"], difficulty)
		}
	}
}

// TestSessionInitialDataPublishesGeneratedSpawnGeometry keeps portal placement on generated facts.
func TestSessionInitialDataPublishesGeneratedSpawnGeometry(t *testing.T) {
	t.Parallel()

	app := &application{
		gameWorldSpawns: map[int][2]float64{1: {10, 12}, 2: {20, 24}, 3: {30, 36}},
		gameWorlds: map[int]*gameworld.Map{
			1: {WidthSubtiles: 100, HeightSubtiles: 80},
			2: {WidthSubtiles: 200, HeightSubtiles: 160},
		},
	}
	want := map[string]any{
		"1": map[string]any{"x": float64(10), "y": float64(12), "width": float64(100), "height": float64(80)},
		"2": map[string]any{"x": float64(20), "y": float64(24), "width": float64(200), "height": float64(160)},
	}

	got := app.sessionInitialData()["d2legacy.world_spawns"]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spawn geometry = %#v, want %#v", got, want)
	}
}
