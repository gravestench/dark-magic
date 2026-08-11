package modruntime

import (
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
	lua "github.com/yuin/gopher-lua"
)

// MapgenModule exposes generated value snapshots, never mutable generator or
// catalog ownership. Lua labs can inspect recipes while gameplay sessions can
// use the same Go generator without a Lua VM.
func MapgenModule(catalog gameDataSnapshotter) Module {
	return Module{Name: "dm.mapgen/v1", Help: documentedModule("Generate deterministic renderer-independent zone recipes.", map[string]CommandHelp{
		"preset": commandHelp("dm.mapgen.preset(level_id, seed [, difficulty])", "Generate a typed preset zone and return its canonical value snapshot."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"preset": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				levelID := state.CheckInt(1)
				level, found := snapshot.LevelsByID[levelID]
				if !found {
					return pushLuaError(state, fmt.Errorf("mapgen: level %d is absent from Levels", levelID))
				}
				zone, err := mapgen.NewPresetGenerator(snapshot).Generate(mapgen.Request{
					Version: mapgen.ContractVersion, Seed: uint64(state.CheckNumber(2)),
					Act: uint8(level.Act + 1), LevelID: levelID, Difficulty: mapgen.Difficulty(state.OptInt(3, 0)),
				})
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(zoneToLua(state, zone))
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func zoneToLua(state *lua.LState, zone *mapgen.Zone) *lua.LTable {
	request, bounds := zone.Request(), zone.Bounds()
	result := state.NewTable()
	result.RawSetString("kind", lua.LString(zone.Kind()))
	result.RawSetString("seed", lua.LNumber(request.Seed))
	setLuaInteger(result, "act", int(request.Act))
	setLuaInteger(result, "level_id", request.LevelID)
	setLuaInteger(result, "difficulty", int(request.Difficulty))
	setLuaInteger(result, "width", bounds.Width)
	setLuaInteger(result, "height", bounds.Height)
	checksum, _ := zone.Checksum()
	result.RawSetString("checksum", lua.LString(checksum))
	stamps := state.NewTable()
	for _, stamp := range zone.Stamps() {
		item := state.NewTable()
		setLuaInteger(item, "id", int(stamp.ID))
		setLuaInteger(item, "preset_def", stamp.PresetDef)
		setLuaInteger(item, "variant", stamp.Variant)
		item.RawSetString("ds1", lua.LString(stamp.DS1Path))
		tiles := state.NewTable()
		for _, tile := range stamp.TilePaths {
			tiles.Append(lua.LString(tile))
		}
		item.RawSetString("dt1", tiles)
		stamps.Append(item)
	}
	result.RawSetString("stamps", stamps)
	trace := state.NewTable()
	for _, line := range zone.Trace() {
		trace.Append(lua.LString(line))
	}
	result.RawSetString("trace", trace)
	return result
}
