package assetinspect

import (
	"fmt"
	"image"
	"image/color"

	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
)

const (
	texturedTileWidth  = 160
	texturedTileHeight = 80
	texturedMargin     = 160
)

// renderTexturedDS1 allocates the map canvas and executes global rendering
// passes in game order so upper walls and roofs occlude earlier layers correctly.
func renderTexturedDS1(stamp *ds1.DS1, lookup map[tileKey][]*dt1.Tile) (image.Image, error) {
	width := int(stamp.Width+stamp.Height)*texturedTileWidth/2 + texturedMargin*2
	height := int(stamp.Width+stamp.Height)*texturedTileHeight/2 + texturedMargin*2
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(canvas, color.RGBA{R: 17, G: 18, B: 22, A: 255})

	originX := int(stamp.Height)*texturedTileWidth/2 + texturedMargin
	// Diablo II renders the entire map in global passes. Interleaving upper
	// walls with floors produces convincing-looking but incorrect occlusion.
	for pass := 1; pass <= 3; pass++ {
		if err := renderTexturedDS1Pass(canvas, stamp, lookup, originX, pass); err != nil {
			return nil, err
		}
	}

	return canvas, nil
}

// renderTexturedDS1Pass visits cells in source order for one global layer,
// retaining deterministic overlap between neighboring tile images.
func renderTexturedDS1Pass(
	canvas *image.RGBA,
	stamp *ds1.DS1,
	lookup map[tileKey][]*dt1.Tile,
	originX int,
	pass int,
) error {
	for y, row := range stamp.Tiles {
		for x, record := range row {
			origin := image.Pt(
				originX+(x-y)*texturedTileWidth/2,
				texturedMargin+(x+y)*texturedTileHeight/2,
			)
			if err := renderTexturedDS1Cell(canvas, lookup, record, x, y, origin, pass); err != nil {
				return err
			}
		}
	}

	return nil
}

// renderTexturedDS1Cell keeps pass membership explicit, preventing a later
// layer addition from accidentally changing the renderer's global ordering.
func renderTexturedDS1Cell(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
	pass int,
) error {
	switch pass {
	case 1:
		return drawDS1GroundLayers(canvas, lookup, record, x, y, origin)
	case 2:
		return drawUpperDS1Walls(canvas, lookup, record, x, y, origin)
	case 3:
		return drawDS1Roofs(canvas, lookup, record, x, y, origin)
	default:
		return nil
	}
}

// drawDS1GroundLayers establishes the structural placeholder before lower walls,
// floors, and shadows are composited in their compatibility-sensitive order.
func drawDS1GroundLayers(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
) error {
	fillDiamond(
		canvas,
		origin.X,
		origin.Y+texturedTileHeight/2,
		texturedTileWidth,
		texturedTileHeight,
		color.RGBA{R: 35, G: 39, B: 42, A: 255},
	)

	if err := drawLowerDS1Walls(canvas, lookup, record, x, y, origin); err != nil {
		return err
	}

	if err := drawDS1Floors(canvas, lookup, record, x, y, origin); err != nil {
		return err
	}

	return drawDS1Shadows(canvas, lookup, record, x, y, origin)
}

// drawLowerDS1Walls draws only wall types that belong beneath floor and shadow
// layers; hidden or empty records must not affect the preview.
func drawLowerDS1Walls(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
) error {
	for _, wall := range record.Walls {
		if wall.Hidden || wall.Prop1 == 0 || !isLowerWall(int32(wall.Type)) {
			continue
		}

		if err := drawWall(canvas, lookup, wall, x, y, origin); err != nil {
			return fmt.Errorf("draw lower wall at %d,%d: %w", x, y, err)
		}
	}

	return nil
}

// drawDS1Floors composites visible floor graphics after lower walls, matching
// the established map pass even when the order appears counterintuitive.
func drawDS1Floors(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
) error {
	for _, floor := range record.Floors {
		if floor.Hidden || floor.Prop1 == 0 {
			continue
		}

		key := tileKey{
			tileType: 0,
			style:    int32(floor.Style),
			sequence: int32(floor.Sequence),
		}
		if err := drawMatchedTile(canvas, lookup[key], x, y, origin, 0); err != nil {
			return fmt.Errorf("draw floor at %d,%d: %w", x, y, err)
		}
	}

	return nil
}

// drawDS1Shadows applies tile-specific vertical adjustment so shadow blocks
// align with the same baseline as their corresponding ground geometry.
func drawDS1Shadows(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
) error {
	for _, shadow := range record.Shadows {
		if shadow.Hidden || shadow.Prop1 == 0 {
			continue
		}

		key := tileKey{
			tileType: 13,
			style:    int32(shadow.Style),
			sequence: int32(shadow.Sequence),
		}
		if err := drawMatchedTileWithAdjust(canvas, lookup[key], x, y, origin, shadowYAdjust); err != nil {
			return fmt.Errorf("draw shadow at %d,%d: %w", x, y, err)
		}
	}

	return nil
}

// drawUpperDS1Walls defers upright wall types to the second global pass so they
// occlude all ground-layer pixels from neighboring cells consistently.
func drawUpperDS1Walls(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
) error {
	for _, wall := range record.Walls {
		if wall.Hidden || wall.Prop1 == 0 || !isUpperWall(int32(wall.Type)) {
			continue
		}

		if err := drawWall(canvas, lookup, wall, x, y, origin); err != nil {
			return fmt.Errorf("draw upper wall at %d,%d: %w", x, y, err)
		}
	}

	return nil
}

// drawDS1Roofs reserves roof type 15 for the final pass. Unlike other wall
// layers, existing behavior treats a non-hidden roof as drawable regardless of Prop1.
func drawDS1Roofs(
	canvas *image.RGBA,
	lookup map[tileKey][]*dt1.Tile,
	record ds1.TileRecord,
	x int,
	y int,
	origin image.Point,
) error {
	for _, wall := range record.Walls {
		if wall.Hidden || int32(wall.Type) != 15 {
			continue
		}

		if err := drawWall(canvas, lookup, wall, x, y, origin); err != nil {
			return fmt.Errorf("draw roof at %d,%d: %w", x, y, err)
		}
	}

	return nil
}
