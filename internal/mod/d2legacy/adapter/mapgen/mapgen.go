// Package mapgen boots the authoritative d2legacy map-generation policy and
// transports its immutable recipes back to the generic engine materializer.
//
// The important split is simple: Lua decides what Diablo II world to build.
// This Go adapter only owns VM lifetime and typed transport across the host
// boundary. It contains no generation formulas, level IDs, or layout rules.
package mapgen

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TownEntry asks d2legacy to choose a marker from copied decoded object facts,
// then applies only the generic collision-space nearest-open-point algorithm.
func (runtime *Runtime) TownEntry(ctx context.Context, world *gameworld.Map) (float64, float64, bool) {
	if runtime == nil || world == nil {
		return 0, 0, false
	}
	objects := make([]any, 0, len(world.Objects))
	for _, object := range world.Objects {
		objects = append(objects, map[string]any{"type": int(object.Type), "id": int(object.ID), "x": int(object.X), "y": int(object.Y)})
	}
	selected, err := modruntime.Call(nonNilContext(ctx), runtime.lua, "d2legacy.mapgen.town_spawn", "choose", objects)
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

// ResolveTownEntry owns the short-lived policy runtime needed after DS1 object
// decoding. It keeps the interactive client from naming campfire identities.
func ResolveTownEntry(ctx context.Context, source fs.FS, records recordsGateway, world *gameworld.Map) (float64, float64, bool) {
	runtime, err := NewRuntime(nonNilContext(ctx), source, records)
	if err != nil {
		return 0, 0, false
	}
	defer runtime.Close(context.Background())
	return runtime.TownEntry(ctx, world)
}

type recordsGateway interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

// Runtime owns the minimal headless Lua host needed by d2legacy map policy.
// Keeping this lifecycle wrapper here lets tests and tools exercise the same
// mod code as production without importing a client composition root.
type Runtime struct {
	lua *modruntime.Runtime
}

func NewRuntime(ctx context.Context, source fs.FS, records recordsGateway) (*Runtime, error) {
	ctx = nonNilContext(ctx)
	luaRuntime := modruntime.New()
	if err := luaRuntime.RegisterInstaller(modruntime.ContentRequire(source, "lua")); err != nil {
		return nil, err
	}
	for _, module := range []modruntime.Module{
		modruntime.RecordsModule(records),
		modruntime.DeterministicModule(),
		modruntime.WorldgenModule(),
	} {
		if err := luaRuntime.RegisterModule(module); err != nil {
			return nil, err
		}
	}
	if err := luaRuntime.Start(ctx); err != nil {
		return nil, err
	}
	return &Runtime{lua: luaRuntime}, nil
}

func (runtime *Runtime) Close(ctx context.Context) error {
	if runtime == nil || runtime.lua == nil {
		return nil
	}
	return runtime.lua.Stop(nonNilContext(ctx))
}

// Generate invokes one named d2legacy map strategy. Arguments are deliberately
// plain serialized values so this adapter cannot grow native Diablo policy.
func (runtime *Runtime) Generate(ctx context.Context, strategy string, arguments ...any) (*worldgen.Zone, error) {
	if runtime == nil || runtime.lua == nil {
		return nil, fmt.Errorf("d2legacy mapgen runtime is not started")
	}
	return modruntime.GenerateWorldRecipe(nonNilContext(ctx), runtime.lua, "d2legacy.mapgen."+strategy, "generate", arguments...)
}

func (runtime *Runtime) generateFrom(ctx context.Context, module, function string, arguments ...any) (*worldgen.Zone, error) {
	if runtime == nil || runtime.lua == nil {
		return nil, fmt.Errorf("d2legacy mapgen runtime is not started")
	}
	return modruntime.GenerateWorldRecipe(nonNilContext(ctx), runtime.lua, module, function, arguments...)
}

// GenerateEntryZones asks d2legacy for the first town and wilderness recipes.
// The short-lived runtime is intentionally headless: world policy must not
// depend on a renderer, window, audio device, or native client startup.
func GenerateEntryZones(ctx context.Context, source fs.FS, records recordsGateway, seed uint64) (*worldgen.Zone, *worldgen.Zone, error) {
	ctx = nonNilContext(ctx)
	runtime, err := NewRuntime(ctx, source, records)
	if err != nil {
		return nil, nil, err
	}
	defer runtime.Close(context.Background())

	town, err := runtime.generateFrom(ctx, "d2legacy.mapgen.entry_world", "town", float64(seed), float64(0))
	if err != nil {
		return nil, nil, fmt.Errorf("generate d2legacy entry town: %w", err)
	}
	encodedTown, err := town.MarshalJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("encode d2legacy entry town: %w", err)
	}
	var townFacts map[string]any
	if err := json.Unmarshal(encodedTown, &townFacts); err != nil {
		return nil, nil, fmt.Errorf("decode d2legacy entry town facts: %w", err)
	}
	moor, err := runtime.generateFrom(ctx, "d2legacy.mapgen.entry_world", "wilderness", float64(seed), float64(0), townFacts)
	if err != nil {
		return nil, nil, fmt.Errorf("generate d2legacy entry wilderness: %w", err)
	}
	return town, moor, nil
}

func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
