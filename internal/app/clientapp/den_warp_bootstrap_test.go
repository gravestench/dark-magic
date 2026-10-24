package clientapp

import (
	"reflect"
	"testing"

	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// Dropping the prepared pair between generation and initial data must fail here.
func TestDenOfEvilProductionBootstrapPublishesPreparedPair(t *testing.T) {
	app := &application{}
	prepared := &entryworld.Prepared{Seams: []gametransition.Seam{{}}, CaveWarp: &entryworld.CaveWarpPair{
		First: entryworld.CaveWarpEndpoint{ID: "static-warp:2:3", PairedID: "static-warp:8:1", LevelID: 2, TargetLevel: 8,
			X: 11, Y: 12, ArrivalX: 21, ArrivalY: 22, Width: 120, Height: 120},
		Second: entryworld.CaveWarpEndpoint{ID: "static-warp:8:1", PairedID: "static-warp:2:3", LevelID: 8, TargetLevel: 2,
			X: 23, Y: 24, ArrivalX: 13, ArrivalY: 14, Width: 640, Height: 480},
	}}
	app.publishEntryWorld(prepared, "topology", nil, 1, nil)
	got := app.sessionInitialData()["d2legacy.world_warps"]
	if !reflect.DeepEqual(got, prepared.WarpData()) {
		t.Fatalf("production warp bootstrap = %#v, want prepared pair %#v", got, prepared.WarpData())
	}
}

// The production composition must create exactly the two prepared caves and
// leave portal appearance ownership with the authored map art.
func TestDenOfEvilProductionAuthorityCreatesTokenlessPairReal(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "game_world", fixtureCharacters: 1, applySceneDefaults: true,
	})
	app := fixture.app
	data := app.sessionInitialData()["d2legacy.world_warps"].(map[string]any)
	endpoints := data["endpoints"].([]any)
	if len(endpoints) != 2 {
		t.Fatalf("production warp endpoints = %d, want 2", len(endpoints))
	}
	want := map[string]string{"static-warp:2:3": "static-warp:8:1", "static-warp:8:1": "static-warp:2:3"}
	for _, raw := range endpoints {
		endpoint := raw.(map[string]any)
		id := endpoint["id"].(string)
		if endpoint["pair_id"] != want[id] || endpoint["static"] != true {
			t.Fatalf("invalid production cave endpoint %#v", endpoint)
		}
		if _, found := endpoint["token"]; found {
			t.Fatalf("cave endpoint %s requested portal appearance", id)
		}
	}
	warps := requireRealStore(t, fixture, "d2legacy.world.warp", "production warp store absent")
	selectables := requireRealStore(t, fixture, "d2legacy.world.selectable", "production selectable store absent")
	appearance := requireRealStore(t, fixture, "d2legacy.world.warp_appearance", "portal appearance store absent")
	if warps.Len() != 2 {
		t.Fatalf("production warp entities = %d, want 2", warps.Len())
	}
	for _, entity := range warps.Entities() {
		selectable, found := selectables.Get(entity)
		if !found {
			t.Fatal("cave warp has no selectable")
		}
		id, _ := selectable.Get("id")
		if _, found := want[id.(string)]; !found {
			t.Fatalf("unexpected production warp %v", id)
		}
		if _, found := appearance.Get(entity); found {
			t.Fatalf("cave warp %s has portal appearance", id)
		}
	}
}

func TestDenOfEvilWarpLabRetainsPortalTokensReal(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene: "warp_lab", applySceneDefaults: true,
	})
	data := fixture.app.sessionInitialData()["d2legacy.world_warps"].(map[string]any)
	endpoints := data["endpoints"].([]any)
	if len(endpoints) != 2 {
		t.Fatalf("Warp Lab endpoints = %d, want 2", len(endpoints))
	}
	want := map[string]string{"warp-lab:town": "TP", "warp-lab:wilderness": "PP"}
	for _, raw := range endpoints {
		endpoint := raw.(map[string]any)
		id := endpoint["id"].(string)
		if endpoint["token"] != want[id] || endpoint["static"] == true {
			t.Fatalf("Warp Lab portal %s lost appearance token: %#v", id, endpoint)
		}
	}
	appearance := requireRealStore(t, fixture, "d2legacy.world.warp_appearance", "portal appearance store absent")
	if appearance.Len() != 2 {
		t.Fatalf("Warp Lab portal appearances = %d, want 2", appearance.Len())
	}
}
