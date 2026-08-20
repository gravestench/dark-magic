package mapgen

import (
	"context"
	"io/fs"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TownEntry asks d2legacy to choose a marker from copied decoded object facts, then applies only the generic
// collision-space nearest-open-point algorithm. Policy selection remains in Lua; Go only validates transported facts.
func (runtime *Runtime) TownEntry(ctx context.Context, world *gameworld.Map) (float64, float64, bool) {
	if runtime == nil || world == nil {
		return 0, 0, false
	}

	selected, err := modruntime.Call(
		nonNilContext(ctx),
		runtime.lua,
		"d2legacy.mapgen.town_spawn",
		"choose",
		townEntryObjects(world),
	)
	if err != nil || selected == nil {
		return 0, 0, false
	}

	facts, ok := selected.(map[string]any)
	if !ok {
		return 0, 0, false
	}

	x, xOK := facts["x"].(float64)
	y, yOK := facts["y"].(float64)

	inset, insetOK := facts["search_inset"].(float64)
	if !xOK || !yOK || !insetOK || inset < 0 {
		return 0, 0, false
	}

	return world.OpenPointNear(x, y, int(inset))
}

// townEntryObjects copies only the decoded facts needed by policy. This prevents Lua from retaining or mutating the
// authoritative materialized object slice.
func townEntryObjects(world *gameworld.Map) []any {
	objects := make([]any, 0, len(world.Objects))
	for _, object := range world.Objects {
		objects = append(objects, map[string]any{
			"type": int(object.Type),
			"id":   int(object.ID),
			"x":    int(object.X),
			"y":    int(object.Y),
		})
	}

	return objects
}

// ResolveTownEntry owns the short-lived policy runtime needed after DS1 object decoding. It keeps composition roots
// from naming campfire identities and treats policy startup or selection failure as an unavailable trusted spawn.
func ResolveTownEntry(
	ctx context.Context,
	source fs.FS,
	records recordsGateway,
	world *gameworld.Map,
) (float64, float64, bool) {
	runtime, err := NewRuntime(nonNilContext(ctx), source, records)
	if err != nil {
		return 0, 0, false
	}

	// Spawn resolution has no error return for cleanup, so shutdown remains best effort after the policy result is known.
	defer func() {
		_ = runtime.Close(context.Background())
	}()

	return runtime.TownEntry(ctx, world)
}
