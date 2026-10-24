package entryworld

import (
	"encoding/json"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// TestPreparedWorldOwnsDestinationAndBootstrapLevelIdentity verifies that admission and population derive level,
// dimensions, spawn, and seed from the same prepared topology instead of composition-root defaults.
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
		Seams:  []gametransition.Seam{{Second: gametransition.SeamEndpoint{LevelID: 7}}},
	}

	destination, err := prepared.Destination(7)
	if err != nil {
		t.Fatal(err)
	}

	if destination.X != 11 || destination.Y != 12 || destination.Width != 100 ||
		destination.Height != 50 || destination.LevelID != 7 {
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

	zones := payload["zones"].([]any)
	if len(zones) != 1 {
		t.Fatalf("population zones = %#v", zones)
	}
	zonePayload := zones[0].(map[string]any)
	if zonePayload["level_id"] != float64(7) || zonePayload["seed"] != float64(41) {
		t.Fatalf("population payload = %#v", payload)
	}
}

// TestPopulationDataPreservesRoomBoundsAndAdjacency verifies the coordinate conversion and topology identity consumed
// by deterministic resident placement.
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
		Worlds: map[int]*gameworld.Map{2: gameMap},
		Zones:  map[int]*worldgen.Zone{2: zone},
		Seams:  []gametransition.Seam{{Second: gametransition.SeamEndpoint{LevelID: 2}}},
	}
	population, err := prepared.PopulationData(0)
	if err != nil {
		t.Fatal(err)
	}
	zones := population["zones"].([]any)
	if len(zones) != 1 {
		t.Fatalf("population zones = %#v", zones)
	}
	zoneData := zones[0].(map[string]any)
	rooms := zoneData["rooms"].([]any)

	first := rooms[0].(map[string]any)
	if first["x"] != float64(10) || first["y"] != float64(15) ||
		first["width"] != float64(20) || first["height"] != float64(25) || first["populate"] != true {
		t.Fatalf("first population room = %#v", first)
	}

	links := zoneData["links"].([]any)
	if len(links) != 1 {
		t.Fatalf("population links = %#v", links)
	}

	firstLink := links[0].(map[string]any)
	if first["id"] != "1" || firstLink["from"] != "1" || firstLink["to"] != "2" {
		t.Fatalf("population links = %#v", links)
	}
}

// TestInteractionDataAttachesDS1TargetsToGeneratedRooms verifies that decoded selectable identity is enriched with the
// same room and resident IDs used by population state.
func TestInteractionDataAttachesDS1TargetsToGeneratedRooms(t *testing.T) {
	zone, err := worldgen.NewZone(worldgen.Definition{
		Request: worldgen.Request{Version: worldgen.ContractVersion, Seed: 41, Act: 1, LevelID: 2},
		Kind:    "test", Bounds: worldgen.Bounds{Width: 20, Height: 10},
		Rooms: []worldgen.Room{{ID: 7, X: 2, Y: 3, Width: 4, Height: 5}},
	})
	if err != nil {
		t.Fatal(err)
	}

	gameMap, err := gameworld.NewOpenMap(100, 50)
	if err != nil {
		t.Fatal(err)
	}

	gameMap.Objects = []gameworld.Object{{
		Type: gameworld.ObjectTypeStatic, ID: 12, X: 11, Y: 16,
		Description: "Chest", Resolved: true,
	}}
	data := InteractionData(map[int]*gameworld.Map{2: gameMap}, map[int]*worldgen.Zone{2: zone}, "alice", "")

	targets := data["targets"].([]any)
	if len(targets) != 2 {
		t.Fatalf("interaction targets = %#v", targets)
	}

	got := targets[1].(map[string]any)
	if got["id"] != "ds1-object:2:12:0" || got["resident_id"] != "level:2:ds1-object:2:12:0" ||
		got["level_id"] != float64(2) || got["room_id"] != "7" {
		t.Fatalf("resident interaction target = %#v", got)
	}
}

// TestInitialDataUsesSharedInteractionAndTransitionContracts verifies that session bootstrap publishes owner,
// development, difficulty, and bidirectional seam policy without client-specific rewrites.
func TestInitialDataUsesSharedInteractionAndTransitionContracts(t *testing.T) {
	seam := gametransition.Seam{
		First:  gametransition.SeamEndpoint{LevelID: 1, X: 4, Y: 5, Width: 20, Height: 30},
		Second: gametransition.SeamEndpoint{LevelID: 2, X: 6, Y: 7, Width: 40, Height: 50},
	}
	seam.TriggerRadius = 2
	prepared := &Prepared{
		Worlds: map[int]*gameworld.Map{}, Difficulty: 2, Seams: []gametransition.Seam{seam},
		TopologyChecksum: "saved-entry-world",
	}
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
	if got := initial["d2legacy.entry_world_checksum"]; got != prepared.TopologyChecksum {
		t.Fatalf("entry world checksum = %v, want %q", got, prepared.TopologyChecksum)
	}
}

func TestInteractionDataPreservesAuthoredTownNPCPlacements(t *testing.T) {
	town, err := gameworld.NewOpenMap(100, 100)
	if err != nil {
		t.Fatal(err)
	}
	town.Objects = []gameworld.Object{
		{Type: gameworld.ObjectTypeDynamic, ID: 2, X: 21, Y: 46, Class: "akara", Resolved: true},
		{Type: gameworld.ObjectTypeStatic, ID: 12, X: 11, Y: 16, Description: "Chest", Resolved: true},
		{Type: gameworld.ObjectTypeDynamic, ID: 8, X: 31, Y: 56, Class: "charsi", Resolved: true},
	}
	data := InteractionData(map[int]*gameworld.Map{1: town}, nil, "alice", "")
	targets := data["targets"].([]any)
	if len(targets) != 3 {
		t.Fatalf("Akara must have one target, got %#v", targets)
	}
	guide := targets[0].(map[string]any)
	if guide["id"] != "act1-akara" || guide["x"] != float64(21) || guide["y"] != float64(46) ||
		guide["act"] != float64(1) ||
		guide["level_id"] != float64(1) || guide["selectable"] != true || guide["radius"] != float64(4) {
		t.Fatalf("authored Akara target = %#v", guide)
	}
	if other := targets[1].(map[string]any); other["npc"] != "Chest" || other["x"] != float64(11) || other["y"] != float64(16) {
		t.Fatalf("DS1 Akara must not be a second target: %#v", other)
	}

	charsi := targets[2].(map[string]any)
	if charsi["class"] != "charsi" || charsi["level_id"] != float64(1) ||
		charsi["x"] != float64(31) || charsi["y"] != float64(56) {
		t.Fatalf("authored NPC facts = %#v", charsi)
	}
}

// Missing or mismatched prepared facts must never produce an admissible command.
func TestPopulationDataOrdersAdmittedLevelsAndRejectsMissingMap(t *testing.T) {
	prepared := populationPreparedFixture(t)
	data, err := prepared.PopulationData(0)
	if err != nil {
		t.Fatal(err)
	}
	zones := data["zones"].([]any)
	if len(zones) != 2 || zones[0].(map[string]any)["level_id"] != float64(2) || zones[1].(map[string]any)["level_id"] != float64(3) {
		t.Fatalf("zones=%#v", zones)
	}
	for _, bad := range []string{"map", "zone", "duplicate", "identity"} {
		t.Run(bad, func(t *testing.T) {
			p := populationPreparedFixture(t)
			switch bad {
			case "map":
				delete(p.Worlds, 3)
			case "zone":
				delete(p.Zones, 3)
			case "duplicate":
				p.Seams = append(p.Seams, p.Seams[1])
			case "identity":
				p.Zones[3] = p.Zones[2]
			}
			if data, err := p.PopulationData(0); err == nil || data != nil {
				t.Fatalf("data=%#v err=%v", data, err)
			}
			if command, err := p.PopulationCommand(0); err == nil || command.Kind != "" {
				t.Fatalf("command=%#v err=%v", command, err)
			}
		})
	}
}

func TestPopulationDataV1KeepsOnlyLevelTwo(t *testing.T) {
	p := populationPreparedFixture(t)
	p.Seams = p.Seams[:1]
	delete(p.Zones, 3)
	delete(p.Worlds, 3)
	data, err := p.PopulationData(1)
	if err != nil {
		t.Fatal(err)
	}
	zones := data["zones"].([]any)
	if len(zones) != 1 {
		t.Fatalf("zones=%#v", zones)
	}
	points := zones[0].(map[string]any)["rooms"].([]any)[0].(map[string]any)["points"].([]any)
	for i, want := range [][2]float64{{20, 10}, {17, 17}, {10, 20}, {3, 17}} {
		point := points[i].(map[string]any)
		if point["x"] != want[0] || point["y"] != want[1] {
			t.Fatalf("point %d=%#v", i, point)
		}
	}
}

func TestPopulationDataLevelThreePointsStayDistinctAndInRoom(t *testing.T) {
	p := populationPreparedFixture(t)
	for _, blocked := range []bool{false, true} {
		if blocked {
			p.Worlds[3] = &gameworld.Map{WidthSubtiles: 100, HeightSubtiles: 100}
		}
		data, err := p.PopulationData(1)
		if err != nil {
			t.Fatal(err)
		}
		room := data["zones"].([]any)[1].(map[string]any)["rooms"].([]any)[0].(map[string]any)
		points := room["points"].([]any)
		if blocked && len(points) != 0 {
			t.Fatalf("blocked points=%#v", points)
		}
		if !blocked && len(points) < 2 {
			t.Fatalf("open points=%#v", points)
		}
		seen := map[[2]float64]bool{}
		for _, raw := range points {
			pt := raw.(map[string]any)
			xy := [2]float64{pt["x"].(float64), pt["y"].(float64)}
			if seen[xy] || xy[0] < 25 || xy[0] >= 50 || xy[1] < 25 || xy[1] >= 50 {
				t.Fatalf("invalid room-local point=%v", xy)
			}
			seen[xy] = true
		}
	}
}

func populationPreparedFixture(t *testing.T) *Prepared {
	t.Helper()
	p := &Prepared{Worlds: map[int]*gameworld.Map{}, Zones: map[int]*worldgen.Zone{}, Spawns: map[int][2]float64{2: {10, 10}}}
	for _, id := range []int{2, 3} {
		zone, err := worldgen.NewZone(worldgen.Definition{Request: worldgen.Request{Version: worldgen.ContractVersion, Seed: 41, Act: 1, LevelID: id}, Kind: "test", Bounds: worldgen.Bounds{Width: 20, Height: 20}, Rooms: []worldgen.Room{{ID: 1, X: 5, Y: 5, Width: 5, Height: 5}}})
		if err != nil {
			t.Fatal(err)
		}
		m, err := gameworld.NewOpenMap(100, 100)
		if err != nil {
			t.Fatal(err)
		}
		p.Zones[id] = zone
		p.Worlds[id] = m
		p.Seams = append(p.Seams, gametransition.Seam{Second: gametransition.SeamEndpoint{LevelID: id}})
	}
	return p
}
