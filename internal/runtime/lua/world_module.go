package modruntime

import (
	"io/fs"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	lua "github.com/yuin/gopher-lua"
)

const worldMapType = "dm.world.map/v1"

// WorldModule decodes immutable world facts from mounted DS1 and DT1 assets.
// Rendering remains a separate capability so headless servers can use the same
// authored collision and object data as clients.
func WorldModule(source fs.FS, resolvers ...gameworld.ObjectResolver) Module {
	return Module{Name: "dm.world/v1", Help: documentedModule("Decode immutable gameplay facts from authored world assets.", map[string]CommandHelp{
		"load": commandHelp("dm.world.load(ds1_path, dt1_paths)", "Decode a DS1 stamp and its DT1 tilesets into an immutable map handle."),
	}, map[string]TypeHelp{worldMapType: {Summary: "An immutable decoded DS1 world map.", Methods: map[string]CommandHelp{
		"dimensions":        commandHelp("map:dimensions()", "Return tile and subtile dimensions plus act and object count."),
		"canvas":            commandHelp("map:canvas()", "Return isometric world-pixel canvas dimensions."),
		"entity_depth":      commandHelp("map:entity_depth(x, y)", "Return shared presentation depth for a world entity baseline."),
		"flags":             commandHelp("map:flags(x, y)", "Return collision flags at zero-based subtile coordinates, or nil outside the map."),
		"blocked":           commandHelp("map:blocked(x, y)", "Report whether a player cannot walk through a zero-based subtile coordinate."),
		"blocked_position":  commandHelp("map:blocked_position(x, y)", "Sample collision at a continuous subtile-center position."),
		"objects":           commandHelp("map:objects()", "Return a copy of the DS1 authored object records."),
		"subtile_to_pixel":  commandHelp("map:subtile_to_pixel(x, y)", "Project continuous DS1 subtile coordinates into rendered map pixels."),
		"pixel_to_subtile":  commandHelp("map:pixel_to_subtile(x, y)", "Convert rendered map pixels into continuous DS1 subtile coordinates."),
		"subtile_to_screen": commandHelp("map:subtile_to_screen(x, y, camera_x, camera_y, anchor_x, anchor_y)", "Project a world subtile through a camera to screen space."),
		"screen_to_subtile": commandHelp("map:screen_to_subtile(x, y, camera_x, camera_y, anchor_x, anchor_y)", "Convert a screen pointer position to a world subtile."),
		"selectable_at":     commandHelp("map:selectable_at(x, y)", "Return the best resolved authored object under a world-subtile point."),
		"line_clear":        commandHelp("map:line_clear(from_x, from_y, to_x, to_y)", "Test authoritative DT1 line-of-sight collision."),
	}}}), Loader: func(state *lua.LState) int {
		registerWorldMapType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"load": func(state *lua.LState) int {
				stampPath := state.CheckString(1)
				paths := state.CheckTable(2)
				tilePaths := make([]string, 0, paths.Len())
				for index := 1; index <= paths.Len(); index++ {
					value, ok := paths.RawGetInt(index).(lua.LString)
					if !ok {
						state.ArgError(2, "DT1 paths must be a sequence of strings")
						return 0
					}
					tilePaths = append(tilePaths, string(value))
				}
				decoded, err := gameworld.Load(source, stampPath, tilePaths, resolvers...)
				if err != nil {
					return pushLuaError(state, err)
				}
				userData := state.NewUserData()
				userData.Value = decoded
				state.SetMetatable(userData, state.GetTypeMetatable(worldMapType))
				state.Push(userData)
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

func registerWorldMapType(state *lua.LState) {
	meta := state.NewTypeMetatable(worldMapType)
	state.SetField(meta, "__index", state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
		"dimensions": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			result := state.NewTable()
			setLuaInteger(result, "width_tiles", world.WidthTiles)
			setLuaInteger(result, "height_tiles", world.HeightTiles)
			setLuaInteger(result, "width_subtiles", world.WidthSubtiles)
			setLuaInteger(result, "height_subtiles", world.HeightSubtiles)
			setLuaInteger(result, "act", world.Act)
			setLuaInteger(result, "object_count", len(world.Objects))
			state.Push(result)
			return 1
		},
		"canvas": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			state.Push(lua.LNumber((world.WidthTiles+world.HeightTiles)*gameworld.TilePixelWidth/2 + gameworld.PreviewMargin*2))
			state.Push(lua.LNumber((world.WidthTiles+world.HeightTiles)*gameworld.TilePixelHeight/2 + gameworld.PreviewMargin*2))
			return 2
		},
		"entity_depth": func(state *lua.LState) int {
			state.Push(lua.LNumber(gameworld.EntityDepth(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))))
			return 1
		},
		"flags": func(state *lua.LState) int {
			flags, ok := checkWorldMap(state, 1).FlagsAt(state.CheckInt(2), state.CheckInt(3))
			if !ok {
				state.Push(lua.LNil)
				return 1
			}
			state.Push(worldFlagsTable(state, flags))
			return 1
		},
		"blocked": func(state *lua.LState) int {
			flags, ok := checkWorldMap(state, 1).FlagsAt(state.CheckInt(2), state.CheckInt(3))
			state.Push(lua.LBool(ok && flags.Blocked()))
			return 1
		},
		"blocked_position": func(state *lua.LState) int {
			flags, ok := checkWorldMap(state, 1).FlagsAtPosition(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			state.Push(lua.LBool(ok && flags.Blocked()))
			return 1
		},
		"subtile_to_pixel": func(state *lua.LState) int {
			x, y := checkWorldMap(state, 1).SubtileToPixel(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			state.Push(lua.LNumber(x))
			state.Push(lua.LNumber(y))
			return 2
		},
		"pixel_to_subtile": func(state *lua.LState) int {
			x, y := checkWorldMap(state, 1).PixelToSubtile(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			state.Push(lua.LNumber(x))
			state.Push(lua.LNumber(y))
			return 2
		},
		"subtile_to_screen": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			point := gameworld.Point{X: float64(state.CheckNumber(2)), Y: float64(state.CheckNumber(3))}
			camera := gameworld.Point{X: float64(state.CheckNumber(4)), Y: float64(state.CheckNumber(5))}
			anchor := gameworld.Point{X: float64(state.CheckNumber(6)), Y: float64(state.CheckNumber(7))}
			result := world.Coordinates().SubtileToScreen(point, camera, anchor)
			state.Push(lua.LNumber(result.X))
			state.Push(lua.LNumber(result.Y))
			return 2
		},
		"screen_to_subtile": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			point := gameworld.Point{X: float64(state.CheckNumber(2)), Y: float64(state.CheckNumber(3))}
			camera := gameworld.Point{X: float64(state.CheckNumber(4)), Y: float64(state.CheckNumber(5))}
			anchor := gameworld.Point{X: float64(state.CheckNumber(6)), Y: float64(state.CheckNumber(7))}
			result := world.Coordinates().ScreenToSubtile(point, camera, anchor)
			state.Push(lua.LNumber(result.X))
			state.Push(lua.LNumber(result.Y))
			return 2
		},
		"selectable_at": func(state *lua.LState) int {
			selected, found := checkWorldMap(state, 1).SelectableAt(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			if !found {
				state.Push(lua.LNil)
				return 1
			}
			result := state.NewTable()
			result.RawSetString("id", lua.LString(selected.ID))
			result.RawSetString("kind", lua.LString(selected.Kind))
			result.RawSetString("x", lua.LNumber(selected.X))
			result.RawSetString("y", lua.LNumber(selected.Y))
			result.RawSetString("radius", lua.LNumber(selected.Radius))
			state.Push(result)
			return 1
		},
		"line_clear": func(state *lua.LState) int {
			clear := checkWorldMap(state, 1).LineClear(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)), float64(state.CheckNumber(4)), float64(state.CheckNumber(5)))
			state.Push(lua.LBool(clear))
			return 1
		},
		"objects": func(state *lua.LState) int {
			result := state.NewTable()
			for _, object := range checkWorldMap(state, 1).Objects {
				item := state.NewTable()
				setLuaInteger(item, "type", int(object.Type))
				setLuaInteger(item, "id", int(object.ID))
				setLuaInteger(item, "x", int(object.X))
				setLuaInteger(item, "y", int(object.Y))
				setLuaInteger(item, "flags", int(object.Flags))
				if object.Resolved {
					if object.Type == gameworld.ObjectTypeStatic {
						setLuaInteger(item, "object_id", object.ObjectID)
						item.RawSetString("description", lua.LString(object.Description))
					} else {
						item.RawSetString("class", lua.LString(object.Class))
					}
				}
				result.Append(item)
			}
			state.Push(result)
			return 1
		},
	}))
}

func checkWorldMap(state *lua.LState, index int) *gameworld.Map {
	userData := state.CheckUserData(index)
	world, ok := userData.Value.(*gameworld.Map)
	if !ok {
		state.ArgError(index, "dm.world/v1 map expected")
		return nil
	}
	return world
}

func setLuaInteger(table *lua.LTable, name string, value int) {
	table.RawSetString(name, lua.LNumber(value))
}

func worldFlagsTable(state *lua.LState, flags gameworld.Flags) *lua.LTable {
	result := state.NewTable()
	result.RawSetString("block_walk", lua.LBool(flags.BlockWalk))
	result.RawSetString("block_los", lua.LBool(flags.BlockLOS))
	result.RawSetString("block_jump", lua.LBool(flags.BlockJump))
	result.RawSetString("block_player_walk", lua.LBool(flags.BlockPlayerWalk))
	result.RawSetString("block_light", lua.LBool(flags.BlockLight))
	result.RawSetString("blocked", lua.LBool(flags.Blocked()))
	return result
}
