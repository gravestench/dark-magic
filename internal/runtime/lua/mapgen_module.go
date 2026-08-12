package modruntime

import (
	"fmt"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/mapgen"
	lua "github.com/yuin/gopher-lua"
)

type gameDataSnapshotter interface {
	Snapshot() (gamedata.Snapshot, error)
}

// MapgenModule exposes generated value snapshots, never mutable generator or
// catalog ownership. Lua labs can inspect recipes while gameplay sessions can
// use the same Go generator without a Lua VM.
func MapgenModule(catalog gameDataSnapshotter) Module {
	return Module{Name: "d2legacy.mapgen/v1", Help: documentedModule("Generate Diablo II zone recipes for the first-party mod.", map[string]CommandHelp{
		"preset":  commandHelp("d2legacy.mapgen.preset(level_id, seed [, difficulty])", "Generate a typed preset zone and return its canonical value snapshot."),
		"maze":    commandHelp("d2legacy.mapgen.maze(level_id, seed [, difficulty])", "Generate a typed maze zone and return rooms, links, recipes, and trace."),
		"outdoor": commandHelp("d2legacy.mapgen.outdoor(level_id, seed, town_exit [, difficulty])", "Generate Blood Moor joined to a north/east/south/west town exit."),
	}), Loader: func(state *lua.LState) int {
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"outdoor": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				levelID := state.CheckInt(1)
				zone, err := mapgen.NewActOneOutdoorGenerator(snapshot).GenerateFromTown(mapgen.Request{Version: mapgen.ContractVersion, Seed: uint64(state.CheckNumber(2)), Act: 1, LevelID: levelID, Difficulty: mapgen.Difficulty(state.OptInt(4, 0))}, mapgen.Stamp{Role: "act1-town:exit-" + state.CheckString(3)})
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(zoneToLua(state, zone))
				return 1
			},
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
			"maze": func(state *lua.LState) int {
				snapshot, err := catalog.Snapshot()
				if err != nil {
					return pushLuaError(state, err)
				}
				levelID := state.CheckInt(1)
				level, found := snapshot.LevelsByID[levelID]
				if !found {
					return pushLuaError(state, fmt.Errorf("mapgen: level %d is absent from Levels", levelID))
				}
				zone, err := mapgen.NewMazeGenerator(snapshot).Generate(mapgen.Request{
					Version: mapgen.ContractVersion, Seed: uint64(state.CheckNumber(2)), Act: uint8(level.Act + 1),
					LevelID: levelID, Difficulty: mapgen.Difficulty(state.OptInt(3, 0)),
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
		item.RawSetString("role", lua.LString(stamp.Role))
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
	rooms := state.NewTable()
	for _, room := range zone.Rooms() {
		item := state.NewTable()
		setLuaInteger(item, "id", int(room.ID))
		setLuaInteger(item, "x", room.X)
		setLuaInteger(item, "y", room.Y)
		setLuaInteger(item, "width", room.Width)
		setLuaInteger(item, "height", room.Height)
		setLuaInteger(item, "stamp_id", int(room.StampID))
		rooms.Append(item)
	}
	result.RawSetString("rooms", rooms)
	links := state.NewTable()
	for _, link := range zone.Links() {
		item := state.NewTable()
		setLuaInteger(item, "from", int(link.From))
		setLuaInteger(item, "to", int(link.To))
		links.Append(item)
	}
	result.RawSetString("links", links)
	warps := state.NewTable()
	for _, warp := range zone.Warps() {
		item := state.NewTable()
		setLuaInteger(item, "id", int(warp.ID))
		setLuaInteger(item, "x", warp.X)
		setLuaInteger(item, "y", warp.Y)
		setLuaInteger(item, "destination_level", warp.DestinationLevel)
		item.RawSetString("role", lua.LString(warp.Role))
		item.RawSetString("direction", lua.LString(warp.Direction))
		warps.Append(item)
	}
	result.RawSetString("warps", warps)
	structures := state.NewTable()
	for _, structure := range zone.Structures() {
		item := state.NewTable()
		setLuaInteger(item, "x", structure.X)
		setLuaInteger(item, "y", structure.Y)
		item.RawSetString("kind", lua.LString(structure.Kind))
		item.RawSetString("passable", lua.LBool(structure.Passable))
		structures.Append(item)
	}
	result.RawSetString("structures", structures)
	trace := state.NewTable()
	for _, line := range zone.Trace() {
		trace.Append(lua.LString(line))
	}
	result.RawSetString("trace", trace)
	return result
}
