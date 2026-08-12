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

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

type recordsGateway interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

// GenerateEntryZones asks d2legacy for the first town and wilderness recipes.
// The short-lived runtime is intentionally headless: world policy must not
// depend on a renderer, window, audio device, or native client startup.
func GenerateEntryZones(ctx context.Context, source fs.FS, records recordsGateway, seed uint64) (*worldgen.Zone, *worldgen.Zone, error) {
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(source, "lua")); err != nil {
		return nil, nil, err
	}
	for _, module := range []modruntime.Module{
		modruntime.RecordsModule(records),
		modruntime.DeterministicModule(),
		modruntime.WorldgenModule(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			return nil, nil, err
		}
	}
	if err := runtime.Start(ctx); err != nil {
		return nil, nil, err
	}
	defer runtime.Stop(context.Background())

	town, err := modruntime.GenerateWorldRecipe(ctx, runtime, "d2legacy.mapgen.preset", "generate", float64(1), float64(seed), float64(0))
	if err != nil {
		return nil, nil, fmt.Errorf("generate d2legacy entry town: %w", err)
	}
	stamps := town.Stamps()
	if len(stamps) == 0 || !strings.HasPrefix(stamps[0].Role, "act1-town:exit-") {
		return nil, nil, fmt.Errorf("d2legacy entry town did not declare a cardinal exit")
	}
	exit := strings.TrimPrefix(stamps[0].Role, "act1-town:exit-")
	moor, err := modruntime.GenerateWorldRecipe(ctx, runtime, "d2legacy.mapgen.outdoor", "generate", float64(2), float64(seed), exit, float64(0))
	if err != nil {
		return nil, nil, fmt.Errorf("generate d2legacy entry wilderness: %w", err)
	}
	return town, moor, nil
}
