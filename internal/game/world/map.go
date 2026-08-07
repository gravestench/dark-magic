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

const SubtilesPerTile = 5

const (
	TilePixelWidth  = 160
	TilePixelHeight = 80
	PreviewMargin   = 160
)

const (
	ObjectTypeDynamic int32 = 1
	ObjectTypeStatic  int32 = 2
)

type Flags struct {
	BlockWalk, BlockLOS, BlockJump, BlockPlayerWalk, BlockLight bool
}

// Blocked reports whether a player-sized point cannot walk through this
// subtile. BlockWalk is shared terrain collision; BlockPlayerWalk is the
// additional player-specific restriction encoded by DT1.
func (f Flags) Blocked() bool { return f.BlockWalk || f.BlockPlayerWalk }

type Object struct {
	Type, ID, X, Y, Flags int32
	ObjectID              int
	Class                 string
	Description           string
	Resolved              bool
}

type ObjectResolver interface {
	ResolveStaticObject(act, id int) (objectID int, description string, found bool)
	ResolveDynamicObject(act, id int) (class string, found bool)
}

type Map struct {
	WidthTiles, HeightTiles       int
	WidthSubtiles, HeightSubtiles int
	Act                           int
	Objects                       []Object
	flags                         []Flags
}

type tileKey struct{ kind, style, sequence int32 }

func Load(source fs.FS, stampPath string, tilePaths []string, resolvers ...ObjectResolver) (*Map, error) {
	data, err := read(source, stampPath)
	if err != nil {
		return nil, err
	}
	stamp, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("world: decode DS1 %q: %w", stampPath, err)
	}
	lookup := make(map[tileKey][]*dt1.Tile)
	for _, path := range tilePaths {
		data, err := read(source, path)
		if err != nil {
			return nil, err
		}
		tiles, err := dt1.FromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("world: decode DT1 %q: %w", path, err)
		}
		for _, tile := range tiles.Tiles {
			key := tileKey{kind: tile.Type, style: tile.Style, sequence: tile.Sequence}
			lookup[key] = append(lookup[key], tile)
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

func read(source fs.FS, name string) ([]byte, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("world: open %q: %w", name, err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("world: read %q: %w", name, err)
	}
	return data, nil
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
