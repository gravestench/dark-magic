package mapgen

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// CaveWarpSpec transports Lua's semantic IDs, Levels slots, and LvlWarp identities
// separately. Recipe tile coordinates are not collision anchors.
type CaveWarpSpec struct {
	FirstLevel       int    `json:"first_level"`
	FirstWarpRole    string `json:"first_warp_role"`
	FirstZoneWarpID  int    `json:"first_zone_warp_id"`
	FirstVisSlot     int    `json:"first_vis_slot"`
	FirstLvlWarpID   int    `json:"first_lvlwarp_id"`
	SecondLevel      int    `json:"second_level"`
	SecondWarpRole   string `json:"second_warp_role"`
	SecondZoneWarpID int    `json:"second_zone_warp_id"`
	SecondVisSlot    int    `json:"second_vis_slot"`
	SecondLvlWarpID  int    `json:"second_lvlwarp_id"`
}

type entryWorldSnapshotV4 struct {
	Version  int                   `json:"version"`
	Zones    []worldgen.Definition `json:"zones"`
	Seams    []SeamSpec            `json:"seams"`
	CaveWarp *CaveWarpSpec         `json:"cave_warp"`
}

// GenerateEntryWorld consumes the single Lua composition without selecting native topology.
func GenerateEntryWorld(
	ctx context.Context, source fs.FS, records recordsGateway, seed uint64, difficulty int,
) (EntryWorld, error) {
	if difficulty < 0 || difficulty > 2 {
		return EntryWorld{}, fmt.Errorf("generate d2legacy entry world: difficulty must be 0, 1, or 2")
	}
	ctx = nonNilContext(ctx)
	runtime, err := NewRuntime(ctx, source, records)
	if err != nil {
		return EntryWorld{}, err
	}
	defer func() { _ = runtime.Close(context.Background()) }()
	value, err := modruntime.Call(ctx, runtime.lua, "d2legacy.mapgen.entry_world", "v4", float64(seed), float64(difficulty))
	if err != nil {
		return EntryWorld{}, err
	}
	facts, ok := value.(map[string]any)
	if !ok {
		return EntryWorld{}, fmt.Errorf("entry world v4 is not a table")
	}
	zones, ok := facts["zones"].([]any)
	if !ok {
		return EntryWorld{}, fmt.Errorf("entry world v4 zones are not an array")
	}
	seams, ok := facts["seams"].([]any)
	if !ok {
		return EntryWorld{}, fmt.Errorf("entry world v4 seams are not an array")
	}
	snapshot := entryWorldSnapshotV4{Version: 4}
	for _, raw := range zones {
		encoded, err := json.Marshal(raw)
		if err != nil {
			return EntryWorld{}, err
		}
		var definition worldgen.Definition
		if err = json.Unmarshal(encoded, &definition); err != nil {
			return EntryWorld{}, err
		}
		snapshot.Zones = append(snapshot.Zones, definition)
	}
	for _, raw := range seams {
		seamFacts, ok := raw.(map[string]any)
		if !ok {
			return EntryWorld{}, fmt.Errorf("entry world v4 seam is not a table")
		}
		seam, err := parseSeamFacts(seamFacts)
		if err != nil {
			return EntryWorld{}, err
		}
		snapshot.Seams = append(snapshot.Seams, seam)
	}
	encoded, err := json.Marshal(facts["cave_warp"])
	if err != nil {
		return EntryWorld{}, err
	}
	if err = json.Unmarshal(encoded, &snapshot.CaveWarp); err != nil {
		return EntryWorld{}, err
	}
	encoded, err = json.Marshal(snapshot)
	if err != nil {
		return EntryWorld{}, err
	}
	return restoreEntryWorldV4(encoded)
}

func snapshotEntryWorldV4(world EntryWorld) ([]byte, error) {
	if err := validateEntryWorldV4(world); err != nil {
		return nil, err
	}
	snapshot := entryWorldSnapshotV4{Version: 4, Seams: world.Seams, CaveWarp: world.CaveWarp}
	for _, id := range []int{1, 2, 3, 4, 8} {
		snapshot.Zones = append(snapshot.Zones, world.Zones[id].Definition())
	}
	return json.Marshal(snapshot)
}

func restoreEntryWorldV4(encoded []byte) (EntryWorld, error) {
	var snapshot entryWorldSnapshotV4
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		return EntryWorld{}, err
	}
	ids := []int{1, 2, 3, 4, 8}
	if len(snapshot.Zones) != len(ids) {
		return EntryWorld{}, fmt.Errorf("entry world v4 requires levels 1,2,3,4,8")
	}
	world := EntryWorld{Version: 4, Zones: make(map[int]*worldgen.Zone), Seams: snapshot.Seams, CaveWarp: snapshot.CaveWarp}
	for i, definition := range snapshot.Zones {
		if definition.Request.LevelID != ids[i] {
			return EntryWorld{}, fmt.Errorf("entry world v4 has noncanonical level order")
		}
		zone, err := worldgen.NewZone(definition)
		if err != nil {
			return EntryWorld{}, err
		}
		world.Zones[ids[i]] = zone
	}
	if err := validateEntryWorldV4(world); err != nil {
		return EntryWorld{}, err
	}
	return world, nil
}

func validateEntryWorldV4(world EntryWorld) error {
	if len(world.Zones) != 5 || world.Zones[8] == nil {
		return fmt.Errorf("entry world v4 requires five zones including Den")
	}
	outdoor := EntryWorld{Version: 4, Zones: make(map[int]*worldgen.Zone), Seams: world.Seams}
	for _, id := range []int{1, 2, 3, 4} {
		outdoor.Zones[id] = world.Zones[id]
	}
	if err := validateEntryWorldTopology(outdoor, 4); err != nil {
		return err
	}
	denZone := world.Zones[8]
	if denZone.Bounds() != (worldgen.Bounds{Width: 24, Height: 24}) ||
		len(denZone.Rooms()) != 1 || len(denZone.Stamps()) != 1 {
		return fmt.Errorf("entry world v4 Den must remain one authored 24x24 room")
	}
	den, town := denZone.Request(), world.Zones[1].Request()
	if den.LevelID != 8 || den.Seed != town.Seed || den.Difficulty != town.Difficulty || den.Act != town.Act {
		return fmt.Errorf("entry world v4 Den identity differs")
	}
	return ValidateCaveWarpSpec(world.Zones, world.CaveWarp)
}

// ValidateCaveWarpSpec rechecks the versioned pair at both snapshot and materialization boundaries.
func ValidateCaveWarpSpec(zones map[int]*worldgen.Zone, spec *CaveWarpSpec) error {
	if spec == nil || spec.FirstLevel != 2 || spec.SecondLevel != 8 ||
		spec.FirstWarpRole != "den-entrance" || spec.SecondWarpRole != "den-exit" ||
		spec.FirstVisSlot != 3 || spec.SecondVisSlot != 0 ||
		spec.FirstLvlWarpID != 0 || spec.SecondLvlWarpID != 4 {
		return fmt.Errorf("entry world v4 has invalid cave relationship")
	}
	for _, side := range []struct {
		level, id, target int
		role              string
	}{
		{spec.FirstLevel, spec.FirstZoneWarpID, spec.SecondLevel, spec.FirstWarpRole},
		{spec.SecondLevel, spec.SecondZoneWarpID, spec.FirstLevel, spec.SecondWarpRole},
	} {
		zone := zones[side.level]
		if zone == nil || zone.Request().LevelID != side.level || !seamRoleIsUnique(zone, side.role) {
			return fmt.Errorf("cave level %d has missing or ambiguous endpoint", side.level)
		}
		found := false
		for _, warp := range zone.Warps() {
			if warp.Role == side.role {
				found = int(warp.ID) == side.id && side.id > 0 && warp.DestinationLevel == side.target
			}
		}
		if !found {
			return fmt.Errorf("cave level %d has mismatched endpoint", side.level)
		}
	}
	return nil
}
