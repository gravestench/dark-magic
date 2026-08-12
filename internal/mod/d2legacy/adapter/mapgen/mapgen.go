// Package mapgen boots the authoritative d2legacy map-generation policy and
// transports its immutable recipes back to the generic engine materializer.
//
// The important split is simple: Lua decides what Diablo II world to build.
// This Go adapter only owns VM lifetime and typed transport across the host
// boundary. It contains no generation formulas, level IDs, or layout rules.
package mapgen

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// ActOneTownEntry resolves the d2legacy Rogue Encampment spawn near its
// authored bonfire object. Object type/ID and search inset are mod policy.
func ActOneTownEntry(world *gameworld.Map) (float64, float64, bool) {
	if world == nil {
		return 0, 0, false
	}
	for _, object := range world.Objects {
		if object.Type == gameworld.ObjectTypeStatic && object.ID == 2 {
			return world.OpenPointNear(float64(object.X), float64(object.Y), 4)
		}
	}
	return 0, 0, false
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

	town, err := runtime.Generate(ctx, "preset", float64(1), float64(seed), float64(0))
	if err != nil {
		return nil, nil, fmt.Errorf("generate d2legacy entry town: %w", err)
	}
	stamps := town.Stamps()
	if len(stamps) == 0 || !strings.HasPrefix(stamps[0].Role, "act1-town:exit-") {
		return nil, nil, fmt.Errorf("d2legacy entry town did not declare a cardinal exit")
	}
	exit := strings.TrimPrefix(stamps[0].Role, "act1-town:exit-")
	moor, err := runtime.Generate(ctx, "outdoor", float64(2), float64(seed), exit, float64(0))
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
