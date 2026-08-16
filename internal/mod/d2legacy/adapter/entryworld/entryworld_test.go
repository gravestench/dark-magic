package entryworld

import (
	"encoding/json"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

func TestPreparedWorldOwnsDestinationAndBootstrapLevelIdentity(t *testing.T) {
	zone, err := worldgen.NewZone(worldgen.Definition{
		Request: worldgen.Request{Version: worldgen.ContractVersion, Seed: 41, Act: 1, LevelID: 7},
		Kind:    "test", Bounds: worldgen.Bounds{Width: 20, Height: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	prepared := &Prepared{
		Worlds: map[int]*gameworld.Map{7: {WidthSubtiles: 100, HeightSubtiles: 50}},
		Zones:  map[int]*worldgen.Zone{7: zone},
		Spawns: map[int][2]float64{7: {11, 12}},
		Seam:   gametransition.Seam{Wilderness: gametransition.SeamEndpoint{LevelID: 7}},
	}
	destination, err := prepared.Destination(7)
	if err != nil {
		t.Fatal(err)
	}
	if destination.X != 11 || destination.Y != 12 || destination.Width != 100 || destination.Height != 50 || destination.LevelID != 7 {
		t.Fatalf("destination = %#v", destination)
	}
	command, err := prepared.PopulationCommand(0)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(command.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["level_id"] != float64(7) || payload["seed"] != float64(41) {
		t.Fatalf("population payload = %#v", payload)
	}
}

func TestPopulationDataPreservesRoomBoundsAndAdjacency(t *testing.T) {
	zone, err := worldgen.NewZone(worldgen.Definition{
		Request: worldgen.Request{Version: worldgen.ContractVersion, Seed: 41, Act: 1, LevelID: 2},
		Kind:    "test", Bounds: worldgen.Bounds{Width: 20, Height: 10},
		Stamps: []worldgen.Stamp{{ID: 1, Width: 20, Height: 10, DS1Path: "blood-moor.ds1", Populate: true}},
		Rooms: []worldgen.Room{
			{ID: 1, X: 2, Y: 3, Width: 4, Height: 5, StampID: 1},
			{ID: 2, X: 6, Y: 3, Width: 4, Height: 5, StampID: 1},
		},
		Links: []worldgen.Link{{From: 1, To: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	gameMap, err := gameworld.NewOpenMap(100, 50)
	if err != nil {
		t.Fatal(err)
	}
	prepared := &Prepared{
		Worlds: map[int]*gameworld.Map{2: gameMap}, Zones: map[int]*worldgen.Zone{2: zone},
		Seam: gametransition.Seam{Wilderness: gametransition.SeamEndpoint{LevelID: 2}},
	}
	population := prepared.PopulationData(0)
	rooms := population["rooms"].([]any)
	first := rooms[0].(map[string]any)
	if first["x"] != float64(10) || first["y"] != float64(15) ||
		first["width"] != float64(20) || first["height"] != float64(25) || first["populate"] != true {
		t.Fatalf("first population room = %#v", first)
	}
	links := population["links"].([]any)
	if len(links) != 1 || links[0].(map[string]any)["from"] != float64(1) || links[0].(map[string]any)["to"] != float64(2) {
		t.Fatalf("population links = %#v", links)
	}
}

func TestInitialDataUsesSharedInteractionAndTransitionContracts(t *testing.T) {
	prepared := &Prepared{Worlds: map[int]*gameworld.Map{}, Difficulty: 2, Seam: gametransition.Seam{
		Town:       gametransition.SeamEndpoint{LevelID: 1, X: 4, Y: 5, Width: 20, Height: 30},
		Wilderness: gametransition.SeamEndpoint{LevelID: 2, X: 6, Y: 7, Width: 40, Height: 50},
	}}
	initial := prepared.InitialData("realm-authority", false)
	interactions := initial["d2legacy.interactions"].(map[string]any)
	if interactions["owner"] != "realm-authority" {
		t.Fatalf("interactions = %#v", interactions)
	}
	transitions := initial["d2legacy.world_transitions"].(map[string]any)
	if len(transitions["seams"].([]any)) != 2 {
		t.Fatalf("transitions = %#v", transitions)
	}
	development := initial["d2legacy.development_items"].(map[string]any)
	if development["enabled"] != false || development["create_empty_containers"] != true {
		t.Fatalf("development = %#v", development)
	}
	rules := initial["d2legacy.game_rules"].(map[string]any)
	if rules["difficulty"] != 2 {
		t.Fatalf("game rules = %#v", rules)
	}
}
