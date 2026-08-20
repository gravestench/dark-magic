package world

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/gravestench/dt1"
)

// TileIdentity is the logical key stored by DS1 cells and DT1 records.
// Historical tools call these fields type/style/sequence. Orientation,
// main-index, and sub-index describe their actual lookup role more clearly.
type TileIdentity struct {
	Orientation int32
	MainIndex   int32
	SubIndex    int32
}

// TileReference retains renderer-independent DT1 metadata and enough source
// identity for a presentation adapter to decode graphics lazily later.
type TileReference struct {
	Identity      TileIdentity
	Path          string
	Index         int
	Direction     int32
	RoofHeight    int16
	MaterialFlags dt1.MaterialFlags
	Width         int32
	Height        int32
	YAdjust       int32
	Rarity        int32
	SubTileFlags  [25]dt1.SubTileFlags
}

// TileCatalog groups every physical DT1 record by the logical key requested by
// a DS1 cell. A key can have several rarity-weighted graphical alternatives.
type TileCatalog struct {
	entries map[TileIdentity][]TileReference
}

// NewTileCatalog constructs an immutable lookup from already-decoded metadata.
// It is public within the internal package so generators and synthetic tests
// use exactly the same selection contract as file-backed maps.
func NewTileCatalog(references []TileReference) *TileCatalog {
	catalog := &TileCatalog{entries: make(map[TileIdentity][]TileReference)}
	for _, reference := range references {
		catalog.entries[reference.Identity] = append(catalog.entries[reference.Identity], reference)
	}

	return catalog
}

// LoadTileCatalog indexes DT1 headers only. Pixel blocks remain compressed and
// unopened until a presentation consumer requests a selected TileReference.
func LoadTileCatalog(source fs.FS, paths []string) (*TileCatalog, error) {
	references := make([]TileReference, 0)

	for _, path := range paths {
		file, err := source.Open(path)
		if err != nil {
			return nil, fmt.Errorf("world: open %q: %w", path, err)
		}

		opened, err := openDT1Metadata(file)
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

			minimumBlockY, _, boundsErr := opened.TileBlockYBounds(index)
			if boundsErr != nil {
				_ = file.Close()
				return nil, fmt.Errorf("world: read DT1 %q tile %d block bounds: %w", path, index, boundsErr)
			}

			references = append(references, TileReference{
				Identity: TileIdentity{Orientation: tile.Type, MainIndex: tile.Style, SubIndex: tile.Sequence},
				Path:     path, Index: index, Direction: tile.Direction, RoofHeight: tile.RoofHeight,
				MaterialFlags: tile.MaterialFlags, Width: tile.Width, Height: tile.Height,
				YAdjust: int32(minimumBlockY) + TilePixelHeight,
				Rarity:  tile.RarityFrameIndex, SubTileFlags: tile.SubTileFlags,
			})
		}

		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("world: close DT1 %q: %w", path, closeErr)
		}
	}

	return NewTileCatalog(references), nil
}

// openDT1Metadata prefers random access so DT1 pixel blocks remain lazy. Filesystems without ReaderAt support fall back
// to an owned byte slice while preserving identical decoded metadata.
func openDT1Metadata(file fs.File) (*dt1.File, error) {
	if reader, ok := file.(io.ReaderAt); ok {
		if info, err := file.Stat(); err == nil {
			return dt1.Open(reader, info.Size())
		}
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return dt1.OpenBytes(data)
}

// Candidates returns a defensive copy so callers cannot mutate catalog order.
func (c *TileCatalog) Candidates(identity TileIdentity) []TileReference {
	if c == nil {
		return nil
	}

	return append([]TileReference(nil), c.entries[identity]...)
}

// Select deterministically chooses one physical record using positive rarity
// as its weight. Zero-weight records are never selected while a positive choice
// exists; an all-zero group falls back to its first authored record.
func (c *TileCatalog) Select(identity TileIdentity, x, y int, seed uint64) (TileReference, bool) {
	if c == nil {
		return TileReference{}, false
	}

	candidates := c.entries[identity]
	if len(candidates) == 0 {
		return TileReference{}, false
	}

	weight := uint64(0)

	for _, candidate := range candidates {
		if candidate.Rarity > 0 {
			weight += uint64(candidate.Rarity)
		}
	}

	if weight == 0 {
		return candidates[0], true
	}

	random := tileRandom(seed, identity, x, y) % weight
	cumulative := uint64(0)

	for _, candidate := range candidates {
		if candidate.Rarity <= 0 {
			continue
		}

		cumulative += uint64(candidate.Rarity)
		if random < cumulative {
			return candidate, true
		}
	}

	return candidates[len(candidates)-1], true
}

// tileRandom mixes seed, full tile identity, and both coordinates through SplitMix64. The pure function makes rarity
// selection independent from traversal order and safe to reproduce in generated-world assembly.
func tileRandom(seed uint64, identity TileIdentity, x, y int) uint64 {
	// SplitMix64's avalanche keeps adjacent map coordinates independent, unlike
	// multiplying x*y (which collapses every coordinate on either zero axis).
	value := seed ^ uint64(uint32(x))*0x9e3779b1 ^ uint64(uint32(y))*0x85ebca77
	value ^= uint64(uint32(identity.Orientation)) << 32
	value ^= uint64(uint32(identity.MainIndex))*0xc2b2ae3d ^ uint64(uint32(identity.SubIndex))*0x27d4eb2f
	value += 0x9e3779b97f4a7c15
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb

	return value ^ (value >> 31)
}
