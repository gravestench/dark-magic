package assetinspect

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	dc6 "github.com/gravestench/dc6/pkg"
	"github.com/gravestench/dcc"
	"github.com/gravestench/ds1"
	"github.com/gravestench/dt1"
	"github.com/gravestench/pl2"
)

// Preview renders supported game assets as PNG diagnostics.
func Preview(source fs.FS, path string, direction, frame int) ([]byte, error) {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch extension {
	case "dc6":
		return DC6Preview(source, path, direction, frame)
	case "dcc":
		return DCCPreview(source, path, direction, frame)
	case "ds1":
		return DS1Preview(source, path)
	default:
		return nil, fmt.Errorf("PNG preview is not supported for %q assets", extension)
	}
}

// DCCPreview renders one decoded DCC animation frame as PNG.
func DCCPreview(source fs.FS, path string, direction, frame int) ([]byte, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DCC asset %q: %w", path, err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading DCC asset %q: %w", path, err)
	}
	asset, err := dcc.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DCC asset %q: %w", path, err)
	}
	directions := asset.Directions()
	if direction < 0 || direction >= len(directions) {
		return nil, fmt.Errorf("direction %d out of range [0,%d)", direction, len(directions))
	}
	frames := directions[direction].Frames()
	if frame < 0 || frame >= len(frames) {
		return nil, fmt.Errorf("frame %d out of range [0,%d)", frame, len(frames))
	}
	var output bytes.Buffer
	if err := png.Encode(&output, frames[frame]); err != nil {
		return nil, fmt.Errorf("encoding DCC preview: %w", err)
	}
	return output.Bytes(), nil
}

type tileKey struct {
	tileType int32
	style    int32
	sequence int32
}

// TexturedDS1Preview composes a DS1 stamp using matching DT1 tile graphics.
// Missing tile definitions retain a structural diamond so incomplete tileset
// lists remain obvious instead of producing an empty image.
func TexturedDS1Preview(source fs.FS, ds1Path string, dt1Paths []string, palettePath string) ([]byte, error) {
	stampFile, err := source.Open(ds1Path)
	if err != nil {
		return nil, fmt.Errorf("opening DS1 asset %q: %w", ds1Path, err)
	}
	stampData, err := io.ReadAll(stampFile)
	stampFile.Close()
	if err != nil {
		return nil, err
	}
	stamp, err := ds1.FromBytes(stampData)
	if err != nil {
		return nil, fmt.Errorf("decoding DS1 asset %q: %w", ds1Path, err)
	}

	var palette color.Palette
	if palettePath != "" {
		file, err := source.Open(palettePath)
		if err != nil {
			return nil, fmt.Errorf("opening PL2 asset %q: %w", palettePath, err)
		}
		data, readErr := io.ReadAll(file)
		file.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading PL2 asset %q: %w", palettePath, readErr)
		}
		paletteData, err := pl2.FromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("decoding PL2 asset %q: %w", palettePath, err)
		}
		palette = paletteData.BasePalette
	}

	lookup := make(map[tileKey][]*dt1.Tile)
	for _, path := range dt1Paths {
		file, err := source.Open(path)
		if err != nil {
			return nil, fmt.Errorf("opening DT1 asset %q: %w", path, err)
		}
		data, readErr := io.ReadAll(file)
		file.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading DT1 asset %q: %w", path, readErr)
		}
		tileset, err := dt1.FromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("decoding DT1 asset %q: %w", path, err)
		}
		if palette != nil {
			tileset.SetPalette(palette)
		}
		for _, tile := range tileset.Tiles {
			key := tileKey{tileType: tile.Type, style: tile.Style, sequence: tile.Sequence}
			lookup[key] = append(lookup[key], tile)
		}
	}

	const tileWidth, tileHeight, margin = 160, 80, 160
	width := int(stamp.Width+stamp.Height)*tileWidth/2 + margin*2
	height := int(stamp.Width+stamp.Height)*tileHeight/2 + margin*2
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(canvas, color.RGBA{R: 17, G: 18, B: 22, A: 255})
	originX := int(stamp.Height)*tileWidth/2 + margin

	// Diablo II renders the map in global passes. Interleaving upper walls with
	// floors produces convincing-looking but incorrect occlusion.
	for pass := 1; pass <= 3; pass++ {
		for y, row := range stamp.Tiles {
			for x, record := range row {
				origin := image.Pt(originX+(x-y)*tileWidth/2, margin+(x+y)*tileHeight/2)
				if pass == 1 {
					fillDiamond(canvas, origin.X, origin.Y+tileHeight/2, tileWidth, tileHeight, color.RGBA{R: 35, G: 39, B: 42, A: 255})
					for _, wall := range record.Walls {
						if !wall.Hidden && wall.Prop1 != 0 && isLowerWall(int32(wall.Type)) {
							drawWall(canvas, lookup, wall, x, y, origin)
						}
					}
					for _, floor := range record.Floors {
						if floor.Hidden || floor.Prop1 == 0 {
							continue
						}
						key := tileKey{tileType: 0, style: int32(floor.Style), sequence: int32(floor.Sequence)}
						drawMatchedTile(canvas, lookup[key], x, y, origin, 0)
					}
					for _, shadow := range record.Shadows {
						if shadow.Hidden || shadow.Prop1 == 0 {
							continue
						}
						key := tileKey{tileType: 13, style: int32(shadow.Style), sequence: int32(shadow.Sequence)}
						drawMatchedTileWithAdjust(canvas, lookup[key], x, y, origin, shadowYAdjust)
					}
				}
				if pass == 2 {
					for _, wall := range record.Walls {
						if !wall.Hidden && wall.Prop1 != 0 && isUpperWall(int32(wall.Type)) {
							drawWall(canvas, lookup, wall, x, y, origin)
						}
					}
				}
				if pass == 3 {
					for _, wall := range record.Walls {
						if !wall.Hidden && int32(wall.Type) == 15 {
							drawWall(canvas, lookup, wall, x, y, origin)
						}
					}
				}
			}
		}
	}

	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		return nil, fmt.Errorf("encoding textured DS1 preview: %w", err)
	}
	return output.Bytes(), nil
}

func drawWall(canvas *image.RGBA, lookup map[tileKey][]*dt1.Tile, wall ds1.WallRecord, x, y int, origin image.Point) {
	key := tileKey{tileType: int32(wall.Type), style: int32(wall.Style), sequence: int32(wall.Sequence)}
	candidates := lookup[key]
	if len(candidates) == 0 {
		return
	}
	tile := selectTile(candidates, x, y, 0)
	// North corners are split across type 3 and type 4 DT1 records. Align both
	// decoded images to the shared block baseline used by the game renderer.
	if int32(wall.Type) == 3 {
		leftKey := tileKey{tileType: 4, style: int32(wall.Style), sequence: int32(wall.Sequence)}
		if leftCandidates := lookup[leftKey]; len(leftCandidates) != 0 {
			left := selectTile(leftCandidates, x, y, 0)
			baseline := tile
			if left.Height < tile.Height {
				baseline = left
			}
			minimumY := minimumBlockY(baseline)
			drawTile(canvas, tile, origin, minimumY+80+minimumBlockY(tile)-minimumY)
			drawTile(canvas, left, origin, minimumY+80+minimumBlockY(left)-minimumY)
			return
		}
	}
	yAdjust := wallYAdjust(tile)
	drawTile(canvas, tile, origin, yAdjust)
}

func drawMatchedTile(canvas *image.RGBA, candidates []*dt1.Tile, x, y int, origin image.Point, yAdjust int) {
	if len(candidates) == 0 {
		return
	}
	tile := selectTile(candidates, x, y, 0)
	drawTile(canvas, tile, origin, yAdjust)
}

func drawMatchedTileWithAdjust(canvas *image.RGBA, candidates []*dt1.Tile, x, y int, origin image.Point, adjust func(*dt1.Tile) int) {
	if len(candidates) == 0 {
		return
	}
	tile := selectTile(candidates, x, y, 0)
	drawTile(canvas, tile, origin, adjust(tile))
}

func drawTile(canvas *image.RGBA, tile *dt1.Tile, origin image.Point, yAdjust int) {
	tileImage := tile.Image()
	if tileImage == nil {
		return
	}
	point := image.Pt(origin.X-80, origin.Y+yAdjust)
	destination := tileImage.Bounds().Add(point)
	draw.Draw(canvas, destination, tileImage, tileImage.Bounds().Min, draw.Over)
}

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

func shadowYAdjust(tile *dt1.Tile) int { return minimumBlockY(tile) + 80 }

func minimumBlockY(tile *dt1.Tile) int {
	minimumY := 0
	for _, block := range tile.Blocks {
		if int(block.Y) < minimumY {
			minimumY = int(block.Y)
		}
	}
	return minimumY
}

func selectTile(tiles []*dt1.Tile, x, y int, seed uint64) *dt1.Tile {
	if len(tiles) == 1 {
		return tiles[0]
	}
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

func isLowerWall(tileType int32) bool { return tileType >= 16 && tileType <= 19 }

func isUpperWall(tileType int32) bool {
	return (tileType >= 1 && tileType <= 9) || tileType == 12 || tileType == 14
}

// DC6Preview renders one DC6 frame as a PNG. The decoder's fallback palette
// is used when no game palette is supplied, making this useful for headless
// diagnostics even before the full renderer is initialized.
func DC6Preview(source fs.FS, path string, direction, frame int) ([]byte, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DC6 asset %q: %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading DC6 asset %q: %w", path, err)
	}
	asset, err := dc6.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DC6 asset %q: %w", path, err)
	}
	if direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("direction %d out of range [0,%d)", direction, len(asset.Directions))
	}
	frames := asset.Directions[direction].Frames
	if frame < 0 || frame >= len(frames) {
		return nil, fmt.Errorf("frame %d out of range [0,%d)", frame, len(frames))
	}

	var output bytes.Buffer
	if err := png.Encode(&output, frames[frame].ToImageRGBA()); err != nil {
		return nil, fmt.Errorf("encoding preview: %w", err)
	}
	return output.Bytes(), nil
}

// DS1Preview renders the structural floor/wall layout of a map stamp. It does
// not require DT1 textures, so malformed placement data can be diagnosed
// independently of palette and tileset configuration.
func DS1Preview(source fs.FS, path string) ([]byte, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DS1 asset %q: %w", path, err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading DS1 asset %q: %w", path, err)
	}
	stamp, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DS1 asset %q: %w", path, err)
	}

	const tileWidth, tileHeight, margin = 64, 32, 48
	width := (int(stamp.Width+stamp.Height) * tileWidth / 2) + margin*2
	height := (int(stamp.Width+stamp.Height) * tileHeight / 2) + margin*2
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(canvas, color.RGBA{R: 17, G: 18, B: 22, A: 255})
	originX := int(stamp.Height)*tileWidth/2 + margin
	for y, row := range stamp.Tiles {
		for x, tile := range row {
			centerX := originX + (x-y)*tileWidth/2
			centerY := margin + (x+y)*tileHeight/2
			shade := uint8(55)
			if len(tile.Floors) > 0 {
				shade += uint8((int(tile.Floors[0].Style) + int(tile.Floors[0].Sequence)) % 90)
			}
			fillDiamond(canvas, centerX, centerY, tileWidth, tileHeight, color.RGBA{R: shade, G: shade + 18, B: shade, A: 255})
			if len(tile.Walls) > 0 && tile.Walls[0].Prop1 != 0 {
				drawDiamond(canvas, centerX, centerY, tileWidth, tileHeight, color.RGBA{R: 205, G: 175, B: 95, A: 255})
			}
		}
	}
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		return nil, fmt.Errorf("encoding DS1 preview: %w", err)
	}
	return output.Bytes(), nil
}

func fill(img *image.RGBA, c color.Color) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func fillDiamond(img *image.RGBA, cx, cy, width, height int, c color.Color) {
	for dy := -height / 2; dy <= height/2; dy++ {
		half := (width / 2) * (height/2 - abs(dy)) / (height / 2)
		for x := cx - half; x <= cx+half; x++ {
			img.Set(x, cy+dy, c)
		}
	}
}

func drawDiamond(img *image.RGBA, cx, cy, width, height int, c color.Color) {
	for dy := -height / 2; dy <= height/2; dy++ {
		half := (width / 2) * (height/2 - abs(dy)) / (height / 2)
		img.Set(cx-half, cy+dy, c)
		img.Set(cx+half, cy+dy, c)
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
