// Package maprender adapts immutable game-world tile placements into sparse
// presentation images. It owns no native textures and does not alter authority.
package maprender

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dt1"
	"github.com/gravestench/pl2"
)

const DefaultChunkSize = 512

// Chunk is one sparse map-space texture. X/Y are top-left coordinates in the
// same full-stamp pixel space returned by world.Map.SubtileToPixel.
type Chunk struct {
	Column, Row   int
	X, Y          int
	Width, Height int
	Layer         world.TileLayer
	Depth         int
	Pixels        *image.RGBA
}

// Set describes the complete logical canvas without allocating that canvas.
type Set struct {
	Width, Height int
	ChunkSize     int
	Chunks        []Chunk
	// Objects stay semantic. Presentation may inspect them, but tile chunks do
	// not bake entities into pixels and therefore never steal simulation's job.
	Objects []world.Object

	draws [][]chunkDraw
}

type chunkDraw struct {
	tile        *dt1.Tile
	destination image.Rectangle
}

type tileSourceKey struct {
	path  string
	index int
}

type draftChunk struct {
	bounds image.Rectangle // full map-canvas coordinates
	pixels *image.RGBA
}

// Compose decodes only selected physical DT1 records and draws them into sparse
// fixed-size chunks. The caller can upload/cull chunks independently.
func Compose(source fs.FS, mapData *world.Map, palettePath string, chunkSize int) (*Set, error) {
	set, err := Index(source, mapData, palettePath, chunkSize)
	if err != nil {
		return nil, err
	}
	for index := range set.Chunks {
		chunk, materializeErr := set.Materialize(index)
		if materializeErr != nil {
			return nil, materializeErr
		}
		set.Chunks[index] = chunk
	}
	return set, nil
}

// Index prepares lightweight chunk geometry without rasterizing the complete
// map. It is the world-runtime path: the camera can inspect these rectangles,
// then call Materialize only for chunks entering its residency margin.
func Index(source fs.FS, mapData *world.Map, palettePath string, chunkSize int) (*Set, error) {
	if source == nil || mapData == nil {
		return nil, fmt.Errorf("maprender: source and map are required")
	}
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	palette, err := loadPalette(source, palettePath)
	if err != nil {
		return nil, fmt.Errorf("maprender: load palette %q: %w", palettePath, err)
	}
	decoded, err := decodeSelectedTiles(source, mapData.Tiles, palette)
	if err != nil {
		return nil, err
	}
	set := &Set{
		Width:     (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelWidth/2 + world.PreviewMargin*2,
		Height:    (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelHeight/2 + world.PreviewMargin*2,
		ChunkSize: chunkSize,
		Objects:   append([]world.Object(nil), mapData.Objects...),
	}
	type indexedChunk struct {
		bounds image.Rectangle
		draws  []chunkDraw
	}
	chunksByLayer := make(map[world.TileLayer]map[int]map[[2]int]*indexedChunk)
	for layer := world.LayerFloor; layer <= world.LayerRoof; layer++ {
		depths := make(map[int]map[[2]int]*indexedChunk)
		chunksByLayer[layer] = depths
		for _, placement := range mapData.Tiles {
			if placement.Layer != layer {
				continue
			}
			tile := decoded[tileSourceKey{path: placement.Reference.Path, index: placement.Reference.Index}]
			if tile == nil {
				continue
			}
			width, height := int(tile.Width), int(tile.Height)
			if height < 0 {
				height = -height
			}
			if width <= 0 || height <= 0 {
				continue
			}
			originX := mapData.HeightTiles*world.TilePixelWidth/2 + world.PreviewMargin + (placement.X-placement.Y)*world.TilePixelWidth/2
			originY := world.PreviewMargin + (placement.X+placement.Y)*world.TilePixelHeight/2
			destination := image.Rect(0, 0, width, height).Add(image.Pt(originX-world.TilePixelWidth/2, originY+tileYAdjust(tile, placement.Layer)))
			clipped := destination.Intersect(image.Rect(0, 0, set.Width, set.Height))
			if clipped.Empty() {
				continue
			}
			depth := world.TileDepth(layer, placement.X, placement.Y)
			chunks := depths[depth]
			if chunks == nil {
				chunks = make(map[[2]int]*indexedChunk)
				depths[depth] = chunks
			}
			firstColumn, lastColumn := clipped.Min.X/chunkSize, (clipped.Max.X-1)/chunkSize
			firstRow, lastRow := clipped.Min.Y/chunkSize, (clipped.Max.Y-1)/chunkSize
			for row := firstRow; row <= lastRow; row++ {
				for column := firstColumn; column <= lastColumn; column++ {
					cell := image.Rect(column*chunkSize, row*chunkSize, min((column+1)*chunkSize, set.Width), min((row+1)*chunkSize, set.Height))
					occupied := clipped.Intersect(cell)
					key := [2]int{column, row}
					indexed := chunks[key]
					if indexed == nil {
						indexed = &indexedChunk{bounds: occupied}
						chunks[key] = indexed
					} else {
						indexed.bounds = indexed.bounds.Union(occupied)
					}
					indexed.draws = append(indexed.draws, chunkDraw{tile: tile, destination: destination})
				}
			}
		}
	}
	for layer := world.LayerFloor; layer <= world.LayerRoof; layer++ {
		depths := make([]int, 0, len(chunksByLayer[layer]))
		for depth := range chunksByLayer[layer] {
			depths = append(depths, depth)
		}
		sort.Ints(depths)
		for _, depth := range depths {
			chunks := chunksByLayer[layer][depth]
			keys := make([][2]int, 0, len(chunks))
			for key := range chunks {
				keys = append(keys, key)
			}
			sort.Slice(keys, func(i, j int) bool {
				if keys[i][1] != keys[j][1] {
					return keys[i][1] < keys[j][1]
				}
				return keys[i][0] < keys[j][0]
			})
			for _, key := range keys {
				indexed := chunks[key]
				set.Chunks = append(set.Chunks, Chunk{Column: key[0], Row: key[1], X: indexed.bounds.Min.X, Y: indexed.bounds.Min.Y, Width: indexed.bounds.Dx(), Height: indexed.bounds.Dy(), Layer: layer, Depth: depth})
				set.draws = append(set.draws, indexed.draws)
			}
		}
	}
	return set, nil
}

// Materialize rasterizes one indexed depth chunk. Pixels are returned to the
// caller rather than retained by Set; the presentation cache therefore owns
// residency and can evict distant chunks independently.
func (set *Set) Materialize(index int) (Chunk, error) {
	if set == nil || index < 0 || index >= len(set.Chunks) || index >= len(set.draws) {
		return Chunk{}, fmt.Errorf("maprender: chunk %d out of range", index)
	}
	metadata := set.Chunks[index]
	key := [2]int{metadata.Column, metadata.Row}
	drafts := make(map[[2]int]*draftChunk)
	for _, operation := range set.draws[index] {
		pixels, err := operation.tile.ImageE()
		if err != nil {
			return Chunk{}, fmt.Errorf("maprender: decode visible tile graphics: %w", err)
		}
		if pixels != nil {
			drawIntoTightChunks(drafts, set.ChunkSize, set.Width, set.Height, operation.destination, pixels)
		}
	}
	pixels, origin, ok := trimTransparentMargins(drafts[key])
	if !ok {
		// Metadata is conservative because indexing intentionally avoids pixel
		// decoding. A DT1's rectangular bounds can overlap a cell while all of
		// its opaque pixels live next door. Keep the stable index with a harmless
		// transparent texel instead of turning camera admission into an error.
		pixels = image.NewRGBA(image.Rect(0, 0, 1, 1))
		origin = image.Pt(metadata.X, metadata.Y)
	}
	metadata.X, metadata.Y = origin.X, origin.Y
	metadata.Width, metadata.Height = pixels.Bounds().Dx(), pixels.Bounds().Dy()
	metadata.Pixels = pixels
	return metadata, nil
}

// trimTransparentMargins removes the empty padding carried by DT1 sprites.
// The returned origin keeps the smaller image in exactly the same map-space
// position, so this is invisible to callers while saving cache and GPU memory.
func trimTransparentMargins(draft *draftChunk) (*image.RGBA, image.Point, bool) {
	if draft == nil || draft.pixels == nil {
		return nil, image.Point{}, false
	}
	bounds := draft.pixels.Bounds()
	visible := image.Rectangle{Min: bounds.Max, Max: bounds.Min}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if draft.pixels.RGBAAt(x, y).A == 0 {
				continue
			}
			visible.Min.X = min(visible.Min.X, x)
			visible.Min.Y = min(visible.Min.Y, y)
			visible.Max.X = max(visible.Max.X, x+1)
			visible.Max.Y = max(visible.Max.Y, y+1)
		}
	}
	if visible.Empty() {
		return nil, image.Point{}, false
	}
	origin := draft.bounds.Min.Add(visible.Min.Sub(bounds.Min))
	if visible == bounds {
		return draft.pixels, origin, true
	}
	trimmed := image.NewRGBA(image.Rect(0, 0, visible.Dx(), visible.Dy()))
	draw.Draw(trimmed, trimmed.Bounds(), draft.pixels, visible.Min, draw.Src)
	return trimmed, origin, true
}

// drawIntoTightChunks keeps only the occupied rectangle inside each logical
// chunk/depth bucket. Outdoor depth ordering often creates one bucket per tile;
// allocating a full 512x512 RGBA for each one turned an 80x80 level into more
// than a gigabyte of temporary and cached pixels.
func drawIntoTightChunks(chunks map[[2]int]*draftChunk, chunkSize, width, height int, destination image.Rectangle, source image.Image) {
	original := destination
	destination = destination.Intersect(image.Rect(0, 0, width, height))
	if destination.Empty() {
		return
	}
	firstColumn, lastColumn := destination.Min.X/chunkSize, (destination.Max.X-1)/chunkSize
	firstRow, lastRow := destination.Min.Y/chunkSize, (destination.Max.Y-1)/chunkSize
	for row := firstRow; row <= lastRow; row++ {
		for column := firstColumn; column <= lastColumn; column++ {
			cell := image.Rect(column*chunkSize, row*chunkSize, min((column+1)*chunkSize, width), min((row+1)*chunkSize, height))
			clipped := destination.Intersect(cell)
			if clipped.Empty() {
				continue
			}
			key := [2]int{column, row}
			draft := chunks[key]
			if draft == nil {
				draft = &draftChunk{bounds: clipped, pixels: image.NewRGBA(image.Rect(0, 0, clipped.Dx(), clipped.Dy()))}
				chunks[key] = draft
			} else if union := draft.bounds.Union(clipped); union != draft.bounds {
				grown := image.NewRGBA(image.Rect(0, 0, union.Dx(), union.Dy()))
				draw.Draw(grown, draft.pixels.Bounds().Add(draft.bounds.Min.Sub(union.Min)), draft.pixels, image.Point{}, draw.Src)
				draft.bounds, draft.pixels = union, grown
			}
			target := clipped.Sub(draft.bounds.Min)
			sourcePoint := source.Bounds().Min.Add(clipped.Min.Sub(original.Min))
			draw.Draw(draft.pixels, target, source, sourcePoint, draw.Over)
		}
	}
}

func loadPalette(source fs.FS, path string) (color.Palette, error) {
	if strings.EqualFold(filepath.Ext(path), ".pl2") {
		file, err := source.Open(path)
		if err != nil {
			return nil, err
		}
		data, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		decoded, err := pl2.FromBytes(data)
		if err != nil {
			return nil, err
		}
		palette := append(color.Palette(nil), decoded.BasePalette...)
		if len(palette) > 0 {
			palette[0] = color.RGBA{}
		}
		return palette, nil
	}
	return assetdecode.Palette(source, path)
}

func decodeSelectedTiles(source fs.FS, placements []world.TilePlacement, palette color.Palette) (map[tileSourceKey]*dt1.Tile, error) {
	needed := make(map[string]map[int]struct{})
	for _, placement := range placements {
		if placement.Reference.Path == "" {
			continue
		} // semantic-only hidden cell
		indexes := needed[placement.Reference.Path]
		if indexes == nil {
			indexes = make(map[int]struct{})
			needed[placement.Reference.Path] = indexes
		}
		indexes[placement.Reference.Index] = struct{}{}
	}
	result := make(map[tileSourceKey]*dt1.Tile)
	for path, indexes := range needed {
		file, err := source.Open(path)
		if err != nil {
			return nil, fmt.Errorf("maprender: open %q: %w", path, err)
		}
		opened, err := openDT1(file)
		if err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("maprender: index %q: %w", path, err)
		}
		opened.SetPalette(palette)
		for index := range indexes {
			tile, decodeErr := opened.DecodeTile(index)
			if decodeErr != nil {
				_ = file.Close()
				return nil, fmt.Errorf("maprender: decode %q tile %d: %w", path, index, decodeErr)
			}
			result[tileSourceKey{path: path, index: index}] = tile
		}
		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("maprender: close %q: %w", path, closeErr)
		}
	}
	return result, nil
}

func openDT1(file fs.File) (*dt1.File, error) {
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

func drawIntoChunks(chunks map[[2]int]*image.RGBA, chunkSize, width, height int, destination image.Rectangle, source image.Image) {
	originalDestination := destination
	destination = destination.Intersect(image.Rect(0, 0, width, height))
	if destination.Empty() {
		return
	}
	firstColumn, lastColumn := destination.Min.X/chunkSize, (destination.Max.X-1)/chunkSize
	firstRow, lastRow := destination.Min.Y/chunkSize, (destination.Max.Y-1)/chunkSize
	for row := firstRow; row <= lastRow; row++ {
		for column := firstColumn; column <= lastColumn; column++ {
			key := [2]int{column, row}
			pixels := chunks[key]
			if pixels == nil {
				pixels = image.NewRGBA(image.Rect(0, 0, min(chunkSize, width-column*chunkSize), min(chunkSize, height-row*chunkSize)))
				chunks[key] = pixels
			}
			chunkBounds := pixels.Bounds().Add(image.Pt(column*chunkSize, row*chunkSize))
			clipped := destination.Intersect(chunkBounds)
			local := clipped.Sub(image.Pt(column*chunkSize, row*chunkSize))
			sourcePoint := source.Bounds().Min.Add(clipped.Min.Sub(originalDestination.Min))
			draw.Draw(pixels, local, source, sourcePoint, draw.Over)
		}
	}
}

func tileYAdjust(tile *dt1.Tile, layer world.TileLayer) int {
	if layer == world.LayerRoof {
		return -int(tile.RoofHeight)
	}
	minimumY := 0
	for _, block := range tile.Blocks {
		if int(block.Y) < minimumY {
			minimumY = int(block.Y)
		}
	}
	if layer == world.LayerFloor {
		return 0
	}
	return minimumY + world.TilePixelHeight
}
