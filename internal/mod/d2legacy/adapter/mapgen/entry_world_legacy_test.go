package mapgen

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// generateLegacyEntryWorldV3 preserves the pre-v4 composition only as a test
// harness. The frozen v3 fixture was generated from entryWorldV3Records, never
// from protected assets. Production only generates the current Lua composition.
func generateLegacyEntryWorldV3(
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
