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

	"github.com/gravestench/dark-magic/internal/game/worldgen"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// recordsGateway is the narrow host boundary exposed to Lua's records module. Map policy can read recovered facts but
// remains independent of the concrete record cache and its composition root.
type recordsGateway interface {
	Load(string) ([]map[string]string, error)
	Invalidate(string)
	Loaded(string) bool
}

// Runtime owns the minimal headless Lua host needed by d2legacy map policy. Keeping this lifecycle wrapper here lets
// tests and tools exercise the production mod code without importing a client composition root.
type Runtime struct {
	lua *modruntime.Runtime
}

// NewRuntime installs only content loading, records, determinism, and generic world-generation modules. Excluding
// rendering and client services prevents authoritative map recipes from depending on local presentation state.
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

// Close stops the owned Lua host and is safe for nil or partially initialized runtimes, which simplifies deferred
// cleanup on failed composition paths.
func (runtime *Runtime) Close(ctx context.Context) error {
	if runtime == nil || runtime.lua == nil {
		return nil
	}

	return runtime.lua.Stop(nonNilContext(ctx))
}

// Generate invokes one named d2legacy map strategy. Arguments are deliberately plain serialized values so this adapter
// cannot grow native Diablo policy.
func (runtime *Runtime) Generate(
	ctx context.Context,
	strategy string,
	arguments ...any,
) (*worldgen.Zone, error) {
	return runtime.generateFrom(ctx, "d2legacy.mapgen."+strategy, "generate", arguments...)
}

// generateFrom transports one call into a fully qualified policy module. Keeping runtime validation here gives public
// and entry-world generation the same failure contract.
func (runtime *Runtime) generateFrom(
	ctx context.Context,
	module string,
	function string,
	arguments ...any,
) (*worldgen.Zone, error) {
	if runtime == nil || runtime.lua == nil {
		return nil, fmt.Errorf("d2legacy mapgen runtime is not started")
	}

	return modruntime.GenerateWorldRecipe(
		nonNilContext(ctx),
		runtime.lua,
		module,
		function,
		arguments...,
	)
}

// nonNilContext gives public entry points consistent nil-context behavior while still preserving cancellation and
// deadlines whenever a caller supplies a real context.
func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
