package world

import (
	"fmt"
	"io/fs"

	"github.com/gravestench/ds1"
)

// Load joins one DS1 stamp with its DT1 collision definitions. It decodes no renderer textures and performs no entity
// spawning, leaving presentation and ECS ownership outside the map-data boundary.
func Load(source fs.FS, stampPath string, tilePaths []string, resolvers ...ObjectResolver) (*Map, error) {
	// Preserve the public loader's useful error ordering: the requested stamp is the primary resource, so report a
	// missing DS1 before inspecting its DT1 dependencies. Materialization bypasses this extra probe for each room.
	if _, err := fs.Stat(source, stampPath); err != nil {
		return nil, fmt.Errorf("world: open %q: %w", stampPath, err)
	}

	catalog, err := LoadTileCatalog(source, tilePaths)
	if err != nil {
		return nil, err
	}

	var resolver ObjectResolver
	if len(resolvers) > 0 {
		// The variadic shape preserves the public API; only the first resolver has historically participated in loading.
		resolver = resolvers[0]
	}

	return loadStamp(source, stampPath, catalog, resolver)
}

// loadStamp owns DS1 file lifetime and turns the decoded stamp into gameplay facts using an already-loaded tile
// catalog. Reusing a catalog is essential when a generated zone contains many rooms with the same DT1 dependencies.
func loadStamp(source fs.FS, stampPath string, catalog *TileCatalog, resolver ObjectResolver) (*Map, error) {
	stamp, err := decodeStamp(source, stampPath)
	if err != nil {
		return nil, err
	}

	result := newMapForStamp(stamp)
	appendStampObjects(result, stamp, resolver)
	appendStampTiles(result, stamp, catalog)

	return result, nil
}

// decodeStamp closes the DS1 before returning and preserves decode-before-close error precedence. Callers therefore get
// the most actionable content error when both parsing and cleanup happen to fail.
func decodeStamp(source fs.FS, stampPath string) (*ds1.DS1, error) {
	stampFile, err := source.Open(stampPath)
	if err != nil {
		return nil, fmt.Errorf("world: open %q: %w", stampPath, err)
	}

	stamp, decodeErr := ds1.FromReader(stampFile)
	closeErr := stampFile.Close()

	if decodeErr != nil {
		return nil, fmt.Errorf("world: decode DS1 %q: %w", stampPath, decodeErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("world: close DS1 %q: %w", stampPath, closeErr)
	}

	return stamp, nil
}

// newMapForStamp establishes tile and subtile dimensions together so every later collision write uses a backing slice
// sized from the same decoded geometry.
func newMapForStamp(stamp *ds1.DS1) *Map {
	result := &Map{
		WidthTiles:     int(stamp.Width),
		HeightTiles:    int(stamp.Height),
		WidthSubtiles:  int(stamp.Width) * SubtilesPerTile,
		HeightSubtiles: int(stamp.Height) * SubtilesPerTile,
		Act:            int(stamp.Act),
	}
	result.flags = make([]Flags, result.WidthSubtiles*result.HeightSubtiles)

	return result
}

// appendStampObjects preserves authored object order while optionally enriching records through the act-local resolver.
// Stable order is used when deriving selectable IDs, so sorting here would change externally visible identities.
func appendStampObjects(result *Map, stamp *ds1.DS1, resolver ObjectResolver) {
	for _, object := range stamp.Objects {
		decoded := resolveObject(result.Act, object.Type, object.ID, object.X, object.Y, object.Flags, resolver)
		result.Objects = append(result.Objects, decoded)
	}
}

// appendStampTiles walks DS1 rows in authored order and delegates each record family to a focused materializer. The
// traversal order feeds deterministic DT1 rarity selection and must remain row-major.
func appendStampTiles(result *Map, stamp *ds1.DS1, catalog *TileCatalog) {
	for tileY, row := range stamp.Tiles {
		for tileX, record := range row {
			result.appendFloorRecords(catalog, tileX, tileY, record.Floors)
			result.appendWallRecords(catalog, tileX, tileY, record.Walls)
			result.appendShadowRecords(catalog, tileX, tileY, record.Shadows)
		}
	}
}

// appendFloorRecords materializes visible authored floors. Hidden or zero-property records carry no rendered/collision
// tile and are intentionally skipped exactly as in the legacy loader.
func (m *Map) appendFloorRecords(catalog *TileCatalog, tileX, tileY int, floors []ds1.FloorShadowRecord) {
	for _, floor := range floors {
		if floor.Hidden || floor.Prop1 == 0 {
			continue
		}

		identity := TileIdentity{MainIndex: int32(floor.Style), SubIndex: int32(floor.Sequence)}
		m.addTile(catalog, tileX, tileY, LayerFloor, identity)
	}
}

// appendWallRecords retains hidden special markers while materializing only visible wall pixels. Special orientation
// facts drive level transitions, so hiding their visual wall must not erase their gameplay meaning.
func (m *Map) appendWallRecords(catalog *TileCatalog, tileX, tileY int, walls []ds1.WallRecord) {
	for _, wall := range walls {
		identity := TileIdentity{
			Orientation: int32(wall.Type),
			MainIndex:   int32(wall.Style),
			SubIndex:    int32(wall.Sequence),
		}
		m.appendSpecialTile(tileX, tileY, identity, wall.Hidden)

		if wall.Hidden || wall.Prop1 == 0 {
			continue
		}

		layer := wallLayer(identity.Orientation)
		m.addTile(catalog, tileX, tileY, layer, identity)
		m.appendNorthCornerCompanion(catalog, tileX, tileY, layer, identity)
	}
}

// appendSpecialTile records orientation-10/11 wall semantics independently from whether the visual tile is hidden.
func (m *Map) appendSpecialTile(tileX, tileY int, identity TileIdentity, hidden bool) {
	if identity.Orientation != 10 && identity.Orientation != 11 {
		return
	}

	m.SpecialTiles = append(m.SpecialTiles, SpecialTile{
		X:           tileX,
		Y:           tileY,
		Orientation: identity.Orientation,
		MainIndex:   identity.MainIndex,
		SubIndex:    identity.SubIndex,
		Hidden:      hidden,
	})
}

// wallLayer maps legacy orientations into global presentation passes. Keeping this decision beside wall loading makes
// the roof/lower-wall exceptions visible without mixing them into DS1 traversal.
func wallLayer(orientation int32) TileLayer {
	if orientation >= 16 && orientation <= 19 {
		return LayerLowerWall
	}

	if orientation == 15 {
		return LayerRoof
	}

	return LayerUpperWall
}

// appendNorthCornerCompanion expands authored orientation 3 into the paired type-3/type-4 DT1 records drawn on one
// baseline. Selection order stays primary then companion, matching the legacy presentation contract.
func (m *Map) appendNorthCornerCompanion(
	catalog *TileCatalog,
	tileX, tileY int,
	layer TileLayer,
	identity TileIdentity,
) {
	if identity.Orientation != 3 {
		return
	}

	companion := identity
	companion.Orientation = 4
	m.addTile(catalog, tileX, tileY, layer, companion)
}

// appendShadowRecords materializes visible shadows without letting their visual metadata affect gameplay collision.
func (m *Map) appendShadowRecords(catalog *TileCatalog, tileX, tileY int, shadows []ds1.FloorShadowRecord) {
	for _, shadow := range shadows {
		if shadow.Hidden || shadow.Prop1 == 0 {
			continue
		}

		identity := TileIdentity{
			Orientation: 13,
			MainIndex:   int32(shadow.Style),
			SubIndex:    int32(shadow.Sequence),
		}
		m.addTile(catalog, tileX, tileY, LayerShadow, identity)
	}
}

// resolveObject preserves raw authored placement even when no catalog mapping exists. Enrichment is type-specific so a
// static mapping can never be applied to a dynamic record with the same numeric ID.
func resolveObject(act int, objectType, id, x, y, flags int32, resolver ObjectResolver) Object {
	result := Object{Type: objectType, ID: id, X: x, Y: y, Flags: flags}
	if resolver == nil {
		return result
	}

	switch objectType {
	case ObjectTypeStatic:
		result.ObjectID, result.Description, result.Resolved = resolver.ResolveStaticObject(act, int(id))
	case ObjectTypeDynamic:
		result.Class, result.Resolved = resolver.ResolveDynamicObject(act, int(id))
	}

	return result
}

// addTile performs deterministic catalog selection before appending placement and collision. Shadows and roofs remain
// visual-only layers; allowing them to mutate collision would create invisible blockers.
func (m *Map) addTile(catalog *TileCatalog, tileX, tileY int, layer TileLayer, identity TileIdentity) {
	reference, found := catalog.Select(identity, tileX, tileY, 0)
	if !found {
		return
	}

	m.Tiles = append(m.Tiles, TilePlacement{
		X: tileX, Y: tileY, Layer: layer, Identity: identity, Reference: reference,
	})
	if layer != LayerShadow && layer != LayerRoof {
		m.apply(tileX, tileY, reference)
	}
}
