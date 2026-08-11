package missile

import (
	"testing"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	models "github.com/gravestench/dark-magic/internal/game/data/model"
)

func TestPresentationFromCatalogCopiesLegacyVisualAndSoundFacts(t *testing.T) {
	snapshot := gamedata.Snapshot{MissilesByName: map[string]models.Missile{"firebolt": {
		Missile: "firebolt", CelFile: "firebolt", AnimSpeed: 16, NumDirections: 8,
		LoopAnim: 1, TravelSound: "firebolt_travel", HitSound: "firebolt_hit",
		XOffset: 2, YOffset: 3, ZOffset: 4,
	}}}
	result, err := PresentationFromCatalog(snapshot, "firebolt")
	if err != nil {
		t.Fatal(err)
	}
	if result.DCC != "data/global/missiles/firebolt.dcc" || result.FramesPerSecond != 25 || result.Directions != 8 || !result.Loop || result.HitSound != "firebolt_hit" || result.OffsetZ != 4 {
		t.Fatalf("presentation = %#v", result)
	}
}
