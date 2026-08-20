package assetinspect

import (
	"image"
	"image/draw"

	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
)

// drawWall selects a stable wall variant and composes a paired north corner
// when available. Missing candidates intentionally leave the placeholder visible.
func drawWall(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	wall ds1.WallRecord,
	x int,
	y int,
	origin image.Point,
) error {
	key := tileKey{
		tileType: int32(wall.Type),
		style:    int32(wall.Style),
		sequence: int32(wall.Sequence),
	}

	candidates := lookup[key]
	if len(candidates) == 0 {
		return nil
	}

	tile := selectTile(candidates, x, y, 0)

	if int32(wall.Type) == 3 {
		leftKey := tileKey{
			tileType: 4,
			style:    int32(wall.Style),
			sequence: int32(wall.Sequence),
		}
		if leftCandidates := lookup[leftKey]; len(leftCandidates) != 0 {
			left := selectTile(leftCandidates, x, y, 0)
			return drawNorthCorner(canvas, tile, left, origin)
		}
	}

	return drawTile(canvas, tile, origin, wallYAdjust(tile))
}

// drawNorthCorner renders both DT1 records in their established right-then-left
// order, preserving pixel overlap where the two north-corner halves meet.
func drawNorthCorner(canvas *image.RGBA, tile *dt1.Tile, left *dt1.Tile, origin image.Point) error {
	if err := drawTile(canvas, tile, origin, minimumBlockY(tile)+80); err != nil {
		return err
	}

	if err := drawTile(canvas, left, origin, minimumBlockY(left)+80); err != nil {
		return err
	}

	return nil
}

// drawMatchedTile selects one candidate by cell coordinates before drawing;
// absent matches are valid and retain the structural placeholder.
func drawMatchedTile(
	canvas *image.RGBA,
	candidates []*dt1.Tile,
	x int,
	y int,
	origin image.Point,
	yAdjust int,
) error {
	if len(candidates) == 0 {
		return nil
	}

	tile := selectTile(candidates, x, y, 0)

	return drawTile(canvas, tile, origin, yAdjust)
}

// drawMatchedTileWithAdjust delays offset calculation until after variant
// selection because shadows derive their baseline from the selected tile blocks.
func drawMatchedTileWithAdjust(
	canvas *image.RGBA,
	candidates []*dt1.Tile,
	x int,
	y int,
	origin image.Point,
	adjust func(*dt1.Tile) int,
) error {
	if len(candidates) == 0 {
		return nil
	}

	tile := selectTile(candidates, x, y, 0)

	return drawTile(canvas, tile, origin, adjust(tile))
}

// drawTile decodes a tile lazily and alpha-composites it at the DT1 horizontal
// anchor. A nil decoded image is a valid no-op rather than a render failure.
func drawTile(canvas *image.RGBA, tile *dt1.Tile, origin image.Point, yAdjust int) error {
	tileImage, err := tile.ImageE()
	if err != nil {
		return err
	}

	if tileImage == nil {
		return nil
	}

	point := image.Pt(origin.X-80, origin.Y+yAdjust)
	destination := tileImage.Bounds().Add(point)
	draw.Draw(canvas, destination, tileImage, tileImage.Bounds().Min, draw.Over)

	return nil
}

// wallYAdjust applies roof height to roofs and block minima to other walls so
// their decoded pixels meet the cell baseline expected by the isometric grid.
func wallYAdjust(tile *dt1.Tile) int {
	if tile.Type == 15 {
		return -int(tile.RoofHeight)
	}

	minimumY := 0
	for _, block := range tile.Blocks {
		if int(block.Y) < minimumY {
			minimumY = int(block.Y)
		}
	}

	return minimumY + 80
}

// shadowYAdjust places shadows on the same block baseline used by ordinary
// walls, while keeping the rule explicit at the shadow call site.
func shadowYAdjust(tile *dt1.Tile) int {
	return minimumBlockY(tile) + 80
}

// minimumBlockY finds the highest negative block offset, defaulting to the tile
// origin when every block begins at or below that baseline.
func minimumBlockY(tile *dt1.Tile) int {
	minimumY := 0
	for _, block := range tile.Blocks {
		if int(block.Y) < minimumY {
			minimumY = int(block.Y)
		}
	}

	return minimumY
}

// selectTile chooses a deterministic weighted variant from cell coordinates so
// repeated previews are stable without forcing every placement to use one frame.
func selectTile(tiles []*dt1.Tile, x, y int, seed uint64) *dt1.Tile {
	if len(tiles) == 1 {
		return tiles[0]
	}

	// Coordinate mixing intentionally remains local and deterministic; changing
	// this sequence would reshuffle tile variants throughout every rendered map.
	tileSeed := (seed + uint64(x)) * uint64(y)
	tileSeed ^= tileSeed << 13
	tileSeed ^= tileSeed >> 17
	tileSeed ^= tileSeed << 5

	weight := 0
	for _, tile := range tiles {
		weight += int(tile.RarityFrameIndex)
	}

	if weight <= 0 {
		return tiles[0]
	}

	random := int(tileSeed % uint64(weight))

	sum := 0
	for _, tile := range tiles {
		sum += int(tile.RarityFrameIndex)
		if sum >= random {
			return tile
		}
	}

	return tiles[0]
}

// isLowerWall identifies DT1 wall types composited in the ground pass, keeping
// their pass membership independent of DS1 record order.
func isLowerWall(tileType int32) bool {
	return tileType >= 16 && tileType <= 19
}

// isUpperWall identifies upright DT1 wall types that must wait for the second
// global pass; roofs remain excluded for their dedicated final pass.
func isUpperWall(tileType int32) bool {
	return (tileType >= 1 && tileType <= 9) || tileType == 12 || tileType == 14
}
