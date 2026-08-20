package modruntime

import (
	"github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/mapeditor"
	"github.com/gravestench/ds1"
)

type autoDrawEdge struct {
	center, neighbor world.TileIdentity
	dx, dy           int
}

type autoDrawModel struct {
	kind       mapeditor.LayerKind
	layer      int
	candidates []world.TileIdentity
	permitted  map[world.TileIdentity]bool
	edges      map[autoDrawEdge]int
}

var autoDrawDirections = [...]mapeditor.Point{{X: -1}, {X: 1}, {Y: -1}, {Y: 1}}

// newAutoDrawModel learns directed neighbor frequencies from the open DS1 for one selected DT1 file.
// Confining candidates to that file prevents auto-draw from silently crossing an artist-authored tileset boundary.
func newAutoDrawModel(
	session *mapEditorSession,
	kind mapeditor.LayerKind,
	layer int,
	path string,
) *autoDrawModel {
	if session == nil || session.document == nil || session.catalog == nil || path == "" {
		return nil
	}
	model := &autoDrawModel{
		kind:      kind,
		layer:     layer,
		permitted: make(map[world.TileIdentity]bool),
		edges:     make(map[autoDrawEdge]int),
	}
	for _, reference := range session.catalog.References(path) {
		if pickerIncludes(kind, reference.Identity) && !model.permitted[reference.Identity] {
			model.permitted[reference.Identity] = true
			model.candidates = append(model.candidates, reference.Identity)
		}
	}
	stamp, err := session.document.Snapshot()
	if err != nil || len(model.candidates) == 0 {
		return nil
	}
	for y, row := range stamp.Tiles {
		for x, record := range row {
			center, _, visible := stampIdentity(record, kind, layer)
			if !visible || !model.permitted[center] {
				continue
			}
			for _, direction := range autoDrawDirections {
				nx, ny := x+direction.X, y+direction.Y
				if ny < 0 || ny >= len(stamp.Tiles) || nx < 0 || nx >= len(stamp.Tiles[ny]) {
					continue
				}
				neighbor, _, found := stampIdentity(stamp.Tiles[ny][nx], kind, layer)
				if found {
					model.edges[autoDrawEdge{center: center, neighbor: neighbor, dx: direction.X, dy: direction.Y}]++
				}
			}
		}
	}
	return model
}

// paint applies the active brush and, in auto mode, reconciles compatible immediate neighbors.
// Every adaptation remains in the same document stroke so one undo reverses the complete drag result.
func (session *mapEditorSession) paint(point mapeditor.Point) (bool, error) {
	brush := *session.activeBrush
	if session.activeAuto && session.autoModel != nil && !brush.Empty {
		brush = session.autoModel.choose(session.document, point, brush)
	}
	changed, err := session.document.Paint(point, brush)
	if err != nil || !session.activeAuto || session.autoModel == nil {
		return changed, err
	}
	for _, direction := range autoDrawDirections {
		neighbor := mapeditor.Point{X: point.X + direction.X, Y: point.Y + direction.Y}
		tile, found := session.document.TileAt(neighbor)
		if !found {
			continue
		}
		identity, properties, visible := tileIdentity(tile, session.activeKind, session.activeLayer)
		if !visible || !session.autoModel.permitted[identity] {
			continue
		}
		candidate := session.autoModel.choose(session.document, neighbor, mapeditor.Brush{
			Identity: identityToBrush(identity), Properties: properties,
		})
		neighborChanged, paintErr := session.document.Paint(neighbor, candidate)
		if paintErr != nil {
			return changed, paintErr
		}
		changed = changed || neighborChanged
	}
	return changed, nil
}

// choose scores compatible tile identities against observed neighbor transitions around one target cell.
// A tie deliberately preserves the explicit brush so sparse evidence never creates surprising substitutions.
func (model *autoDrawModel) choose(
	document *mapeditor.Document,
	point mapeditor.Point,
	fallback mapeditor.Brush,
) mapeditor.Brush {
	best := brushIdentity(fallback.Identity)
	bestScore := 1 // a tie always preserves the user's explicit choice
	for _, candidate := range model.candidates {
		score := 0
		for _, direction := range autoDrawDirections {
			tile, found := document.TileAt(mapeditor.Point{X: point.X + direction.X, Y: point.Y + direction.Y})
			if !found {
				continue
			}
			neighbor, _, visible := tileIdentity(tile, model.kind, model.layer)
			if visible {
				score += model.edges[autoDrawEdge{center: candidate, neighbor: neighbor, dx: direction.X, dy: direction.Y}]
			}
		}
		if score > bestScore {
			best, bestScore = candidate, score
		}
	}
	fallback.Identity = identityToBrush(best)
	return fallback
}

// stampIdentity extracts one visible logical identity and its original packed flags from a DS1 tile record.
func stampIdentity(
	record ds1.TileRecord,
	kind mapeditor.LayerKind,
	layer int,
) (world.TileIdentity, uint32, bool) {
	switch kind {
	case mapeditor.LayerFloor:
		if layer < 0 || layer >= len(record.Floors) || record.Floors[layer].Prop1 == 0 || record.Floors[layer].Hidden {
			return world.TileIdentity{}, 0, false
		}
		value := record.Floors[layer]
		return world.TileIdentity{MainIndex: int32(value.Style), SubIndex: int32(value.Sequence)}, value.Packed(), true
	case mapeditor.LayerWall:
		if layer < 0 || layer >= len(record.Walls) || record.Walls[layer].Prop1 == 0 || record.Walls[layer].Hidden {
			return world.TileIdentity{}, 0, false
		}
		value := record.Walls[layer]
		return world.TileIdentity{
			Orientation: int32(value.Type),
			MainIndex:   int32(value.Style),
			SubIndex:    int32(value.Sequence),
		}, value.Packed(), true
	case mapeditor.LayerShadow:
		if layer < 0 || layer >= len(record.Shadows) || record.Shadows[layer].Prop1 == 0 || record.Shadows[layer].Hidden {
			return world.TileIdentity{}, 0, false
		}
		value := record.Shadows[layer]
		return world.TileIdentity{
			Orientation: 13,
			MainIndex:   int32(value.Style),
			SubIndex:    int32(value.Sequence),
		}, value.Packed(), true
	default:
		return world.TileIdentity{}, 0, false
	}
}

// tileIdentity adapts the editor's defensive tile copy to the shared record extractor.
func tileIdentity(
	tile mapeditor.Tile,
	kind mapeditor.LayerKind,
	layer int,
) (world.TileIdentity, uint32, bool) {
	return stampIdentity(ds1.TileRecord{Floors: tile.Floors, Walls: tile.Walls, Shadows: tile.Shadows}, kind, layer)
}

// brushIdentity converts an authored brush key into the world catalog's signed identity type.
func brushIdentity(identity mapeditor.Identity) world.TileIdentity {
	return world.TileIdentity{
		Orientation: int32(identity.Orientation),
		MainIndex:   int32(identity.Style),
		SubIndex:    int32(identity.Sequence),
	}
}

// identityToBrush converts a catalog identity back into DS1's bounded authored fields.
func identityToBrush(identity world.TileIdentity) mapeditor.Identity {
	return mapeditor.Identity{
		Orientation: uint8(identity.Orientation),
		Style:       uint8(identity.MainIndex),
		Sequence:    uint8(identity.SubIndex),
	}
}
