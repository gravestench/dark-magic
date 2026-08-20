// Package mapeditor owns editable DS1 source documents.
//
// It deliberately does not depend on renderer or world packages: a DS1 source
// document retains serialized fields that the runtime map intentionally folds
// into derived presentation and collision facts.
package mapeditor

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/gravestench/ds1"
)

const (
	maxWidth       = 4096
	maxHeight      = 4096
	maxRecordCount = 1 << 20
)

// LayerKind identifies the authored DS1 record family being edited.
type LayerKind uint8

const (
	LayerFloor LayerKind = iota
	LayerWall
	LayerShadow
)

// String returns the stable UI/API spelling for a record family.
func (kind LayerKind) String() string {
	switch kind {
	case LayerFloor:
		return "floor"
	case LayerWall:
		return "wall"
	case LayerShadow:
		return "shadow"
	default:
		return "unknown"
	}
}

// ParseLayerKind converts the deliberately small editor API vocabulary into a
// typed layer selector.
func ParseLayerKind(value string) (LayerKind, error) {
	switch value {
	case "floor":
		return LayerFloor, nil
	case "wall":
		return LayerWall, nil
	case "shadow":
		return LayerShadow, nil
	default:
		return 0, fmt.Errorf("map editor: unknown layer kind %q", value)
	}
}

// Point is a zero-based DS1 tile-grid coordinate.
type Point struct{ X, Y int }

// Identity is the logical DT1 lookup key authored by a DS1 record. Physical
// DT1 indexes are intentionally absent: they are selected later by the tile
// catalog using orientation, style, sequence, and rarity.
type Identity struct {
	Orientation uint8
	Style       uint8
	Sequence    uint8
}

// Brush describes one authored result. Properties, when non-zero, preserves a
// caller-selected packed DS1 cell value; Style and Sequence are then applied
// explicitly so a tile picker cannot accidentally paint a different identity.
//
// Empty clears the selected record family. The remaining fields are ignored in
// that case.
type Brush struct {
	Identity   Identity
	Properties uint32
	Empty      bool
}

// NewConfig describes a modern, empty v18 map. Width and Height are serialized
// DS1 grid dimensions, including the terminal row and column represented by
// the codec model.
type NewConfig struct {
	Width, Height int
	Act           int
	Files         []string
	WallLayers    int
	FloorLayers   int
}

// Summary is a cheap immutable document projection for UI chrome.
type Summary struct {
	Width, Height           int
	Act                     int
	Version                 ds1.Version
	WallLayers, FloorLayers int
	ShadowLayers            int
	SubstitutionLayers      int
	ObjectCount             int
	DependencyCount         int
	Revision                uint64
	Dirty                   bool
}

// Tile is a defensive copy of all source records at one DS1 grid position.
type Tile struct {
	Floors        []ds1.FloorShadowRecord
	Walls         []ds1.WallRecord
	Shadows       []ds1.FloorShadowRecord
	Substitutions []ds1.SubstitutionRecord
}

type mutation struct {
	point       Point
	beforeFloor ds1.FloorShadowRecord
	afterFloor  ds1.FloorShadowRecord
	beforeWall  ds1.WallRecord
	afterWall   ds1.WallRecord
}

type historyEntry struct {
	kind      LayerKind
	layer     int
	mutations []mutation
}

type activeStroke struct {
	entry  historyEntry
	points map[Point]int
}

// Document serializes all source mutations through one lock and tracks
// coalesced brush strokes for ordinary editor undo/redo.
type Document struct {
	mu       sync.RWMutex
	stamp    *ds1.DS1
	saved    []byte
	revision uint64
	history  []historyEntry
	cursor   int
	active   *activeStroke
}

// Open decodes one DS1 byte stream into an editable source document.
func Open(data []byte) (*Document, error) {
	stamp, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("map editor: decode DS1: %w", err)
	}

	canonical, err := ds1.Encode(stamp)
	if err != nil {
		return nil, fmt.Errorf("map editor: validate decoded DS1: %w", err)
	}

	return &Document{stamp: stamp, saved: canonical, revision: 1}, nil
}

// New creates a modern v18 document. Existing maps should be opened rather
// than rebuilt so their version-specific fields remain lossless.
func New(config NewConfig) (*Document, error) {
	if config.Width <= 0 || config.Width > maxWidth || config.Height <= 0 || config.Height > maxHeight {
		return nil, fmt.Errorf("map editor: invalid dimensions %dx%d", config.Width, config.Height)
	}
	// Check before allocating the tile grid. The DS1 codec rejects oversized
	// documents too, but doing it here prevents an editor request from first
	// allocating a huge, unencodable in-memory map.
	if int64(config.Width)*int64(config.Height) > maxRecordCount {
		return nil, fmt.Errorf("map editor: invalid dimensions %dx%d", config.Width, config.Height)
	}
	if config.Act == 0 {
		config.Act = 1
	}
	if config.Act < 1 || config.Act > 5 {
		return nil, fmt.Errorf("map editor: invalid act %d", config.Act)
	}
	if config.WallLayers == 0 {
		config.WallLayers = 1
	}
	if config.FloorLayers == 0 {
		config.FloorLayers = 1
	}
	if config.WallLayers < 0 || config.WallLayers > 4 || config.FloorLayers < 0 || config.FloorLayers > 2 {
		return nil, fmt.Errorf("map editor: invalid layer counts %d walls, %d floors", config.WallLayers, config.FloorLayers)
	}

	stamp := &ds1.DS1{
		Version:                    ds1.LatestVersion,
		Width:                      int32(config.Width),
		Height:                     int32(config.Height),
		Act:                        int32(config.Act),
		Files:                      append([]string(nil), config.Files...),
		NumberOfWalls:              int32(config.WallLayers),
		NumberOfFloors:             int32(config.FloorLayers),
		NumberOfShadowLayers:       1,
		NumberOfSubstitutionLayers: 0,
	}
	stamp.Tiles = make([][]ds1.TileRecord, config.Height)
	for y := range stamp.Tiles {
		stamp.Tiles[y] = make([]ds1.TileRecord, config.Width)
		for x := range stamp.Tiles[y] {
			stamp.Tiles[y][x] = newTile(config.WallLayers, config.FloorLayers, 1, 0)
		}
	}

	encoded, err := ds1.Encode(stamp)
	if err != nil {
		return nil, fmt.Errorf("map editor: create DS1: %w", err)
	}
	return &Document{stamp: stamp, saved: encoded, revision: 1}, nil
}

// newTile allocates every declared record family so later edits can index layers without reshaping the DS1 grid.
func newTile(walls, floors, shadows, substitutions int) ds1.TileRecord {
	return ds1.TileRecord{
		Walls:         make([]ds1.WallRecord, walls),
		Floors:        make([]ds1.FloorShadowRecord, floors),
		Shadows:       make([]ds1.FloorShadowRecord, shadows),
		Substitutions: make([]ds1.SubstitutionRecord, substitutions),
	}
}

// Snapshot returns a validated deep copy. Callers may materialize or inspect
// the result without retaining writable document memory.
func (document *Document) Snapshot() (*ds1.DS1, error) {
	document.mu.RLock()
	defer document.mu.RUnlock()
	return clone(document.stamp)
}

// Encode returns a validated, lossless DS1 payload. It does not change dirty
// state; call MarkSaved only after a durable write succeeds.
func (document *Document) Encode() ([]byte, error) {
	document.mu.RLock()
	defer document.mu.RUnlock()
	return ds1.Encode(document.stamp)
}

// MarkSaved establishes the current validated encoding as the document's
// durable checkpoint. It is intentionally separate from Encode so storage can
// remain an explicitly scoped capability.
func (document *Document) MarkSaved(encoded []byte) {
	document.mu.Lock()
	defer document.mu.Unlock()
	document.saved = append(document.saved[:0], encoded...)
}

// Summary returns document chrome state without exposing the mutable model.
func (document *Document) Summary() Summary {
	document.mu.RLock()
	defer document.mu.RUnlock()
	encoded, err := ds1.Encode(document.stamp)
	return Summary{
		Width:              int(document.stamp.Width),
		Height:             int(document.stamp.Height),
		Act:                int(document.stamp.Act),
		Version:            document.stamp.Version,
		WallLayers:         int(document.stamp.NumberOfWalls),
		FloorLayers:        int(document.stamp.NumberOfFloors),
		ShadowLayers:       int(document.stamp.NumberOfShadowLayers),
		SubstitutionLayers: int(document.stamp.NumberOfSubstitutionLayers),
		ObjectCount:        len(document.stamp.Objects),
		DependencyCount:    len(document.stamp.Files),
		Revision:           document.revision,
		Dirty:              err != nil || !bytes.Equal(encoded, document.saved),
	}
}

// TileAt returns copied source records for one DS1 coordinate.
func (document *Document) TileAt(point Point) (Tile, bool) {
	document.mu.RLock()
	defer document.mu.RUnlock()
	if !document.inBounds(point) {
		return Tile{}, false
	}
	record := document.stamp.Tiles[point.Y][point.X]
	return Tile{
		Floors:        append([]ds1.FloorShadowRecord(nil), record.Floors...),
		Walls:         append([]ds1.WallRecord(nil), record.Walls...),
		Shadows:       append([]ds1.FloorShadowRecord(nil), record.Shadows...),
		Substitutions: append([]ds1.SubstitutionRecord(nil), record.Substitutions...),
	}, true
}

// BeginStroke starts one undoable brush gesture. Call Paint while the pointer
// moves, then EndStroke on release. Repeated visits to a cell remain one
// mutation whose original value is retained for undo.
func (document *Document) BeginStroke(kind LayerKind, layer int, brush Brush) error {
	document.mu.Lock()
	defer document.mu.Unlock()
	if document.active != nil {
		return fmt.Errorf("map editor: a stroke is already active")
	}
	if err := document.validateLayer(kind, layer); err != nil {
		return err
	}
	if !brush.Empty && (brush.Identity.Style > 0x3f || brush.Identity.Sequence > 0x3f) {
		return fmt.Errorf("map editor: style and sequence must fit DS1's 6-bit fields")
	}
	document.active = &activeStroke{
		entry:  historyEntry{kind: kind, layer: layer},
		points: make(map[Point]int),
	}
	return nil
}

// Paint applies the active brush at one tile-grid coordinate.
func (document *Document) Paint(point Point, brush Brush) (bool, error) {
	document.mu.Lock()
	defer document.mu.Unlock()
	if document.active == nil {
		return false, fmt.Errorf("map editor: no active stroke")
	}
	if !document.inBounds(point) {
		return false, fmt.Errorf(
			"map editor: tile (%d,%d) is outside %dx%d",
			point.X,
			point.Y,
			document.stamp.Width,
			document.stamp.Height,
		)
	}
	stroke := document.active
	changed, mutation := document.paintLocked(stroke.entry.kind, stroke.entry.layer, point, brush)
	if !changed {
		return false, nil
	}
	if index, exists := stroke.points[point]; exists {
		if stroke.entry.kind == LayerWall {
			stroke.entry.mutations[index].afterWall = mutation.afterWall
		} else {
			stroke.entry.mutations[index].afterFloor = mutation.afterFloor
		}
	} else {
		stroke.points[point] = len(stroke.entry.mutations)
		stroke.entry.mutations = append(stroke.entry.mutations, mutation)
	}
	document.revision++
	return true, nil
}

// EndStroke commits the active gesture to history and reports whether it
// changed any source record.
func (document *Document) EndStroke() bool {
	return len(document.EndStrokePoints()) > 0
}

// EndStrokePoints commits the active gesture and returns the distinct cells
// whose derived preview may have changed. The compact dirty set lets editor
// presentation invalidate only intersecting chunks.
func (document *Document) EndStrokePoints() []Point {
	document.mu.Lock()
	defer document.mu.Unlock()
	if document.active == nil {
		return nil
	}
	stroke := document.active
	document.active = nil
	if len(stroke.entry.mutations) == 0 {
		return nil
	}
	document.history = append(document.history[:document.cursor], stroke.entry)
	document.cursor = len(document.history)
	return mutationPoints(stroke.entry.mutations)
}

// CancelStroke restores every mutation made during the active drag.
func (document *Document) CancelStroke() bool {
	document.mu.Lock()
	defer document.mu.Unlock()
	if document.active == nil {
		return false
	}
	for index := len(document.active.entry.mutations) - 1; index >= 0; index-- {
		document.applyMutation(
			document.active.entry.kind,
			document.active.entry.layer,
			document.active.entry.mutations[index],
			false,
		)
	}
	changed := len(document.active.entry.mutations) > 0
	document.active = nil
	if changed {
		document.revision++
	}
	return changed
}

// Undo restores the preceding completed stroke.
func (document *Document) Undo() bool {
	return len(document.UndoPoints()) > 0
}

// UndoPoints restores the preceding stroke and returns its dirty cells.
func (document *Document) UndoPoints() []Point {
	document.mu.Lock()
	defer document.mu.Unlock()
	if document.active != nil || document.cursor == 0 {
		return nil
	}
	document.cursor--
	entry := document.history[document.cursor]
	for index := len(entry.mutations) - 1; index >= 0; index-- {
		document.applyMutation(entry.kind, entry.layer, entry.mutations[index], false)
	}
	document.revision++
	return mutationPoints(entry.mutations)
}

// Redo reapplies the next undone stroke.
func (document *Document) Redo() bool {
	return len(document.RedoPoints()) > 0
}

// RedoPoints reapplies the next stroke and returns its dirty cells.
func (document *Document) RedoPoints() []Point {
	document.mu.Lock()
	defer document.mu.Unlock()
	if document.active != nil || document.cursor >= len(document.history) {
		return nil
	}
	entry := document.history[document.cursor]
	for _, mutation := range entry.mutations {
		document.applyMutation(entry.kind, entry.layer, mutation, true)
	}
	document.cursor++
	document.revision++
	return mutationPoints(entry.mutations)
}

// mutationPoints preserves stroke order while exposing only the cells presentation must invalidate.
func mutationPoints(mutations []mutation) []Point {
	result := make([]Point, 0, len(mutations))
	for _, mutation := range mutations {
		result = append(result, mutation.point)
	}
	return result
}

// paintLocked applies one brush result while the caller owns the document lock.
// It returns the complete before/after pair needed to coalesce, undo, or redo the edit.
func (document *Document) paintLocked(
	kind LayerKind,
	layer int,
	point Point,
	brush Brush,
) (bool, mutation) {
	record := &document.stamp.Tiles[point.Y][point.X]
	result := mutation{point: point}
	switch kind {
	case LayerWall:
		before := record.Walls[layer]
		after := wallRecord(brush)
		if before == after {
			return false, mutation{}
		}
		record.Walls[layer] = after
		result.beforeWall, result.afterWall = before, after
	case LayerFloor:
		before := record.Floors[layer]
		after := floorRecord(brush, false)
		if before == after {
			return false, mutation{}
		}
		record.Floors[layer] = after
		result.beforeFloor, result.afterFloor = before, after
	case LayerShadow:
		before := record.Shadows[layer]
		after := floorRecord(brush, true)
		if before == after {
			return false, mutation{}
		}
		record.Shadows[layer] = after
		result.beforeFloor, result.afterFloor = before, after
	}
	return true, result
}

// wallRecord translates the editor brush into one DS1 wall record without discarding caller-authored flag bits.
func wallRecord(brush Brush) ds1.WallRecord {
	if brush.Empty {
		return ds1.WallRecord{}
	}
	record := ds1.WallRecord{}
	properties := brush.Properties
	if properties == 0 {
		properties = 0x01 // authored wall bit
	}
	record.SetPacked(properties)
	record.Type = ds1.TileType(brush.Identity.Orientation)
	record.RawOrientation = brush.Identity.Orientation
	record.Style, record.Sequence = brush.Identity.Style, brush.Identity.Sequence
	return record
}

// floorRecord translates a brush into the shared DS1 floor/shadow representation.
// Default packed bits differ because an authored shadow must retain its format-specific visibility bit.
func floorRecord(brush Brush, shadow bool) ds1.FloorShadowRecord {
	if brush.Empty {
		return ds1.FloorShadowRecord{}
	}
	record := ds1.FloorShadowRecord{}
	properties := brush.Properties
	if properties == 0 {
		if shadow {
			properties = 0x08000001 // visible record plus the shadow/roof-like bit
		} else {
			properties = 0x02 // authored floor bit
		}
	}
	record.SetPacked(properties)
	record.Style, record.Sequence = brush.Identity.Style, brush.Identity.Sequence
	return record
}

// applyMutation selects either side of a recorded edit and writes it back to the matching DS1 layer.
// Callers use reverse history order for undo so overlapping gesture history remains correct.
func (document *Document) applyMutation(kind LayerKind, layer int, mutation mutation, forward bool) {
	record := &document.stamp.Tiles[mutation.point.Y][mutation.point.X]
	switch kind {
	case LayerWall:
		if forward {
			record.Walls[layer] = mutation.afterWall
		} else {
			record.Walls[layer] = mutation.beforeWall
		}
	case LayerFloor:
		if forward {
			record.Floors[layer] = mutation.afterFloor
		} else {
			record.Floors[layer] = mutation.beforeFloor
		}
	case LayerShadow:
		if forward {
			record.Shadows[layer] = mutation.afterFloor
		} else {
			record.Shadows[layer] = mutation.beforeFloor
		}
	}
}

// validateLayer rejects absent record arrays before a stroke can retain an invalid layer index.
func (document *Document) validateLayer(kind LayerKind, layer int) error {
	if layer < 0 {
		return fmt.Errorf("map editor: negative layer %d", layer)
	}
	var count int32
	switch kind {
	case LayerFloor:
		count = document.stamp.NumberOfFloors
	case LayerWall:
		count = document.stamp.NumberOfWalls
	case LayerShadow:
		count = document.stamp.NumberOfShadowLayers
	default:
		return fmt.Errorf("map editor: invalid layer kind %d", kind)
	}
	if layer >= int(count) {
		return fmt.Errorf("map editor: %s layer %d is unavailable", kind, layer)
	}
	return nil
}

// inBounds applies the DS1 codec's serialized terminal row and column dimensions exactly.
func (document *Document) inBounds(point Point) bool {
	return point.X >= 0 && point.Y >= 0 && point.X < int(document.stamp.Width) && point.Y < int(document.stamp.Height)
}

// clone round-trips through the codec to provide a validated deep copy without sharing source slices.
func clone(value *ds1.DS1) (*ds1.DS1, error) {
	encoded, err := ds1.Encode(value)
	if err != nil {
		return nil, fmt.Errorf("map editor: validate document: %w", err)
	}
	result, err := ds1.FromBytes(encoded)
	if err != nil {
		return nil, fmt.Errorf("map editor: clone document: %w", err)
	}
	return result, nil
}
