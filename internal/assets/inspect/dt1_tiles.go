package assetinspect

import (
	"fmt"
	"image/color"
	"io"
	"io/fs"

	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
)

// tileKey identifies the DT1 variants that can satisfy one DS1 placement. The
// remaining tile metadata only affects deterministic selection and drawing.
type tileKey struct {
	tileType int32
	style    int32
	sequence int32
}

// texturedDS1TileKeys finds only visible placements. Restricting indexing to
// this set avoids decoding unrelated graphics while retaining north-corner pairs.
func texturedDS1TileKeys(stamp *ds1.DS1) map[tileKey]struct{} {
	result := make(map[tileKey]struct{})

	for _, row := range stamp.Tiles {
		for _, record := range row {
			collectFloorTileKeys(result, record)
			collectShadowTileKeys(result, record)
			collectWallTileKeys(result, record)
		}
	}

	return result
}

// collectFloorTileKeys records visible floor graphics; hidden and empty records
// cannot contribute pixels and therefore must not force unnecessary DT1 decoding.
func collectFloorTileKeys(result map[tileKey]struct{}, record ds1.TileRecord) {
	for _, floor := range record.Floors {
		if floor.Hidden || floor.Prop1 == 0 {
			continue
		}

		result[tileKey{
			tileType: 0,
			style:    int32(floor.Style),
			sequence: int32(floor.Sequence),
		}] = struct{}{}
	}
}

// collectShadowTileKeys records visible shadow graphics under their fixed DT1
// type, preserving the renderer's separate shadow compositing phase.
func collectShadowTileKeys(result map[tileKey]struct{}, record ds1.TileRecord) {
	for _, shadow := range record.Shadows {
		if shadow.Hidden || shadow.Prop1 == 0 {
			continue
		}

		result[tileKey{
			tileType: 13,
			style:    int32(shadow.Style),
			sequence: int32(shadow.Sequence),
		}] = struct{}{}
	}
}

// collectWallTileKeys includes the paired type-4 graphic for north corners so a
// visible type-3 placement can be drawn as the two-part structure used by DT1.
func collectWallTileKeys(result map[tileKey]struct{}, record ds1.TileRecord) {
	for _, wall := range record.Walls {
		if wall.Hidden || wall.Prop1 == 0 {
			continue
		}

		key := tileKey{
			tileType: int32(wall.Type),
			style:    int32(wall.Style),
			sequence: int32(wall.Sequence),
		}

		result[key] = struct{}{}

		if key.tileType == 3 {
			key.tileType = 4
			result[key] = struct{}{}
		}
	}
}

// loadNeededDT1Tiles indexes libraries in caller-provided order and closes each
// one before advancing, preserving both candidate ordering and file ownership.
func loadNeededDT1Tiles(
	source fs.FS,
	dt1Paths []string,
	palette color.Palette,
	needed map[tileKey]struct{},
) (map[tileKey][]*dt1.Tile, error) {
	lookup := make(map[tileKey][]*dt1.Tile)

	for _, path := range dt1Paths {
		file, err := source.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening DT1 asset %q: %w", path, err)
		}

		tileset, err := openDT1Index(file)
		if err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("opening DT1 index %q: %w", path, err)
		}

		if palette != nil {
			tileset.SetPalette(palette)
		}

		if err := appendNeededDT1Tiles(lookup, tileset, needed, path); err != nil {
			_ = file.Close()
			return nil, err
		}

		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("closing DT1 asset %q: %w", path, closeErr)
		}
	}

	return lookup, nil
}

// openDT1Index uses random access when the filesystem supports it and falls back
// to an in-memory index otherwise, allowing archive and directory sources alike.
func openDT1Index(file fs.File) (*dt1.File, error) {
	var tileset *dt1.File

	var err error

	if reader, ok := file.(io.ReaderAt); ok {
		if info, statErr := file.Stat(); statErr == nil {
			tileset, err = dt1.Open(reader, info.Size())
		}
	}

	if tileset == nil && err == nil {
		data, readErr := io.ReadAll(file)
		if readErr != nil {
			return nil, readErr
		}

		tileset, err = dt1.OpenBytes(data)
	}

	return tileset, err
}

// appendNeededDT1Tiles decodes matching tiles in on-disk order. That order is
// later significant to weighted variant selection and must remain deterministic.
func appendNeededDT1Tiles(
	lookup map[tileKey][]*dt1.Tile,
	tileset *dt1.File,
	needed map[tileKey]struct{},
	path string,
) error {
	for index := 0; index < tileset.NumTiles(); index++ {
		metadata, err := tileset.TileMetadata(index)
		if err != nil {
			return fmt.Errorf("indexing DT1 asset %q tile %d: %w", path, index, err)
		}

		key := tileKey{
			tileType: metadata.Type,
			style:    metadata.Style,
			sequence: metadata.Sequence,
		}
		if _, used := needed[key]; !used {
			continue
		}

		tile, err := tileset.DecodeTile(index)
		if err != nil {
			return fmt.Errorf("decoding DT1 asset %q tile %d: %w", path, index, err)
		}

		lookup[key] = append(lookup[key], tile)
	}

	return nil
}
