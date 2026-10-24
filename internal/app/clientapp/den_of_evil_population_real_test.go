package clientapp

import (
	"strings"
	"testing"
)

// The ordinary clicked cave journey must activate a real level-8 resident in
// authority and project that same identity to the player inside the Den.
func TestDenOfEvilPopulationEntryReal(t *testing.T) {
	app, input, _, _ := newDenPointerClient(t, false, realD2LegacyFixtureConfig{fixtureWorldLevel: 2})
	denFrames(t, app, 10)
	pair := *app.caveWarp
	denMoveTo(t, app, pair.First.X, pair.First.Y, 0.1)
	denClick(t, app, input, pair.First)
	denAssertArrival(t, app, pair.First, 8)
	denFrames(t, app, 2)
	checkpoint := denCheckpoint(t, app)
	ids := checkpointPopulationIDs(t, checkpoint)
	view := coldPopulationWorldView(t, "local-player", checkpoint)
	visible := 0
	for _, entity := range view.Entities {
		if entity.Kind != "hostile" || entity.LevelID != 8 {
			continue
		}
		if !strings.HasPrefix(entity.SpawnID, "level:8:room:") {
			t.Fatalf("Den hostile has unqualified ID %q", entity.SpawnID)
		}
		if _, ok := ids[entity.SpawnID]; !ok {
			t.Fatalf("public Den hostile %q is missing from authority", entity.SpawnID)
		}
		visible++
	}
	if visible == 0 {
		t.Fatal("clicked Den entry has no visible authoritative hostile")
	}
}
