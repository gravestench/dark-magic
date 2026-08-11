package maprender

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"sort"
	"sync"

	"github.com/gravestench/dark-magic/internal/game/world"
)

const DefaultTileBucketSize = 512

// TileGraphic is one decoded physical DT1 record. A map may place this picture
// thousands of times, but it should only occupy CPU and GPU memory once.
type TileGraphic struct {
	Path          string
	Index         int
	Width, Height int
	once          sync.Once
	pixels        image.Image
	err           error
}

// TileDraw is the tiny, repeatable half of map presentation. Graphic indexes a
// shared picture; Bounds says where that occurrence belongs in map pixels.
type TileDraw struct {
	Graphic int
	Bounds  image.Rectangle
	Layer   world.TileLayer
	Depth   int
}

// TileBucket is one deterministic cell in the placement spatial index. Draws
// contains indexes into TileSet.Draws, never copies of graphics or placements.
type TileBucket struct {
	Column, Row int
	Draws       []int
}

// TileSet separates immutable pictures from their placement commands. This is
// deliberately renderer-neutral: a backend may use ordinary retained nodes,
// instancing, or a future tile batch without changing authoritative map data.
type TileSet struct {
	Width, Height int
	Graphics      []*TileGraphic
	Draws         []TileDraw
	BucketSize    int
	Buckets       []TileBucket
	source        fs.FS
	palette       color.Palette
	buckets       map[[2]int][]int
}

// Place builds shared DT1 graphics and lightweight draw commands. Unlike the
// chunk compositor, it never copies a tile's pixels into one RGBA per depth
// bucket. Transparent pixels remain in the one shared source graphic.
func Place(source fs.FS, mapData *world.Map, palettePath string) (*TileSet, error) {
	if source == nil || mapData == nil {
		return nil, fmt.Errorf("maprender: source and map are required")
	}
	palette, err := loadPalette(source, palettePath)
	if err != nil {
		return nil, fmt.Errorf("maprender: load palette %q: %w", palettePath, err)
	}
	result := &TileSet{
		Width:      (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelWidth/2 + world.PreviewMargin*2,
		Height:     (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelHeight/2 + world.PreviewMargin*2,
		BucketSize: DefaultTileBucketSize, source: source, palette: palette,
	}
	graphicIndexes := make(map[tileSourceKey]int)
	canvas := image.Rect(0, 0, result.Width, result.Height)
	for _, placement := range mapData.Tiles {
		key := tileSourceKey{path: placement.Reference.Path, index: placement.Reference.Index}
		if key.path == "" {
			continue
		}
		graphic, exists := graphicIndexes[key]
		if !exists {
			graphic = len(result.Graphics)
			graphicIndexes[key] = graphic
			result.Graphics = append(result.Graphics, &TileGraphic{
				Path: key.path, Index: key.index, Width: max(1, int(placement.Reference.Width)),
				Height: max(1, absolute(int(placement.Reference.Height))),
			})
		}
		picture := result.Graphics[graphic]
		originX := mapData.HeightTiles*world.TilePixelWidth/2 + world.PreviewMargin + (placement.X-placement.Y)*world.TilePixelWidth/2
		originY := world.PreviewMargin + (placement.X+placement.Y)*world.TilePixelHeight/2
		bounds := image.Rect(0, 0, picture.Width, picture.Height).Add(image.Pt(
			originX-world.TilePixelWidth/2,
			originY+referenceYAdjust(placement.Reference, placement.Layer),
		))
		if bounds.Intersect(canvas).Empty() {
			continue
		}
		result.Draws = append(result.Draws, TileDraw{
			Graphic: graphic,
			Bounds:  bounds,
			Layer:   placement.Layer,
			Depth:   world.TileDepth(placement.Layer, placement.X, placement.Y),
		})
	}
	result.buildBuckets()
	return result, nil
}

func (set *TileSet) buildBuckets() {
	if set == nil {
		return
	}
	if set.BucketSize <= 0 {
		set.BucketSize = DefaultTileBucketSize
	}
	set.buckets = make(map[[2]int][]int)
	for index, draw := range set.Draws {
		if draw.Bounds.Empty() {
			continue
		}
		firstColumn, lastColumn := draw.Bounds.Min.X/set.BucketSize, (draw.Bounds.Max.X-1)/set.BucketSize
		firstRow, lastRow := draw.Bounds.Min.Y/set.BucketSize, (draw.Bounds.Max.Y-1)/set.BucketSize
		for row := firstRow; row <= lastRow; row++ {
			for column := firstColumn; column <= lastColumn; column++ {
				key := [2]int{column, row}
				set.buckets[key] = append(set.buckets[key], index)
			}
		}
	}
	keys := make([][2]int, 0, len(set.buckets))
	for key := range set.buckets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][1] != keys[j][1] {
			return keys[i][1] < keys[j][1]
		}
		return keys[i][0] < keys[j][0]
	})
	set.Buckets = make([]TileBucket, 0, len(keys))
	for _, key := range keys {
		set.Buckets = append(set.Buckets, TileBucket{
			Column: key[0], Row: key[1], Draws: append([]int(nil), set.buckets[key]...),
		})
	}
}

// MaterializeGraphic expands one unique DT1 picture at most once. It is safe
// for multiple viewport preload workers to request the same picture together.
func (set *TileSet) MaterializeGraphic(index int) (image.Image, error) {
	if set == nil || index < 0 || index >= len(set.Graphics) {
		return nil, fmt.Errorf("maprender: tile graphic %d out of range", index)
	}
	graphic := set.Graphics[index]
	graphic.once.Do(func() {
		file, openErr := set.source.Open(graphic.Path)
		if openErr != nil {
			graphic.err = openErr
			return
		}
		opened, openErr := openDT1(file)
		if openErr != nil {
			_ = file.Close()
			graphic.err = openErr
			return
		}
		opened.SetPalette(set.palette)
		tile, decodeErr := opened.DecodeTile(graphic.Index)
		if decodeErr != nil {
			_ = file.Close()
			graphic.err = decodeErr
			return
		}
		graphic.pixels, graphic.err = tile.ImageE()
		closeErr := file.Close()
		if graphic.err == nil && closeErr != nil {
			graphic.err = closeErr
		}
		if graphic.err != nil {
			graphic.err = fmt.Errorf("maprender: decode %q tile %d graphics: %w", graphic.Path, graphic.Index, graphic.err)
			return
		}
		if graphic.pixels == nil {
			graphic.pixels = image.NewRGBA(image.Rect(0, 0, graphic.Width, graphic.Height))
		}
	})
	return graphic.pixels, graphic.err
}

func referenceYAdjust(reference world.TileReference, layer world.TileLayer) int {
	if layer == world.LayerRoof {
		return -int(reference.RoofHeight)
	}
	if layer == world.LayerFloor {
		return 0
	}
	// OpenDiablo2 derives wall/object placement from the minimum block-header Y,
	// not the DT1 tile header Height. TileCatalog obtains this compact metadata
	// without decoding or reading pixel payloads.
	return int(reference.YAdjust)
}

// Visible returns draw indexes intersecting a map-pixel viewport. Callers keep
// ownership of residency policy; this method only performs deterministic culling.
func (set *TileSet) Visible(view image.Rectangle, destination []int) []int {
	if set == nil || view.Empty() {
		return destination
	}
	if set.buckets == nil {
		set.buildBuckets()
	}
	start := len(destination)
	seen := make(map[int]struct{})
	firstColumn, lastColumn := view.Min.X/set.BucketSize, (view.Max.X-1)/set.BucketSize
	firstRow, lastRow := view.Min.Y/set.BucketSize, (view.Max.Y-1)/set.BucketSize
	for row := firstRow; row <= lastRow; row++ {
		for column := firstColumn; column <= lastColumn; column++ {
			for _, index := range set.buckets[[2]int{column, row}] {
				if _, duplicate := seen[index]; duplicate || set.Draws[index].Bounds.Intersect(view).Empty() {
					continue
				}
				seen[index] = struct{}{}
				destination = append(destination, index)
			}
		}
	}
	sort.Ints(destination[start:])
	return destination
}

func absolute(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
