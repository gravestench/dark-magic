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

// EntryWorld contains admitted entry-area recipes and their mod-authored connections.
type EntryWorld struct {
	Town, Wilderness  *worldgen.Zone
	Seam              SeamSpec
	Dungeon           *worldgen.Zone
	DungeonConnection *ConnectionSpec
}

// entryWorldSnapshot is the versioned durable envelope. Definitions are stored rather than live zones so restoration
// reruns generic validation without rerunning Lua policy.
type entryWorldSnapshot struct {
	Version           int                  `json:"version"`
	Town              worldgen.Definition  `json:"town"`
	Wilderness        worldgen.Definition  `json:"wilderness"`
	Seam              SeamSpec             `json:"seam"`
	Dungeon           *worldgen.Definition `json:"dungeon,omitempty"`
	DungeonConnection *ConnectionSpec      `json:"dungeon_connection,omitempty"`
}

// ConnectionSpec joins authored exit tiles within two policy-selected stamps.
type ConnectionSpec struct {
	FirstLevel, SecondLevel int
	FirstStamp, SecondStamp uint32
}

// SeamSpec is serialized policy chosen by d2legacy. Coordinates remain in tile space until the generic transition
// adapter resolves materialized maps.
type SeamSpec struct {
	FirstLevel, SecondLevel         int
	FirstDirection, SecondDirection string
	SecondTileX, SecondTileY        int
}

// Snapshot returns the canonical durable envelope for the joined entry topology. It stores admitted recipes and their
// Lua-authored seam, not transient generator or runtime state.
func (world EntryWorld) Snapshot() ([]byte, error) {
	if world.Town == nil || world.Wilderness == nil || (world.Dungeon == nil) != (world.DungeonConnection == nil) {
		return nil, fmt.Errorf("d2legacy entry world is incomplete")
	}

	var dungeon *worldgen.Definition
	if world.Dungeon != nil {
		definition := world.Dungeon.Definition()
		dungeon = &definition
	}
	return json.Marshal(entryWorldSnapshot{
		Version:           1,
		Town:              world.Town.Definition(),
		Wilderness:        world.Wilderness.Definition(),
		Seam:              world.Seam,
		Dungeon:           dungeon,
		DungeonConnection: world.DungeonConnection,
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

	result := EntryWorld{Town: town, Wilderness: wilderness, Seam: snapshot.Seam}
	if (snapshot.Dungeon == nil) != (snapshot.DungeonConnection == nil) {
		return EntryWorld{}, fmt.Errorf("restore entry world: incomplete dungeon connection")
	}
	if snapshot.Dungeon != nil {
		result.Dungeon, err = worldgen.NewZone(*snapshot.Dungeon)
		if err != nil {
			return EntryWorld{}, fmt.Errorf("restore entry dungeon: %w", err)
		}
		result.DungeonConnection = snapshot.DungeonConnection
		if wilderness.Request().LevelID != result.DungeonConnection.FirstLevel ||
			result.Dungeon.Request().LevelID != result.DungeonConnection.SecondLevel ||
			result.Dungeon.Request().LevelID == town.Request().LevelID ||
			result.Dungeon.Request().LevelID == wilderness.Request().LevelID ||
			!hasConnectionStamp(wilderness, result.DungeonConnection.FirstStamp) ||
			!hasConnectionStamp(result.Dungeon, result.DungeonConnection.SecondStamp) {
			return EntryWorld{}, fmt.Errorf("restore entry world: dungeon connection identities do not match recipes")
		}
	}
	return result, nil
}

func hasConnectionStamp(zone *worldgen.Zone, id uint32) bool {
	for _, stamp := range zone.Stamps() {
		if stamp.ID == id {
			return true
		}
	}
	return false
}

// GenerateEntryWorld asks d2legacy for entry-area recipes and their authored connections.
// The headless runtime prevents authoritative policy from observing client state.
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

	dungeon, err := runtime.generateFrom(ctx, "d2legacy.mapgen.entry_world", "dungeon", float64(seed), float64(difficulty))
	if err != nil {
		return EntryWorld{}, fmt.Errorf("generate d2legacy entry dungeon: %w", err)
	}
	dungeonFacts, err := zoneFacts(dungeon, "dungeon")
	if err != nil {
		return EntryWorld{}, err
	}
	value, err := modruntime.Call(ctx, runtime.lua, "d2legacy.mapgen.entry_world", "dungeon_connection", wildernessFacts, dungeonFacts)
	if err != nil {
		return EntryWorld{}, fmt.Errorf("describe d2legacy dungeon connection: %w", err)
	}
	facts, ok := value.(map[string]any)
	if !ok {
		return EntryWorld{}, fmt.Errorf("describe d2legacy dungeon connection: result is not a table")
	}
	firstLevel, err := seamNumber(facts, "first_level")
	if err != nil {
		return EntryWorld{}, err
	}
	secondLevel, err := seamNumber(facts, "second_level")
	if err != nil {
		return EntryWorld{}, err
	}
	firstStamp, err := seamNumber(facts, "first_stamp")
	if err != nil {
		return EntryWorld{}, err
	}
	secondStamp, err := seamNumber(facts, "second_stamp")
	if err != nil {
		return EntryWorld{}, err
	}
	connection := &ConnectionSpec{FirstLevel: firstLevel, SecondLevel: secondLevel,
		FirstStamp: uint32(firstStamp), SecondStamp: uint32(secondStamp)}
	return EntryWorld{Town: town, Wilderness: wilderness, Seam: seam, Dungeon: dungeon, DungeonConnection: connection}, nil
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

	return SeamSpec{
		FirstLevel:      firstLevel,
		FirstDirection:  firstDirection,
		SecondLevel:     secondLevel,
		SecondDirection: secondDirection,
		SecondTileX:     secondX,
		SecondTileY:     secondY,
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
