// Package world decodes deterministic gameplay facts from DS1 stamps and DT1
// tilesets. It does not own presentation textures or native renderer state.
package world

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
)

// SubtilesPerTile is the fixed collision resolution encoded by DT1 tiles.
const SubtilesPerTile = 5

const (
	// Tile pixel dimensions describe isometric presentation space only; collision
	// and navigation use subtile coordinates.
	TilePixelWidth  = 160
	TilePixelHeight = 80
	PreviewMargin   = 160
)

const (
	// Object type values are authored DS1 record identities.
	ObjectTypeDynamic int32 = 1
	ObjectTypeStatic  int32 = 2
)

// Flags is the gameplay-relevant union of DT1 subtile collision bits.
type Flags struct {
	BlockWalk, BlockLOS, BlockJump, BlockPlayerWalk, BlockLight bool
}

// Blocked reports whether a player-sized point cannot walk through this
// subtile. BlockWalk is shared terrain collision; BlockPlayerWalk is the
// additional player-specific restriction encoded by DT1.
func (f Flags) Blocked() bool { return f.BlockWalk || f.BlockPlayerWalk }

// Object preserves authored DS1 placement plus optional catalog resolution.
// Loading identifies objects; authoritative systems decide whether to spawn them.
type Object struct {
	Type, ID, X, Y, Flags int32
	ObjectID              int
	Class                 string
	Description           string
	Resolved              bool
}

// ObjectResolver supplies act-local recovered identity joins without coupling
// world decoding to a concrete catalog generation.
type ObjectResolver interface {
	ResolveStaticObject(act, id int) (objectID int, description string, found bool)
	ResolveDynamicObject(act, id int) (class string, found bool)
}

// Map is an immutable decoded stamp in tile/subtile coordinates.
type Map struct {
	WidthTiles, HeightTiles       int
	WidthSubtiles, HeightSubtiles int
	Act                           int
	Objects                       []Object
	flags                         []Flags
}

type tileKey struct{ kind, style, sequence int32 }

// Load joins one DS1 stamp with its DT1 collision definitions. It decodes no
// renderer textures and performs no entity spawning.
func Load(source fs.FS, stampPath string, tilePaths []string, resolvers ...ObjectResolver) (*Map, error) {
	stampFile, err := source.Open(stampPath)
	if err != nil {
		return nil, fmt.Errorf("world: open %q: %w", stampPath, err)
	}
	stamp, err := ds1.FromReader(stampFile)
	closeErr := stampFile.Close()
	if err != nil {
		return nil, fmt.Errorf("world: decode DS1 %q: %w", stampPath, err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("world: close DS1 %q: %w", stampPath, closeErr)
	}
	lookup := make(map[tileKey][]*dt1.Tile)
	for _, path := range tilePaths {
		file, err := source.Open(path)
		if err != nil {
			return nil, fmt.Errorf("world: open %q: %w", path, err)
		}
		var opened *dt1.File
		if reader, ok := file.(io.ReaderAt); ok {
			info, statErr := file.Stat()
			if statErr == nil {
				opened, err = dt1.Open(reader, info.Size())
			}
		}
		if opened == nil && err == nil {
			data, readErr := io.ReadAll(file)
			if readErr != nil {
				err = readErr
			} else {
				opened, err = dt1.OpenBytes(data)
			}
		}
		if err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("world: decode DT1 %q: %w", path, err)
		}
		for index := 0; index < opened.NumTiles(); index++ {
			tile, metadataErr := opened.TileMetadata(index)
			if metadataErr != nil {
				_ = file.Close()
				return nil, fmt.Errorf("world: index DT1 %q tile %d: %w", path, index, metadataErr)
			}
			key := tileKey{kind: tile.Type, style: tile.Style, sequence: tile.Sequence}
			lookup[key] = append(lookup[key], tile)
		}
		closeErr := file.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("world: close DT1 %q: %w", path, closeErr)
		}
	}
	result := &Map{
		WidthTiles: int(stamp.Width), HeightTiles: int(stamp.Height), Act: int(stamp.Act),
		WidthSubtiles: int(stamp.Width) * SubtilesPerTile, HeightSubtiles: int(stamp.Height) * SubtilesPerTile,
	}
	result.flags = make([]Flags, result.WidthSubtiles*result.HeightSubtiles)
	var resolver ObjectResolver
	if len(resolvers) > 0 {
		resolver = resolvers[0]
	}
	for _, object := range stamp.Objects {
		decoded := resolveObject(result.Act, object.Type, object.ID, object.X, object.Y, object.Flags, resolver)
		result.Objects = append(result.Objects, decoded)
	}
	for tileY, row := range stamp.Tiles {
		for tileX, record := range row {
			for _, floor := range record.Floors {
				if !floor.Hidden && floor.Prop1 != 0 {
					result.apply(tileX, tileY, choose(lookup[tileKey{style: int32(floor.Style), sequence: int32(floor.Sequence)}], tileX, tileY))
				}
			}
			for _, wall := range record.Walls {
				if !wall.Hidden && wall.Prop1 != 0 {
					result.apply(tileX, tileY, choose(lookup[tileKey{kind: int32(wall.Type), style: int32(wall.Style), sequence: int32(wall.Sequence)}], tileX, tileY))
				}
			}
		}
	}
	return result, nil
}

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

func choose(tiles []*dt1.Tile, x, y int) *dt1.Tile {
	if len(tiles) == 0 {
		return nil
	}
	weight := 0
	for _, tile := range tiles {
		weight += int(tile.RarityFrameIndex)
	}
	if weight <= 0 {
		return tiles[0]
	}
	seed := uint64(x) * uint64(y)
	seed ^= seed << 13
	seed ^= seed >> 17
	seed ^= seed << 5
	random, sum := int(seed%uint64(weight)), 0
	for _, tile := range tiles {
		sum += int(tile.RarityFrameIndex)
		if sum >= random {
			return tile
		}
	}
	return tiles[0]
}

func (m *Map) apply(tileX, tileY int, tile *dt1.Tile) {
	if tile == nil {
		return
	}
	for index, source := range tile.SubTileFlags {
		x := tileX*SubtilesPerTile + index%SubtilesPerTile
		y := tileY*SubtilesPerTile + index/SubtilesPerTile
		if x < 0 || y < 0 || x >= m.WidthSubtiles || y >= m.HeightSubtiles {
			continue
		}
		target := &m.flags[y*m.WidthSubtiles+x]
		target.BlockWalk = target.BlockWalk || source.BlockWalk
		target.BlockLOS = target.BlockLOS || source.BlockLOS
		target.BlockJump = target.BlockJump || source.BlockJump
		target.BlockPlayerWalk = target.BlockPlayerWalk || source.BlockPlayerWalk
		target.BlockLight = target.BlockLight || source.BlockLight
	}
}

func (m *Map) FlagsAt(x, y int) (Flags, bool) {
	if x < 0 || y < 0 || x >= m.WidthSubtiles || y >= m.HeightSubtiles {
		return Flags{}, false
	}
	return m.flags[y*m.WidthSubtiles+x], true
}

// SubtileToPixel projects continuous DS1 subtile coordinates into the same
// image-space diamond centers used by TexturedDS1Preview.
func (m *Map) SubtileToPixel(x, y float64) (float64, float64) {
	originX := float64(m.HeightTiles*TilePixelWidth/2 + PreviewMargin)
	originY := float64(PreviewMargin + TilePixelHeight/2)
	return originX + (x-y)*TilePixelWidth/(2*SubtilesPerTile),
		originY + (x+y)*TilePixelHeight/(2*SubtilesPerTile)
}

// PixelToSubtile reverses SubtileToPixel. Fractional values are preserved so
// callers can choose their own collision sampling policy.
func (m *Map) PixelToSubtile(x, y float64) (float64, float64) {
	originX := float64(m.HeightTiles*TilePixelWidth/2 + PreviewMargin)
	originY := float64(PreviewMargin + TilePixelHeight/2)
	difference := (x - originX) * (2 * SubtilesPerTile) / TilePixelWidth
	sum := (y - originY) * (2 * SubtilesPerTile) / TilePixelHeight
	return (difference + sum) / 2, (sum - difference) / 2
}
