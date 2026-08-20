package modruntime

import (
	"fmt"
	"io/fs"

	assetinspect "github.com/gravestench/dark-magic/internal/assets/inspect"
	"github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mapeditor"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	lua "github.com/yuin/gopher-lua"
)

// MapEditorModule exposes one scene-owned editable DS1 document. Its API is
// intentionally narrow: Lua owns the interaction, while source integrity,
// history, preview materialization, and confined saving remain in Go.
func MapEditorModule(source fs.FS, storage *mapeditor.Storage, resolvers ...world.ObjectResolver) Module {
	session := &mapEditorSession{source: source, storage: storage, resolvers: resolvers}
	description := "Edit one lossless DS1 source document and preview it through Dark Magic's world pipeline."
	help := documentedModule(description, map[string]CommandHelp{
		"open": commandHelp(
			"engine.map_editor.open(path)",
			"Open a mounted DS1 and resolve its declared DT1 dependencies.",
		),
		"new": commandHelp("engine.map_editor.new(options)", "Create a modern empty v18 DS1 document."),
		"summary": commandHelp(
			"engine.map_editor.summary()",
			"Return dimensions, layer counts, dependency count, revision, and dirty state.",
		),
		"tiles": commandHelp(
			"engine.map_editor.tiles(kind)",
			"List logical DT1 identities available to a floor, wall, or shadow tile picker.",
		),
		"tilesets": commandHelp(
			"engine.map_editor.tilesets(kind)",
			"List declared DT1 files and their compatible physical tile counts.",
		),
		"tileset_tiles": commandHelp(
			"engine.map_editor.tileset_tiles(path, kind)",
			"List concrete DT1 records in one declared tileset.",
		),
		"cell": commandHelp(
			"engine.map_editor.cell(x, y)",
			"Inspect copied raw source records at a zero-based DS1 tile coordinate.",
		),
		"begin_stroke": commandHelp(
			"engine.map_editor.begin_stroke(kind, layer, brush)",
			"Begin one coalesced paint or erase gesture.",
		),
		"paint": commandHelp(
			"engine.map_editor.paint(x, y)",
			"Apply the active brush at a zero-based DS1 tile coordinate.",
		),
		"end_stroke":    commandHelp("engine.map_editor.end_stroke()", "Commit the current paint gesture as one undo step."),
		"cancel_stroke": commandHelp("engine.map_editor.cancel_stroke()", "Discard the active paint gesture."),
		"undo":          commandHelp("engine.map_editor.undo()", "Undo the previous completed gesture."),
		"redo":          commandHelp("engine.map_editor.redo()", "Redo the next previously undone gesture."),
		"preview": commandHelp(
			"engine.map_editor.preview()",
			"Materialize the unsaved document as an immutable world-map handle.",
		),
		"save": commandHelp(
			"engine.map_editor.save(relative_path)",
			"Atomically save below the configured map-editor output directory.",
		),
	})

	return Module{Name: "engine.map_editor/v1", Help: help, Loader: func(state *lua.LState) int {
		registerWorldMapType(state)
		module := state.SetFuncs(state.NewTable(), map[string]lua.LGFunction{
			"open": func(state *lua.LState) int {
				if err := session.open(state.CheckString(1)); err != nil {
					return pushLuaError(state, err)
				}
				state.Push(summaryTable(state, session.document.Summary()))
				return 1
			},
			"new": func(state *lua.LState) int {
				config, paths, err := newDocumentConfig(state.CheckTable(1))
				if err == nil {
					err = session.new(config, paths)
				}
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(summaryTable(state, session.document.Summary()))
				return 1
			},
			"summary": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				state.Push(summaryTable(state, session.document.Summary()))
				return 1
			},
			"tiles": func(state *lua.LState) int {
				if session.document == nil || session.catalog == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				kind, err := mapeditor.ParseLayerKind(state.CheckString(1))
				if err != nil {
					return pushLuaError(state, err)
				}
				result := state.NewTable()
				for _, identity := range session.catalog.Identities() {
					if !pickerIncludes(kind, identity) {
						continue
					}
					entry := state.NewTable()
					setLuaInteger(entry, "orientation", int(identity.Orientation))
					setLuaInteger(entry, "style", int(identity.MainIndex))
					setLuaInteger(entry, "sequence", int(identity.SubIndex))
					label := fmt.Sprintf(
						"%02d / %02d / %02d",
						identity.Orientation,
						identity.MainIndex,
						identity.SubIndex,
					)
					entry.RawSetString("label", lua.LString(label))
					candidates := session.catalog.Candidates(identity)
					setLuaInteger(entry, "variants", len(candidates))
					if len(candidates) > 0 {
						entry.RawSetString("source_path", lua.LString(candidates[0].Path))
						setLuaInteger(entry, "source_index", candidates[0].Index)
					}
					result.Append(entry)
				}
				state.Push(result)
				return 1
			},
			"tilesets": func(state *lua.LState) int {
				if session.document == nil || session.catalog == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				kind, err := mapeditor.ParseLayerKind(state.CheckString(1))
				if err != nil {
					return pushLuaError(state, err)
				}
				result := state.NewTable()
				for _, path := range session.tilePaths {
					count := 0
					for _, reference := range session.catalog.References(path) {
						if pickerIncludes(kind, reference.Identity) {
							count++
						}
					}
					if count == 0 {
						continue
					}
					entry := state.NewTable()
					entry.RawSetString("path", lua.LString(path))
					entry.RawSetString("label", lua.LString(fileName(path)))
					setLuaInteger(entry, "count", count)
					result.Append(entry)
				}
				state.Push(result)
				return 1
			},
			"tileset_tiles": func(state *lua.LState) int {
				if session.document == nil || session.catalog == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				path := state.CheckString(1)
				kind, err := mapeditor.ParseLayerKind(state.CheckString(2))
				if err != nil {
					return pushLuaError(state, err)
				}
				result := state.NewTable()
				for _, reference := range session.catalog.References(path) {
					if pickerIncludes(kind, reference.Identity) {
						result.Append(referenceTable(state, reference))
					}
				}
				state.Push(result)
				return 1
			},
			"cell": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				tile, ok := session.document.TileAt(mapeditor.Point{X: state.CheckInt(1), Y: state.CheckInt(2)})
				if !ok {
					return pushLuaError(state, fmt.Errorf("map editor: tile is outside the document"))
				}
				state.Push(session.tileTable(state, tile, mapeditor.Point{X: state.CheckInt(1), Y: state.CheckInt(2)}))
				return 1
			},
			"begin_stroke": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				kind, err := mapeditor.ParseLayerKind(state.CheckString(1))
				if err != nil {
					return pushLuaError(state, err)
				}
				brush, err := brushFromLua(state.CheckTable(3))
				if err != nil {
					return pushLuaError(state, err)
				}
				if err := session.document.BeginStroke(kind, state.CheckInt(2), brush); err != nil {
					return pushLuaError(state, err)
				}
				session.activeBrush = &brush
				session.activeKind = kind
				session.activeLayer = state.CheckInt(2)
				session.activeAuto = state.OptBool(4, false)
				session.activeSourcePath = ""
				if path, ok := state.CheckTable(3).RawGetString("source_path").(lua.LString); ok {
					session.activeSourcePath = string(path)
				}
				if session.activeAuto {
					session.autoModel = newAutoDrawModel(session, kind, session.activeLayer, session.activeSourcePath)
				}
				state.Push(lua.LTrue)
				return 1
			},
			"paint": func(state *lua.LState) int {
				if session.document == nil || session.activeBrush == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no active stroke"))
				}
				changed, err := session.paint(mapeditor.Point{X: state.CheckInt(1), Y: state.CheckInt(2)})
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(lua.LBool(changed))
				return 1
			},
			"end_stroke": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				session.activeBrush = nil
				session.autoModel = nil
				points := session.document.EndStrokePoints()
				state.Push(lua.LBool(len(points) > 0))
				state.Push(pointsTable(state, points))
				return 2
			},
			"cancel_stroke": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				session.activeBrush = nil
				session.autoModel = nil
				state.Push(lua.LBool(session.document.CancelStroke()))
				return 1
			},
			"undo": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				points := session.document.UndoPoints()
				state.Push(lua.LBool(len(points) > 0))
				state.Push(pointsTable(state, points))
				return 2
			},
			"redo": func(state *lua.LState) int {
				if session.document == nil {
					return pushLuaError(state, fmt.Errorf("map editor: no document is open"))
				}
				points := session.document.RedoPoints()
				state.Push(lua.LBool(len(points) > 0))
				state.Push(pointsTable(state, points))
				return 2
			},
			"preview": func(state *lua.LState) int {
				points, err := optionalPoints(state, 1)
				if err != nil {
					return pushLuaError(state, err)
				}
				preview, err := session.preview(points)
				if err != nil {
					return pushLuaError(state, err)
				}
				pushWorldMap(state, preview)
				return 1
			},
			"save": func(state *lua.LState) int {
				path, err := session.save(state.CheckString(1))
				if err != nil {
					return pushLuaError(state, err)
				}
				state.Push(lua.LString(path))
				return 1
			},
		})
		module.RawSetString("api", lua.LNumber(1))
		state.Push(module)
		return 1
	}}
}

type mapEditorSession struct {
	source           fs.FS
	storage          *mapeditor.Storage
	resolvers        []world.ObjectResolver
	document         *mapeditor.Document
	catalog          *world.TileCatalog
	tilePaths        []string
	activeBrush      *mapeditor.Brush
	activeKind       mapeditor.LayerKind
	activeLayer      int
	activeAuto       bool
	activeSourcePath string
	autoModel        *autoDrawModel
	previewWorld     *world.Map
}

// open reads one mounted DS1 and indexes only the DT1 dependencies declared for that map.
// Session state is replaced only after every decode and catalog operation succeeds.
func (session *mapEditorSession) open(path string) error {
	if session.source == nil {
		return fmt.Errorf("map editor: mounted asset filesystem is unavailable")
	}
	data, err := fs.ReadFile(session.source, path)
	if err != nil {
		return fmt.Errorf("map editor: read %q: %w", path, err)
	}
	document, err := mapeditor.Open(data)
	if err != nil {
		return err
	}
	paths, err := assetinspect.DS1TilePaths(session.source, path)
	if err != nil {
		return fmt.Errorf("map editor: resolve tile dependencies: %w", err)
	}
	catalog, err := world.LoadTileCatalog(session.source, paths)
	if err != nil {
		return fmt.Errorf("map editor: index tile dependencies: %w", err)
	}
	session.document, session.catalog = document, catalog
	session.tilePaths = append(session.tilePaths[:0], paths...)
	session.activeBrush = nil
	session.autoModel = nil
	session.previewWorld = nil
	return nil
}

// new creates an empty source document and resolves its caller-supplied DT1 dependencies.
func (session *mapEditorSession) new(config mapeditor.NewConfig, tilePaths []string) error {
	if session.source == nil {
		return fmt.Errorf("map editor: mounted asset filesystem is unavailable")
	}
	document, err := mapeditor.New(config)
	if err != nil {
		return err
	}
	catalog, err := world.LoadTileCatalog(session.source, tilePaths)
	if err != nil {
		return fmt.Errorf("map editor: index tile dependencies: %w", err)
	}
	session.document, session.catalog = document, catalog
	session.tilePaths = append(session.tilePaths[:0], tilePaths...)
	session.activeBrush = nil
	session.autoModel = nil
	session.previewWorld = nil
	return nil
}

// preview returns a presentation snapshot, updating only explicitly dirty cells after initial materialization.
func (session *mapEditorSession) preview(points []mapeditor.Point) (*world.Map, error) {
	if session.document == nil || session.catalog == nil {
		return nil, fmt.Errorf("map editor: no document is open")
	}
	if session.previewWorld == nil {
		stamp, err := session.document.Snapshot()
		if err != nil {
			return nil, err
		}
		preview, err := world.Materialize(stamp, session.catalog, session.resolvers...)
		if err != nil {
			return nil, fmt.Errorf("map editor: materialize preview: %w", err)
		}
		session.previewWorld = preview
	} else {
		for _, point := range points {
			tile, found := session.document.TileAt(point)
			if !found {
				continue
			}
			record := ds1.TileRecord{
				Floors:        tile.Floors,
				Walls:         tile.Walls,
				Shadows:       tile.Shadows,
				Substitutions: tile.Substitutions,
			}
			session.previewWorld.ReplaceAuthoredCell(session.catalog, point.X, point.Y, record)
		}
	}
	return session.previewWorld.PresentationSnapshot(), nil
}

// save encodes the current document and marks it clean only after confined atomic storage succeeds.
func (session *mapEditorSession) save(path string) (string, error) {
	if session.document == nil {
		return "", fmt.Errorf("map editor: no document is open")
	}
	encoded, err := session.document.Encode()
	if err != nil {
		return "", err
	}
	destination, err := session.storage.Save(path, encoded)
	if err != nil {
		return "", err
	}
	session.document.MarkSaved(encoded)
	return destination, nil
}

// pickerIncludes filters catalog identities according to DS1's orientation conventions for each record family.
func pickerIncludes(kind mapeditor.LayerKind, identity world.TileIdentity) bool {
	switch kind {
	case mapeditor.LayerFloor:
		return identity.Orientation == 0
	case mapeditor.LayerShadow:
		return identity.Orientation == 13
	case mapeditor.LayerWall:
		return identity.Orientation != 0 && identity.Orientation != 13
	default:
		return false
	}
}

// fileName extracts a display label from either slash convention while preserving the mounted path itself.
func fileName(path string) string {
	for index := len(path) - 1; index >= 0; index-- {
		if path[index] == '/' || path[index] == '\\' {
			return path[index+1:]
		}
	}
	return path
}

// referenceTable translates one physical DT1 record into a stable Lua table.
func referenceTable(state *lua.LState, reference world.TileReference) *lua.LTable {
	entry := state.NewTable()
	setReferenceFields(entry, reference)
	return entry
}

// setReferenceFields publishes physical tile metadata and compact collision summaries for inspection UI.
func setReferenceFields(entry *lua.LTable, reference world.TileReference) {
	setLuaInteger(entry, "orientation", int(reference.Identity.Orientation))
	setLuaInteger(entry, "style", int(reference.Identity.MainIndex))
	setLuaInteger(entry, "sequence", int(reference.Identity.SubIndex))
	entry.RawSetString("source_path", lua.LString(reference.Path))
	setLuaInteger(entry, "source_index", reference.Index)
	setLuaInteger(entry, "rarity", int(reference.Rarity))
	setLuaInteger(entry, "width", int(reference.Width))
	setLuaInteger(entry, "height", int(reference.Height))
	setLuaInteger(entry, "direction", int(reference.Direction))
	setLuaInteger(entry, "roof_height", int(reference.RoofHeight))
	setLuaInteger(entry, "blocked_walk", countSubtileFlags(reference, blocksWalk))
	setLuaInteger(entry, "blocked_los", countSubtileFlags(reference, blocksLOS))
	setLuaInteger(entry, "blocked_jump", countSubtileFlags(reference, blocksJump))
	setLuaInteger(entry, "blocked_player", countSubtileFlags(reference, blocksPlayer))
	setLuaInteger(entry, "blocked_light", countSubtileFlags(reference, blocksLight))
	label := fmt.Sprintf(
		"%02d / %02d / %02d  #%d",
		reference.Identity.Orientation,
		reference.Identity.MainIndex,
		reference.Identity.SubIndex,
		reference.Index,
	)
	entry.RawSetString("label", lua.LString(label))
}

// countSubtileFlags reduces a physical tile's 5×5 flags to one inspector statistic.
func countSubtileFlags(reference world.TileReference, match func(dt1.SubTileFlags) bool) int {
	count := 0
	for _, flags := range reference.SubTileFlags {
		if match(flags) {
			count++
		}
	}
	return count
}

// blocksWalk selects the corresponding DT1 collision flag for inspector aggregation.
func blocksWalk(flags dt1.SubTileFlags) bool { return flags.BlockWalk }

// blocksLOS selects the corresponding DT1 collision flag for inspector aggregation.
func blocksLOS(flags dt1.SubTileFlags) bool { return flags.BlockLOS }

// blocksJump selects the corresponding DT1 collision flag for inspector aggregation.
func blocksJump(flags dt1.SubTileFlags) bool { return flags.BlockJump }

// blocksPlayer selects the corresponding DT1 collision flag for inspector aggregation.
func blocksPlayer(flags dt1.SubTileFlags) bool { return flags.BlockPlayerWalk }

// blocksLight selects the corresponding DT1 collision flag for inspector aggregation.
func blocksLight(flags dt1.SubTileFlags) bool { return flags.BlockLight }

// pointsTable serializes dirty DS1 coordinates without exposing mutable Go slices.
func pointsTable(state *lua.LState, points []mapeditor.Point) *lua.LTable {
	result := state.NewTable()
	for _, point := range points {
		entry := state.NewTable()
		setLuaInteger(entry, "x", point.X)
		setLuaInteger(entry, "y", point.Y)
		result.Append(entry)
	}
	return result
}

// optionalPoints validates the compact dirty-cell sequence accepted by incremental preview updates.
func optionalPoints(state *lua.LState, index int) ([]mapeditor.Point, error) {
	value := state.Get(index)
	if value == lua.LNil {
		return nil, nil
	}
	table, ok := value.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("map editor: dirty points must be a sequence")
	}
	result := make([]mapeditor.Point, 0, table.Len())
	for item := 1; item <= table.Len(); item++ {
		entry, ok := table.RawGetInt(item).(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("map editor: dirty point %d must be a table", item)
		}
		result = append(result, mapeditor.Point{
			X: int(lua.LVAsNumber(entry.RawGetString("x"))),
			Y: int(lua.LVAsNumber(entry.RawGetString("y"))),
		})
	}
	return result, nil
}

// newDocumentConfig decodes the bounded creation settings shared by the Lua form and Go document model.
func newDocumentConfig(table *lua.LTable) (mapeditor.NewConfig, []string, error) {
	config := mapeditor.NewConfig{
		Width:       int(lua.LVAsNumber(table.RawGetString("width"))),
		Height:      int(lua.LVAsNumber(table.RawGetString("height"))),
		Act:         int(lua.LVAsNumber(table.RawGetString("act"))),
		WallLayers:  int(lua.LVAsNumber(table.RawGetString("wall_layers"))),
		FloorLayers: int(lua.LVAsNumber(table.RawGetString("floor_layers"))),
	}
	paths, err := luaStringList(table.RawGetString("files"))
	if err != nil {
		return mapeditor.NewConfig{}, nil, err
	}
	config.Files = append([]string(nil), paths...)
	return config, paths, nil
}

// luaStringList copies one Lua sequence into Go and rejects sparse or non-string file lists.
func luaStringList(value lua.LValue) ([]string, error) {
	if value == lua.LNil {
		return nil, nil
	}
	table, ok := value.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("map editor: files must be a sequence of strings")
	}
	result := make([]string, 0, table.Len())
	for index := 1; index <= table.Len(); index++ {
		item, ok := table.RawGetInt(index).(lua.LString)
		if !ok {
			return nil, fmt.Errorf("map editor: files must be a sequence of strings")
		}
		result = append(result, string(item))
	}
	return result, nil
}

// brushFromLua validates packed DS1 fields before narrowing them to their serialized widths.
func brushFromLua(table *lua.LTable) (mapeditor.Brush, error) {
	brush := mapeditor.Brush{Empty: lua.LVAsBool(table.RawGetString("empty"))}
	if brush.Empty {
		return brush, nil
	}
	style := float64(lua.LVAsNumber(table.RawGetString("style")))
	sequence := float64(lua.LVAsNumber(table.RawGetString("sequence")))
	orientation := float64(lua.LVAsNumber(table.RawGetString("orientation")))
	if style < 0 || style > 0x3f || sequence < 0 || sequence > 0x3f || orientation < 0 || orientation > 0xff ||
		style != float64(int(style)) || sequence != float64(int(sequence)) || orientation != float64(int(orientation)) {
		return mapeditor.Brush{}, fmt.Errorf(
			"map editor: brush orientation, style, and sequence must be unsigned integers in range",
		)
	}
	properties := float64(lua.LVAsNumber(table.RawGetString("properties")))
	if properties < 0 || properties > float64(^uint32(0)) || properties != float64(uint32(properties)) {
		return mapeditor.Brush{}, fmt.Errorf("map editor: brush properties must be a uint32")
	}
	brush.Identity = mapeditor.Identity{
		Orientation: uint8(orientation),
		Style:       uint8(style),
		Sequence:    uint8(sequence),
	}
	brush.Properties = uint32(properties)
	return brush, nil
}

// summaryTable publishes immutable document chrome state with stable field names.
func summaryTable(state *lua.LState, summary mapeditor.Summary) *lua.LTable {
	result := state.NewTable()
	setLuaInteger(result, "width", summary.Width)
	setLuaInteger(result, "height", summary.Height)
	setLuaInteger(result, "act", summary.Act)
	setLuaInteger(result, "version", int(summary.Version))
	setLuaInteger(result, "wall_layers", summary.WallLayers)
	setLuaInteger(result, "floor_layers", summary.FloorLayers)
	setLuaInteger(result, "shadow_layers", summary.ShadowLayers)
	setLuaInteger(result, "substitution_layers", summary.SubstitutionLayers)
	setLuaInteger(result, "objects", summary.ObjectCount)
	setLuaInteger(result, "dependencies", summary.DependencyCount)
	result.RawSetString("revision", lua.LNumber(summary.Revision))
	result.RawSetString("dirty", lua.LBool(summary.Dirty))
	return result
}

// tileTable serializes every authored record family at one selected DS1 cell.
func (session *mapEditorSession) tileTable(
	state *lua.LState,
	tile mapeditor.Tile,
	point mapeditor.Point,
) *lua.LTable {
	result := state.NewTable()
	result.RawSetString("floors", session.floorTable(state, tile.Floors, point, 0))
	result.RawSetString("walls", session.wallTable(state, tile.Walls, point))
	result.RawSetString("shadows", session.floorTable(state, tile.Shadows, point, 13))
	substitutions := state.NewTable()
	for _, substitution := range tile.Substitutions {
		entry := state.NewTable()
		entry.RawSetString("raw", lua.LNumber(substitution.Unknown))
		substitutions.Append(entry)
	}
	result.RawSetString("substitutions", substitutions)
	return result
}

// floorTable serializes the shared floor/shadow fields and annotates each record with its resolved DT1 tile.
func (session *mapEditorSession) floorTable(
	state *lua.LState,
	values []ds1.FloorShadowRecord,
	point mapeditor.Point,
	orientation int32,
) *lua.LTable {
	result := state.NewTable()
	for _, value := range values {
		entry := state.NewTable()
		entry.RawSetString("properties", lua.LNumber(value.Packed()))
		setLuaInteger(entry, "prop1", int(value.Prop1))
		setLuaInteger(entry, "unknown1", int(value.Unknown1))
		setLuaInteger(entry, "unknown2", int(value.Unknown2))
		setLuaInteger(entry, "style", int(value.Style))
		setLuaInteger(entry, "sequence", int(value.Sequence))
		entry.RawSetString("hidden", lua.LBool(value.Hidden))
		identity := world.TileIdentity{
			Orientation: orientation,
			MainIndex:   int32(value.Style),
			SubIndex:    int32(value.Sequence),
		}
		session.annotateReference(entry, identity, point)
		result.Append(entry)
	}
	return result
}

// wallTable serializes authored wall fields and annotates each record with its resolved physical tile.
func (session *mapEditorSession) wallTable(
	state *lua.LState,
	values []ds1.WallRecord,
	point mapeditor.Point,
) *lua.LTable {
	result := state.NewTable()
	for _, value := range values {
		entry := state.NewTable()
		entry.RawSetString("properties", lua.LNumber(value.Packed()))
		setLuaInteger(entry, "prop1", int(value.Prop1))
		setLuaInteger(entry, "unknown1", int(value.Unknown1))
		setLuaInteger(entry, "unknown2", int(value.Unknown2))
		setLuaInteger(entry, "orientation", int(value.Type))
		setLuaInteger(entry, "style", int(value.Style))
		setLuaInteger(entry, "sequence", int(value.Sequence))
		entry.RawSetString("hidden", lua.LBool(value.Hidden))
		identity := world.TileIdentity{
			Orientation: int32(value.Type),
			MainIndex:   int32(value.Style),
			SubIndex:    int32(value.Sequence),
		}
		session.annotateReference(entry, identity, point)
		result.Append(entry)
	}
	return result
}

// annotateReference resolves the exact deterministic DT1 variant used by the world materializer at this coordinate.
func (session *mapEditorSession) annotateReference(
	table *lua.LTable,
	identity world.TileIdentity,
	point mapeditor.Point,
) {
	if reference, found := session.catalog.Select(identity, point.X, point.Y, 0); found {
		setReferenceFields(table, reference)
	}
}
