package mapgen

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// EntryWorld is the admitted level topology and mod-authored seam policy.
type EntryWorld struct {
	Zones   map[int]*worldgen.Zone
	Seams   []SeamSpec
	Version int
}

// entryWorldSnapshot is the versioned durable envelope. Definitions are stored rather than live zones so restoration
// reruns generic validation without rerunning Lua policy.
type entryWorldSnapshot struct {
	Version    int                 `json:"version"`
	Town       worldgen.Definition `json:"town"`
	Wilderness worldgen.Definition `json:"wilderness"`
	Seam       SeamSpec            `json:"seam"`
}

// v2 fixes array order so equivalent three-zone topologies produce identical snapshot bytes and checksums.
type entryWorldSnapshotV2 struct {
	Version int                   `json:"version"`
	Zones   []worldgen.Definition `json:"zones"`
	Seams   []SeamSpec            `json:"seams"`
}

// v3 preserves the four admitted recipes and three seams in canonical level order.
type entryWorldSnapshotV3 struct {
	Version int                   `json:"version"`
	Zones   []worldgen.Definition `json:"zones"`
	Seams   []SeamSpec            `json:"seams"`
}

// SeamSpec is serialized policy chosen by d2legacy. Coordinates remain in tile space until the generic transition
// adapter resolves materialized maps.
type SeamSpec struct {
	FirstLevel, SecondLevel         int
	FirstDirection, SecondDirection string
	SecondTileX, SecondTileY        int
	FirstTileX, FirstTileY          int    `json:",omitempty"`
	FirstWarpID, SecondWarpID       int    `json:",omitempty"`
	FirstWarpRole, SecondWarpRole   string `json:",omitempty"`
	FirstGenerated                  bool   `json:",omitempty"`
	TriggerRadius                   int    `json:",omitempty"`
	MaxArrivalDistance              int    `json:",omitempty"`
	ArrivalInset                    int    `json:",omitempty"`
}

// Snapshot returns the canonical durable envelope for the joined entry topology. It stores admitted recipes and their
// Lua-authored seam, not transient generator or runtime state.
func (world EntryWorld) Snapshot() ([]byte, error) {
	if world.Version == 3 {
		if err := validateEntryWorldTopology(world, 4); err != nil {
			return nil, err
		}
		return json.Marshal(entryWorldSnapshotV3{
			Version: 3,
			Zones: []worldgen.Definition{
				world.Zones[1].Definition(), world.Zones[2].Definition(),
				world.Zones[3].Definition(), world.Zones[4].Definition(),
			},
			Seams: world.Seams,
		})
	}
	if world.Version == 2 {
		if err := validateEntryWorldV2(world); err != nil {
			return nil, err
		}
		return json.Marshal(entryWorldSnapshotV2{
			Version: 2,
			Zones: []worldgen.Definition{
				world.Zones[1].Definition(), world.Zones[2].Definition(), world.Zones[3].Definition(),
			},
			Seams: world.Seams,
		})
	}
	if world.Version != 1 || len(world.Zones) != 2 || len(world.Seams) != 1 ||
		world.Zones[1] == nil || world.Zones[2] == nil {
		return nil, fmt.Errorf("d2legacy entry world is incomplete")
	}

	return json.Marshal(entryWorldSnapshot{
		Version:    1,
		Town:       world.Zones[1].Definition(),
		Wilderness: world.Zones[2].Definition(),
		Seam:       world.Seams[0],
	})
}

// Checksum hashes the canonical snapshot bytes so authorities and reconnecting clients can compare complete entry
// topology without depending on pointer identity or generator internals.
func (world EntryWorld) Checksum() (string, error) {
	encoded, err := world.Snapshot()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha256.Sum256(encoded)), nil
}

// RestoreEntryWorld reconstructs validated immutable zones without rerunning Diablo selection policy, exactly as a
// checkpoint or reconnect path requires. Seam level IDs are checked after zone validation to prevent mismatched joins.
func RestoreEntryWorld(encoded []byte) (EntryWorld, error) {
	var header struct {
		Version int `json:"version"`
	}
	if err := json.Unmarshal(encoded, &header); err != nil {
		return EntryWorld{}, err
	}
	if header.Version == 3 {
		return restoreEntryWorldV3(encoded)
	}
	if header.Version == 2 {
		return restoreEntryWorldV2(encoded)
	}

	var snapshot entryWorldSnapshot
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		return EntryWorld{}, err
	}

	if snapshot.Version != 1 {
		return EntryWorld{}, fmt.Errorf("unsupported entry-world snapshot version %d", snapshot.Version)
	}

	town, err := worldgen.NewZone(snapshot.Town)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("restore entry town: %w", err)
	}

	wilderness, err := worldgen.NewZone(snapshot.Wilderness)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("restore entry wilderness: %w", err)
	}

	if town.Request().LevelID != snapshot.Seam.FirstLevel ||
		wilderness.Request().LevelID != snapshot.Seam.SecondLevel {
		return EntryWorld{}, fmt.Errorf("restore entry world: seam level identities do not match recipes")
	}

	return EntryWorld{
		Zones: map[int]*worldgen.Zone{1: town, 2: wilderness}, Seams: []SeamSpec{snapshot.Seam}, Version: 1,
	}, nil
}

// restoreEntryWorldV2 keeps saved Lua decisions intact while re-admitting every recipe and seam identity.
func restoreEntryWorldV2(encoded []byte) (EntryWorld, error) {
	var snapshot entryWorldSnapshotV2
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		return EntryWorld{}, err
	}
	if len(snapshot.Zones) != 3 || len(snapshot.Seams) != 2 {
		return EntryWorld{}, fmt.Errorf("restore entry world: v2 requires three zones and two seams")
	}
	zones := make(map[int]*worldgen.Zone, 3)
	for index, definition := range snapshot.Zones {
		zone, err := worldgen.NewZone(definition)
		if err != nil {
			return EntryWorld{}, fmt.Errorf("restore entry zone %d: %w", index, err)
		}
		level := zone.Request().LevelID
		if level != index+1 {
			return EntryWorld{}, fmt.Errorf("restore entry world: zones are not in canonical level order")
		}
		zones[level] = zone
	}
	world := EntryWorld{Zones: zones, Seams: snapshot.Seams, Version: 2}
	if err := validateEntryWorldV2(world); err != nil {
		return EntryWorld{}, err
	}
	return world, nil
}

// restoreEntryWorldV3 re-admits saved definitions without consulting current Lua policy.
func restoreEntryWorldV3(encoded []byte) (EntryWorld, error) {
	var snapshot entryWorldSnapshotV3
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		return EntryWorld{}, err
	}
	if len(snapshot.Zones) != 4 || len(snapshot.Seams) != 3 {
		return EntryWorld{}, fmt.Errorf("restore entry world: v3 requires four zones and three seams")
	}
	zones := make(map[int]*worldgen.Zone, 4)
	for index, definition := range snapshot.Zones {
		zone, err := worldgen.NewZone(definition)
		if err != nil {
			return EntryWorld{}, fmt.Errorf("restore entry zone %d: %w", index, err)
		}
		level := zone.Request().LevelID
		if level != index+1 {
			return EntryWorld{}, fmt.Errorf("restore entry world: zones are not in canonical level order")
		}
		zones[level] = zone
	}
	world := EntryWorld{Zones: zones, Seams: snapshot.Seams, Version: 3}
	if err := validateEntryWorldTopology(world, 4); err != nil {
		return EntryWorld{}, err
	}
	return world, nil
}

// validateEntryWorldV2 binds each policy-selected edge to exactly one admitted warp in canonical level order.
func validateEntryWorldV2(world EntryWorld) error {
	return validateEntryWorldTopology(world, 3)
}

// validateEntryWorldTopology checks shared identities and binds every generated side to one admitted warp.
func validateEntryWorldTopology(world EntryWorld, levelCount int) error {
	if len(world.Zones) != levelCount || len(world.Seams) != levelCount-1 {
		return fmt.Errorf("d2legacy entry world: v%d requires %d zones and %d seams", world.Version, levelCount, levelCount-1)
	}
	var seed uint64
	var difficulty int
	for level := 1; level <= levelCount; level++ {
		zone := world.Zones[level]
		if zone == nil || zone.Request().LevelID != level {
			return fmt.Errorf("d2legacy entry world: missing or mismatched zone %d", level)
		}
		request := zone.Request()
		if level == 1 {
			seed, difficulty = request.Seed, int(request.Difficulty)
		} else if request.Seed != seed || int(request.Difficulty) != difficulty {
			return fmt.Errorf("d2legacy entry world: seed or difficulty differs between zones")
		}
	}
	for index, seam := range world.Seams {
		if seam.FirstLevel != index+1 || seam.SecondLevel != index+2 ||
			(index == 0 && seam.FirstGenerated) || (index > 0 && !seam.FirstGenerated) {
			return fmt.Errorf("d2legacy entry world: duplicate or mismatched seam %d", index)
		}
		if !oppositeSeamDirections(seam.FirstDirection, seam.SecondDirection) ||
			seam.TriggerRadius <= 0 || seam.MaxArrivalDistance <= 0 || seam.ArrivalInset <= seam.TriggerRadius {
			return fmt.Errorf("d2legacy entry world: invalid seam geometry %d", index)
		}
		if seam.FirstGenerated && !seamTileInside(world.Zones[seam.FirstLevel], seam.FirstTileX, seam.FirstTileY) {
			return fmt.Errorf("d2legacy entry world: first seam tile is outside zone %d", seam.FirstLevel)
		}
		if !seamTileInside(world.Zones[seam.SecondLevel], seam.SecondTileX, seam.SecondTileY) {
			return fmt.Errorf("d2legacy entry world: second seam tile is outside zone %d", seam.SecondLevel)
		}
		// V3 preserves Lua's unique-role binding even when duplicate endpoints use different IDs.
		// Older envelopes retain their original ID-based acceptance contract.
		if world.Version == 3 {
			if !seamRoleIsUnique(world.Zones[seam.SecondLevel], seam.SecondWarpRole) ||
				(seam.FirstGenerated && !seamRoleIsUnique(world.Zones[seam.FirstLevel], seam.FirstWarpRole)) {
				return fmt.Errorf("d2legacy entry world: seam %d has an ambiguous endpoint role", index)
			}
		}
		if !seamMatchesWarp(world.Zones[seam.SecondLevel], seam.SecondWarpID, seam.SecondWarpRole,
			seam.FirstLevel, seam.SecondDirection, seam.SecondTileX, seam.SecondTileY) {
			return fmt.Errorf("d2legacy entry world: seam %d does not match admitted second-zone warp", index)
		}
		if seam.FirstGenerated && !seamMatchesWarp(world.Zones[seam.FirstLevel], seam.FirstWarpID,
			seam.FirstWarpRole, seam.SecondLevel, seam.FirstDirection, seam.FirstTileX, seam.FirstTileY) {
			return fmt.Errorf("d2legacy entry world: seam %d does not match admitted first-zone warp", index)
		}
	}
	return nil
}

// seamRoleIsUnique binds a semantic endpoint independently of its numeric ID, destination, or coordinates.
func seamRoleIsUnique(zone *worldgen.Zone, role string) bool {
	matches := 0

	for _, warp := range zone.Warps() {
		if warp.Role == role {
			matches++
		}
	}

	return matches == 1
}

// seamMatchesWarp rejects stale or ambiguous warp claims after a snapshot has been restored.
func seamMatchesWarp(zone *worldgen.Zone, warpID int, role string, destination int, direction string, x, y int) bool {
	if warpID <= 0 || role == "" {
		return false
	}
	matches := 0
	for _, warp := range zone.Warps() {
		if uint64(warp.ID) != uint64(warpID) {
			continue
		}
		matches++
		if warp.Role != role || warp.DestinationLevel != destination ||
			warp.Direction != direction || warp.X != x || warp.Y != y {
			return false
		}
	}
	return matches == 1
}

// seamTileInside checks tile coordinates against the admitted recipe before materialization.
func seamTileInside(zone *worldgen.Zone, x, y int) bool {
	bounds := zone.Bounds()
	return x >= bounds.X && y >= bounds.Y && x < bounds.X+bounds.Width && y < bounds.Y+bounds.Height
}

// oppositeSeamDirections rejects noncardinal and mismatched edge pairs together.
func oppositeSeamDirections(first, second string) bool {
	return map[string]string{"north": "south", "east": "west", "south": "north", "west": "east"}[first] == second
}

// GenerateEntryWorld asks d2legacy for the first four recipes and three policy-owned seams. The headless runtime
// prevents authoritative policy from observing client state; all recipes are admitted before the topology is returned.
func GenerateEntryWorld(
	ctx context.Context,
	source fs.FS,
	records recordsGateway,
	seed uint64,
	difficulty int,
) (EntryWorld, error) {
	if difficulty < 0 || difficulty > 2 {
		return EntryWorld{}, fmt.Errorf(
			"generate d2legacy entry world: difficulty must be 0, 1, or 2",
		)
	}

	ctx = nonNilContext(ctx)

	runtime, err := NewRuntime(ctx, source, records)
	if err != nil {
		return EntryWorld{}, err
	}

	// Generation returns its own errors, so runtime shutdown remains best effort and must not replace those results.
	defer func() {
		_ = runtime.Close(context.Background())
	}()

	town, err := runtime.generateFrom(
		ctx,
		"d2legacy.mapgen.entry_world",
		"town",
		float64(seed),
		float64(difficulty),
	)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("generate d2legacy entry town: %w", err)
	}

	townFacts, err := zoneFacts(town, "town")
	if err != nil {
		return EntryWorld{}, err
	}

	wilderness, err := runtime.generateFrom(
		ctx,
		"d2legacy.mapgen.entry_world",
		"wilderness",
		float64(seed),
		float64(difficulty),
		townFacts,
	)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("generate d2legacy entry wilderness: %w", err)
	}

	wildernessFacts, err := zoneFacts(wilderness, "wilderness")
	if err != nil {
		return EntryWorld{}, err
	}

	seam, err := describeEntrySeam(ctx, runtime, townFacts, wildernessFacts)
	if err != nil {
		return EntryWorld{}, err
	}

	plains, err := runtime.generateFrom(
		ctx, "d2legacy.mapgen.entry_world", "cold_plains_v3",
		float64(seed), float64(difficulty), wildernessFacts,
	)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("generate d2legacy Cold Plains: %w", err)
	}
	plainsFacts, err := zoneFacts(plains, "Cold Plains")
	if err != nil {
		return EntryWorld{}, err
	}
	plainsSeam, err := describeGeneratedSeam(ctx, runtime, "generated_seam", wildernessFacts, plainsFacts)
	if err != nil {
		return EntryWorld{}, err
	}
	stony, err := runtime.generateFrom(
		ctx, "d2legacy.mapgen.entry_world", "stony_field",
		float64(seed), float64(difficulty), plainsFacts,
	)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("generate d2legacy Stony Field: %w", err)
	}
	stonyFacts, err := zoneFacts(stony, "Stony Field")
	if err != nil {
		return EntryWorld{}, err
	}
	stonySeam, err := describeGeneratedSeam(ctx, runtime, "stony_field_seam", plainsFacts, stonyFacts)
	if err != nil {
		return EntryWorld{}, err
	}
	world := EntryWorld{
		Zones: map[int]*worldgen.Zone{1: town, 2: wilderness, 3: plains, 4: stony},
		Seams: []SeamSpec{seam, plainsSeam, stonySeam}, Version: 3,
	}
	if err := validateEntryWorldTopology(world, 4); err != nil {
		return EntryWorld{}, err
	}
	return world, nil
}

// zoneFacts round-trips an admitted zone through its JSON contract before passing it back to Lua. This guarantees seam
// policy observes exactly the serializable recipe that snapshots and remote authorities will later consume.
func zoneFacts(zone *worldgen.Zone, label string) (map[string]any, error) {
	encoded, err := zone.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("encode d2legacy entry %s: %w", label, err)
	}

	var facts map[string]any
	if err := json.Unmarshal(encoded, &facts); err != nil {
		return nil, fmt.Errorf("decode d2legacy entry %s facts: %w", label, err)
	}

	return facts, nil
}

// describeEntrySeam calls the policy module only after both zone definitions are admitted, then converts the returned
// table into a typed seam without introducing native layout decisions.
func describeEntrySeam(
	ctx context.Context,
	runtime *Runtime,
	townFacts map[string]any,
	wildernessFacts map[string]any,
) (SeamSpec, error) {
	value, err := modruntime.Call(
		ctx,
		runtime.lua,
		"d2legacy.mapgen.entry_world",
		"seam",
		townFacts,
		wildernessFacts,
	)
	if err != nil {
		return SeamSpec{}, fmt.Errorf("describe d2legacy entry seam: %w", err)
	}

	facts, ok := value.(map[string]any)
	if !ok {
		return SeamSpec{}, fmt.Errorf("describe d2legacy entry seam: result is not a table")
	}

	return parseSeamFacts(facts)
}

// describeGeneratedSeam passes admitted recipes back to the selected Lua policy to bind their edges and warp roles.
func describeGeneratedSeam(
	ctx context.Context,
	runtime *Runtime,
	function string,
	firstFacts map[string]any,
	secondFacts map[string]any,
) (SeamSpec, error) {
	value, err := modruntime.Call(
		ctx, runtime.lua, "d2legacy.mapgen.entry_world", function, firstFacts, secondFacts,
	)
	if err != nil {
		return SeamSpec{}, fmt.Errorf("describe generated d2legacy seam: %w", err)
	}
	facts, ok := value.(map[string]any)
	if !ok {
		return SeamSpec{}, fmt.Errorf("describe generated d2legacy seam: result is not a table")
	}
	return parseSeamFacts(facts)
}

// parseSeamFacts validates each policy-owned field before publishing a typed seam. Field-specific numeric failures are
// retained because they identify the exact Lua contract violation.
func parseSeamFacts(facts map[string]any) (SeamSpec, error) {
	firstLevel, err := seamNumber(facts, "first_level")
	if err != nil {
		return SeamSpec{}, err
	}

	secondLevel, err := seamNumber(facts, "second_level")
	if err != nil {
		return SeamSpec{}, err
	}

	secondX, err := seamNumber(facts, "second_tile_x")
	if err != nil {
		return SeamSpec{}, err
	}

	secondY, err := seamNumber(facts, "second_tile_y")
	if err != nil {
		return SeamSpec{}, err
	}

	firstDirection, firstOK := facts["first_direction"].(string)

	secondDirection, secondOK := facts["second_direction"].(string)
	if !firstOK || !secondOK {
		return SeamSpec{}, fmt.Errorf("describe d2legacy entry seam: directions are not strings")
	}
	firstX, firstY, firstWarpID := 0, 0, 0
	firstWarpRole := ""
	firstGenerated, _ := facts["first_generated"].(bool)
	if firstGenerated {
		firstX, err = seamNumber(facts, "first_tile_x")
		if err != nil {
			return SeamSpec{}, err
		}
		firstY, err = seamNumber(facts, "first_tile_y")
		if err != nil {
			return SeamSpec{}, err
		}
		firstWarpID, err = seamNumber(facts, "first_warp_id")
		if err != nil {
			return SeamSpec{}, err
		}
		firstWarpRole, _ = facts["first_warp_role"].(string)
		if firstWarpRole == "" {
			return SeamSpec{}, fmt.Errorf("describe d2legacy entry seam: first warp role is absent")
		}
	}
	secondWarpID, err := seamNumber(facts, "second_warp_id")
	if err != nil {
		return SeamSpec{}, err
	}
	secondWarpRole, _ := facts["second_warp_role"].(string)
	if secondWarpRole == "" {
		return SeamSpec{}, fmt.Errorf("describe d2legacy entry seam: second warp role is absent")
	}
	triggerRadius, err := seamNumber(facts, "trigger_radius")
	if err != nil {
		return SeamSpec{}, err
	}
	maxArrivalDistance, err := seamNumber(facts, "max_arrival_distance")
	if err != nil {
		return SeamSpec{}, err
	}
	arrivalInset, err := seamNumber(facts, "arrival_inset")
	if err != nil {
		return SeamSpec{}, err
	}

	return SeamSpec{
		FirstLevel:         firstLevel,
		FirstDirection:     firstDirection,
		SecondLevel:        secondLevel,
		SecondDirection:    secondDirection,
		SecondTileX:        secondX,
		SecondTileY:        secondY,
		FirstTileX:         firstX,
		FirstTileY:         firstY,
		FirstWarpID:        firstWarpID,
		SecondWarpID:       secondWarpID,
		FirstWarpRole:      firstWarpRole,
		SecondWarpRole:     secondWarpRole,
		FirstGenerated:     firstGenerated,
		TriggerRadius:      triggerRadius,
		MaxArrivalDistance: maxArrivalDistance,
		ArrivalInset:       arrivalInset,
	}, nil
}

// seamNumber reads Lua's numeric representation and preserves the adapter's historical integer conversion semantics.
func seamNumber(facts map[string]any, name string) (int, error) {
	value, ok := facts[name].(float64)
	if !ok {
		return 0, fmt.Errorf("describe d2legacy entry seam: %s is not numeric", name)
	}

	return int(value), nil
}
