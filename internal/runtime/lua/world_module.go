package modruntime

import (
	"context"
	"fmt"
	"io/fs"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	lua "github.com/yuin/gopher-lua"
)

// SetWorldMap applies world map through the capability boundary so validation completes before shared state
// changes.
func SetWorldMap(
	ctx context.Context,
	runtime *Runtime,
	moduleName, setterName string,
	world *gameworld.Map,
) error {
	return setWorldMap(ctx, runtime, moduleName, setterName, world)
}

// SetWorldMapForLevel installs one immutable collision map into a level-indexed
// authoritative registry. Multiple players may occupy different levels at the
// same tick, so a network authority cannot rely on one process-global current
// collision map selected by a presentation scene.
func SetWorldMapForLevel(
	ctx context.Context,
	runtime *Runtime,
	moduleName, setterName string,
	levelID int,
	world *gameworld.Map,
) error {
	if levelID <= 0 {
		return fmt.Errorf("world module: positive level ID is required")
	}

	return setWorldMap(ctx, runtime, moduleName, setterName, world, lua.LNumber(levelID))
}

// setWorldMap applies world map through the capability boundary so validation completes before shared state
// changes.
func setWorldMap(
	ctx context.Context,
	runtime *Runtime,
	moduleName, setterName string,
	world *gameworld.Map,
	leading ...lua.LValue,
) error {
	if runtime == nil || world == nil || moduleName == "" || setterName == "" {
		return fmt.Errorf("world module: runtime, module, setter, and map are required")
	}

	return runtime.Run(ctx, func(state *lua.LState) error {
		// Headless/listen authorities may receive a native world map without
		// installing engine.world/v1. Register the userdata contract at this
		// boundary so methods such as blocked_position always exist.
		registerWorldMapType(state)

		if err := state.CallByParam(
			lua.P{Fn: state.GetGlobal("require"), NRet: 1, Protect: true},
			lua.LString(moduleName),
		); err != nil {
			return err
		}

		value := state.Get(-1)
		state.Pop(1)

		table, ok := value.(*lua.LTable)
		if !ok {
			return fmt.Errorf("world module: %q is invalid", moduleName)
		}

		setter := table.RawGetString(setterName)

		pushWorldMap(state, world)
		collision := state.Get(-1)
		state.Pop(1)

		arguments := append(append([]lua.LValue(nil), leading...), collision)

		return state.CallByParam(lua.P{Fn: setter, NRet: 0, Protect: true}, arguments...)
	})
}

const worldMapType = "engine.world.map/v1"

// CurrentWorld is the immutable authoritative map and the asset recipe used by
// presentation to draw that same map. The slice is copied into Lua.
type CurrentWorld struct {
	Map     *gameworld.Map
	DS1     string
	DT1     []string
	LevelID int
}

type CurrentWorldProvider func() CurrentWorld

// WorldModule decodes immutable world facts from mounted DS1 and DT1 assets.
// Rendering remains a separate capability so headless servers can use the same
// authored collision and object data as clients.
func WorldModule(source fs.FS, resolvers ...gameworld.ObjectResolver) Module {
	return worldModule(source, nil, resolvers...)
}

// SessionWorldModule adds read-only access to the current session-owned map.
// Asset loading remains available for labs, but gameplay scenes use current().
func SessionWorldModule(
	source fs.FS,
	current CurrentWorldProvider,
	resolvers ...gameworld.ObjectResolver,
) Module {
	return worldModule(source, current, resolvers...)
}

// worldModule defines the world module Lua boundary in one place so scripts receive a stable command and error
// contract.
func worldModule(
	source fs.FS,
	current CurrentWorldProvider,
	resolvers ...gameworld.ObjectResolver,
) Module {
	return Module{
		Name: "engine.world/v1",
		Help: documentedModule(
			"Decode immutable gameplay facts from authored world assets.",
			map[string]CommandHelp{
				"load": commandHelp(
					"engine.world.load(ds1_path, dt1_paths)",
					"Decode a DS1 stamp and its DT1 tilesets into an immutable map handle.",
				),
				"current": commandHelp(
					"engine.world.current()",
					"Return the current session-owned map and its copied presentation recipe.",
				),
				"current_level": commandHelp(
					"engine.world.current_level()",
					"Return the active authoritative level ID without allocating a map handle.",
				),
			},
			map[string]TypeHelp{
				worldMapType: {
					Summary: "An immutable decoded DS1 world map.",
					Methods: map[string]CommandHelp{
						"dimensions": commandHelp(
							"map:dimensions()",
							"Return tile and subtile dimensions plus act and object count.",
						),
						"canvas": commandHelp(
							"map:canvas()",
							"Return isometric world-pixel canvas dimensions.",
						),
						"entity_depth": commandHelp(
							"map:entity_depth(x, y)",
							"Return shared presentation depth for a world entity baseline.",
						),
						"flags": commandHelp(
							"map:flags(x, y)",
							"Return collision flags at zero-based subtile coordinates, or nil outside the map.",
						),
						"blocked": commandHelp(
							"map:blocked(x, y)",
							"Report whether a player cannot walk through a zero-based subtile coordinate.",
						),
						"blocked_position": commandHelp(
							"map:blocked_position(x, y)",
							"Sample collision at a continuous subtile-center position.",
						),
						"integrate_velocity": commandHelp(
							"map:integrate_velocity(x, y, velocity_x, velocity_y, radius, width, height, elapsed)",
							"Apply one shared bounded collision-aware velocity step.",
						),
						"objects": commandHelp(
							"map:objects()",
							"Return a copy of the DS1 authored object records.",
						),
						"subtile_to_pixel": commandHelp(
							"map:subtile_to_pixel(x, y)",
							"Project continuous DS1 subtile coordinates into rendered map pixels.",
						),
						"pixel_to_subtile": commandHelp(
							"map:pixel_to_subtile(x, y)",
							"Convert rendered map pixels into continuous DS1 subtile coordinates.",
						),
						"subtile_to_screen": commandHelp(
							"map:subtile_to_screen(x, y, camera_x, camera_y, anchor_x, anchor_y)",
							"Project a world subtile through a camera to screen space.",
						),
						"screen_to_subtile": commandHelp(
							"map:screen_to_subtile(x, y, camera_x, camera_y, anchor_x, anchor_y)",
							"Convert a screen pointer position to a world subtile.",
						),
						"selectable_at": commandHelp(
							"map:selectable_at(x, y)",
							"Return the best resolved authored object under a world-subtile point.",
						),
						"line_clear": commandHelp(
							"map:line_clear(from_x, from_y, to_x, to_y)",
							"Test authoritative DT1 line-of-sight collision.",
						),
						"barrier_clear": commandHelp(
							"map:barrier_clear(from_x, from_y, to_x, to_y)",
							"Test authoritative DT1 flying/melee-barrier collision.",
						),
						"find_path": commandHelp(
							"map:find_path(from_x, from_y, to_x, to_y [, radius, stop_radius])",
							"Find a deterministic collision-aware subtile route, or return nil and an explanation.",
						),
					},
				},
			},
		),
		Loader: func(state *lua.LState) int {
			registerWorldMapType(state)
			module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
				"current_level": func(state *lua.LState) int {
					if current == nil {
						state.Push(lua.LNil)
						return 1
					}

					value := current()
					if value.Map == nil {
						state.Push(lua.LNil)
					} else {
						state.Push(lua.LNumber(value.LevelID))
					}

					return 1
				},
				"current": func(state *lua.LState) int {
					if current == nil {
						state.Push(lua.LNil)
						return 1
					}

					value := current()
					if value.Map == nil {
						state.Push(lua.LNil)
						return 1
					}

					pushWorldMap(state, value.Map)
					recipe := state.NewTable()
					recipe.RawSetString("ds1", lua.LString(value.DS1))
					setLuaInteger(recipe, "level_id", value.LevelID)

					tiles := state.NewTable()
					for _, path := range value.DT1 {
						tiles.Append(lua.LString(path))
					}

					recipe.RawSetString("dt1", tiles)
					state.Push(recipe)

					return 2
				},
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

					pushWorldMap(state, decoded)

					return 1
				},
			})
			module.RawSetString("api", lua.LNumber(1))
			state.Push(module)

			return 1
		},
	}
}

// pushWorldMap publishes world map in one canonical Lua representation, keeping scripts independent of Go
// storage details.
func pushWorldMap(state *lua.LState, world *gameworld.Map) {
	userData := state.NewUserData()
	userData.Value = world
	state.SetMetatable(userData, state.GetTypeMetatable(worldMapType))
	state.Push(userData)
}

// registerWorldMapType registers register world map type before use so Lua cannot observe a partially configured
// capability.
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
			state.Push(
				lua.LNumber(
					(world.WidthTiles+world.HeightTiles)*gameworld.TilePixelWidth/2 + gameworld.PreviewMargin*2,
				),
			)
			state.Push(
				lua.LNumber(
					(world.WidthTiles+world.HeightTiles)*gameworld.TilePixelHeight/2 + gameworld.PreviewMargin*2,
				),
			)

			return 2
		},
		"entity_depth": func(state *lua.LState) int {
			state.Push(
				lua.LNumber(
					gameworld.EntityDepth(
						float64(state.CheckNumber(2)),
						float64(state.CheckNumber(3)),
					),
				),
			)

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
			flags, ok := checkWorldMap(
				state,
				1,
			).FlagsAtPosition(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			state.Push(lua.LBool(ok && flags.Blocked()))

			return 1
		},
		"integrate_velocity": func(state *lua.LState) int {
			position := gameworld.Point{
				X: float64(state.CheckNumber(2)),
				Y: float64(state.CheckNumber(3)),
			}
			velocity := gameworld.Point{
				X: float64(state.CheckNumber(4)),
				Y: float64(state.CheckNumber(5)),
			}
			radius := float64(state.CheckNumber(6))
			bounds := gameworld.Point{
				X: float64(state.CheckNumber(7)),
				Y: float64(state.CheckNumber(8)),
			}
			elapsed := float64(state.CheckNumber(9))
			result := gameworld.IntegrateVelocity(
				checkWorldMap(state, 1),
				position,
				velocity,
				bounds,
				radius,
				elapsed,
			)
			state.Push(lua.LNumber(result.X))
			state.Push(lua.LNumber(result.Y))

			return 2
		},
		"subtile_to_pixel": func(state *lua.LState) int {
			x, y := checkWorldMap(
				state,
				1,
			).SubtileToPixel(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			state.Push(lua.LNumber(x))
			state.Push(lua.LNumber(y))

			return 2
		},
		"pixel_to_subtile": func(state *lua.LState) int {
			x, y := checkWorldMap(
				state,
				1,
			).PixelToSubtile(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			state.Push(lua.LNumber(x))
			state.Push(lua.LNumber(y))

			return 2
		},
		"subtile_to_screen": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			point := gameworld.Point{
				X: float64(state.CheckNumber(2)),
				Y: float64(state.CheckNumber(3)),
			}
			camera := gameworld.Point{
				X: float64(state.CheckNumber(4)),
				Y: float64(state.CheckNumber(5)),
			}
			anchor := gameworld.Point{
				X: float64(state.CheckNumber(6)),
				Y: float64(state.CheckNumber(7)),
			}
			result := world.Coordinates().SubtileToScreen(point, camera, anchor)
			state.Push(lua.LNumber(result.X))
			state.Push(lua.LNumber(result.Y))

			return 2
		},
		"screen_to_subtile": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			point := gameworld.Point{
				X: float64(state.CheckNumber(2)),
				Y: float64(state.CheckNumber(3)),
			}
			camera := gameworld.Point{
				X: float64(state.CheckNumber(4)),
				Y: float64(state.CheckNumber(5)),
			}
			anchor := gameworld.Point{
				X: float64(state.CheckNumber(6)),
				Y: float64(state.CheckNumber(7)),
			}
			result := world.Coordinates().ScreenToSubtile(point, camera, anchor)
			state.Push(lua.LNumber(result.X))
			state.Push(lua.LNumber(result.Y))

			return 2
		},
		"selectable_at": func(state *lua.LState) int {
			selected, found := checkWorldMap(
				state,
				1,
			).SelectableAt(float64(state.CheckNumber(2)), float64(state.CheckNumber(3)))
			if !found {
				state.Push(lua.LNil)
				return 1
			}

			result := state.NewTable()
			result.RawSetString("id", lua.LString(selected.ID))
			result.RawSetString("kind", lua.LString(selected.Kind))
			result.RawSetString("label", lua.LString(selected.Label))
			result.RawSetString("x", lua.LNumber(selected.X))
			result.RawSetString("y", lua.LNumber(selected.Y))
			result.RawSetString("radius", lua.LNumber(selected.Radius))
			state.Push(result)

			return 1
		},
		"line_clear": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			fromX, fromY := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			toX, toY := float64(state.CheckNumber(4)), float64(state.CheckNumber(5))
			clear := world.LineClear(fromX, fromY, toX, toY)
			state.Push(lua.LBool(clear))

			return 1
		},
		"barrier_clear": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			fromX, fromY := float64(state.CheckNumber(2)), float64(state.CheckNumber(3))
			toX, toY := float64(state.CheckNumber(4)), float64(state.CheckNumber(5))
			clear := world.BarrierClear(fromX, fromY, toX, toY)
			state.Push(lua.LBool(clear))

			return 1
		},
		"find_path": func(state *lua.LState) int {
			world := checkWorldMap(state, 1)
			request := gameworld.PathRequest{
				Start: gameworld.Point{
					X: float64(state.CheckNumber(2)),
					Y: float64(state.CheckNumber(3)),
				},
				Goal: gameworld.Point{
					X: float64(state.CheckNumber(4)),
					Y: float64(state.CheckNumber(5)),
				},
				Radius:     float64(state.OptNumber(6, 0)),
				StopRadius: float64(state.OptNumber(7, 0)),
			}

			path, err := world.FindPath(request)
			if err != nil {
				state.Push(lua.LNil)
				state.Push(lua.LString(err.Error()))

				return 2
			}

			result := state.NewTable()
			for _, point := range path {
				waypoint := state.NewTable()
				waypoint.RawSetString("x", lua.LNumber(point.X))
				waypoint.RawSetString("y", lua.LNumber(point.Y))
				result.Append(waypoint)
			}

			state.Push(result)
			state.Push(lua.LNil)

			return 2
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

// checkWorldMap validates check world map at the Lua boundary so invalid values fail before shared state can
// change.
func checkWorldMap(state *lua.LState, index int) *gameworld.Map {
	userData := state.CheckUserData(index)

	world, ok := userData.Value.(*gameworld.Map)
	if !ok {
		state.ArgError(index, "engine.world/v1 map expected")
		return nil
	}

	return world
}

// setLuaInteger applies lua integer through the capability boundary so validation completes before shared state
// changes.
func setLuaInteger(table *lua.LTable, name string, value int) {
	table.RawSetString(name, lua.LNumber(value))
}

// worldFlagsTable owns the world flags table step at this boundary, keeping its side effects and failure point
// explicit to callers.
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
