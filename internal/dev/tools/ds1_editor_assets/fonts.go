package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"slices"

	"github.com/gravestench/dc6"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/gosmallcaps"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	fontTableHeaderSize = 12
	fontGlyphRecordSize = 14
)

type editorFontSpec struct {
	Name          string
	Source        []byte
	PixelSize     float64
	LineHeight    int
	OutlineRadius int
	LetterSpacing int
	FlatPixels    bool
	PixelArt      bool
	CellWidth     int
	CellHeight    int
	SourceTop     int
}

type fontManifest struct {
	Schema  string              `json:"schema"`
	Version int                 `json:"version"`
	Fonts   []fontManifestEntry `json:"fonts"`
}

type fontManifestEntry struct {
	Name       string `json:"name"`
	Table      string `json:"table"`
	Sheet      string `json:"sheet"`
	Palette    string `json:"palette"`
	LineHeight int    `json:"line_height"`
	Glyphs     int    `json:"glyphs"`
}

type rasterizedGlyph struct {
	Code    rune
	Advance int
	OffsetX int
	Pixels  []byte
	Width   int
	Height  int
}

var editorFontSpecs = []editorFontSpec{
	{Name: "large", Source: gosmallcaps.TTF, PixelSize: 24, LineHeight: 32, OutlineRadius: 1, LetterSpacing: 1},
	{Name: "medium", Source: gomedium.TTF, PixelSize: 16, LineHeight: 23, OutlineRadius: 1},
	{Name: "small", LineHeight: 16, PixelArt: true, CellWidth: 7, CellHeight: 13},
	{Name: "very_small", LineHeight: 12, PixelArt: true, CellWidth: 7, CellHeight: 12, SourceTop: 1},
}

var editorFontRunes = fontRunes()

// buildFonts creates every editor-owned font and a human-readable manifest in one pass.
// Sharing one palette and glyph order makes the four sizes deterministic and cheap to preload.
func buildFonts(outputRoot string, palette color.Palette) error {
	fontRoot := filepath.Join(outputRoot, "fonts")
	if err := os.MkdirAll(fontRoot, 0o755); err != nil {
		return err
	}

	manifest := fontManifest{Schema: "ds1editor.fonts/v1", Version: 1}
	for _, spec := range editorFontSpecs {
		if err := buildFont(fontRoot, palette, spec); err != nil {
			return fmt.Errorf("build %s editor font: %w", spec.Name, err)
		}

		manifest.Fonts = append(manifest.Fonts, fontManifestEntry{
			Name:       spec.Name,
			Table:      fontVFSPath(spec.Name, ".tbl"),
			Sheet:      fontVFSPath(spec.Name, ".dc6"),
			Palette:    "darkmagic/ds1-editor/ui/palette.dat",
			LineHeight: spec.LineHeight,
			Glyphs:     len(editorFontRunes),
		})
	}

	return writeJSON(filepath.Join(fontRoot, "manifest.json"), manifest)
}

// buildFont rasterizes one size, then writes matching Woo! metrics and DC6 frames.
// Keeping metrics and pixels in one transaction prevents frame indices from drifting.
func buildFont(fontRoot string, palette color.Palette, spec editorFontSpec) error {
	if spec.PixelArt {
		glyphs := make([]rasterizedGlyph, 0, len(editorFontRunes))
		for _, code := range editorFontRunes {
			glyphs = append(glyphs, rasterizePixelGlyph(palette, spec, code))
		}
		if err := writeFontTable(filepath.Join(fontRoot, spec.Name+".tbl"), glyphs); err != nil {
			return err
		}
		return writeFontSheet(filepath.Join(fontRoot, spec.Name+".dc6"), glyphs)
	}

	parsed, err := opentype.Parse(spec.Source)
	if err != nil {
		return fmt.Errorf("parse source: %w", err)
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    spec.PixelSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("create face: %w", err)
	}
	if closer, ok := face.(interface{ Close() error }); ok {
		defer closer.Close()
	}

	glyphs := make([]rasterizedGlyph, 0, len(editorFontRunes))
	for _, code := range editorFontRunes {
		glyphs = append(glyphs, rasterizeGlyph(face, palette, spec, code))
	}

	if err := writeFontTable(filepath.Join(fontRoot, spec.Name+".tbl"), glyphs); err != nil {
		return err
	}

	return writeFontSheet(filepath.Join(fontRoot, spec.Name+".dc6"), glyphs)
}

// rasterizePixelGlyph starts from Go's fixed 7x13 bitmap face, never an
// antialiased outline. Small keeps that authored grid exactly; very-small
// reduces it with binary cell occupancy and therefore cannot acquire fringe colors.
func rasterizePixelGlyph(palette color.Palette, spec editorFontSpec, code rune) rasterizedGlyph {
	const sourceWidth, sourceHeight = 7, 13
	source := image.NewAlpha(image.Rect(0, 0, sourceWidth, sourceHeight))
	drawer := font.Drawer{
		Dst:  source,
		Src:  image.White,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0, basicfont.Face7x13.Metrics().Ascent.Ceil()),
	}
	drawer.DrawString(string(code))

	fill := uint8(palette.Index(color.NRGBA{R: 255, G: 238, B: 187, A: 255}))
	pixels := make([]byte, spec.CellWidth*spec.LineHeight)
	top := max(0, (spec.LineHeight-spec.CellHeight)/2)
	for y := 0; y < spec.CellHeight; y++ {
		for x := 0; x < spec.CellWidth; x++ {
			if bitmapCellOccupied(source, x, y, spec.CellWidth, spec.CellHeight, spec.SourceTop) {
				pixels[(top+y)*spec.CellWidth+x] = fill
			}
		}
	}

	advance := spec.CellWidth
	if code == ' ' {
		advance = max(3, spec.CellWidth-2)
	}
	return rasterizedGlyph{
		Code: code, Advance: advance, Pixels: pixels,
		Width: spec.CellWidth, Height: spec.LineHeight,
	}
}

// bitmapCellOccupied uses exact source pixels for the small face. Very-small
// removes only the X11 face's unused top row; it never scales or merges strokes.
func bitmapCellOccupied(source *image.Alpha, x, y, width, height, sourceTop int) bool {
	sourceBounds := source.Bounds()
	sourceX := sourceBounds.Min.X + min(sourceBounds.Dx()-1, (2*x+1)*sourceBounds.Dx()/(2*width))
	availableHeight := sourceBounds.Dy() - sourceTop
	sourceY := sourceBounds.Min.Y + sourceTop + min(availableHeight-1, (2*y+1)*availableHeight/(2*height))
	return source.AlphaAt(sourceX, sourceY).A > 0
}

// rasterizeGlyph converts an outline glyph into a crisp, shaded indexed bitmap.
// Full line-height frames preserve baseline alignment without format-specific vertical offsets.
func rasterizeGlyph(
	face font.Face,
	palette color.Palette,
	spec editorFontSpec,
	code rune,
) rasterizedGlyph {
	text := string(code)
	bounds, advance := font.BoundString(face, text)
	left := min(0, fixedFloor(bounds.Min.X)-spec.OutlineRadius)
	right := max(fixedCeil(advance), fixedCeil(bounds.Max.X)+spec.OutlineRadius)
	width := max(1, right-left)

	mask := image.NewAlpha(image.Rect(0, 0, width, spec.LineHeight))
	drawer := font.Drawer{
		Dst:  mask,
		Src:  image.White,
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(-left),
			Y: fixed.I(spec.OutlineRadius) + face.Metrics().Ascent,
		},
	}
	drawer.DrawString(text)

	advancePixels := max(1, fixedCeil(advance)+spec.LetterSpacing)
	return rasterizedGlyph{
		Code:    code,
		Advance: advancePixels,
		OffsetX: left,
		Pixels:  shadeGlyph(mask, palette, spec.OutlineRadius, spec.FlatPixels),
		Width:   width,
		Height:  spec.LineHeight,
	}
}

// shadeGlyph reduces antialiasing to intentional pixel edges and adds restrained depth.
// Palette-authored highlights survive runtime tinting, while index zero remains transparent.
func shadeGlyph(mask *image.Alpha, palette color.Palette, outlineRadius int, flat bool) []byte {
	width, height := mask.Bounds().Dx(), mask.Bounds().Dy()
	result := make([]byte, width*height)
	solid := make([]bool, len(result))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			threshold := uint8(112)
			if flat {
				threshold = 160
			}
			solid[y*width+x] = mask.AlphaAt(x, y).A >= threshold
		}
	}

	outline := uint8(palette.Index(color.NRGBA{R: 34, G: 29, B: 24, A: 255}))
	fill := uint8(palette.Index(color.NRGBA{R: 221, G: 204, B: 160, A: 255}))
	highlight := uint8(palette.Index(color.NRGBA{R: 255, G: 238, B: 187, A: 255}))
	shadow := uint8(palette.Index(color.NRGBA{R: 153, G: 119, B: 68, A: 255}))

	if outlineRadius > 0 {
		paintGlyphOutline(result, solid, width, height, outlineRadius, outline)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*width + x
			if !solid[index] {
				continue
			}

			result[index] = fill
			if flat {
				continue
			}
			if y == 0 || !solid[(y-1)*width+x] {
				result[index] = highlight
			} else if y+1 == height || !solid[(y+1)*width+x] {
				result[index] = shadow
			}
		}
	}

	return result
}

// paintGlyphOutline dilates the solid glyph mask into a compact eight-connected border.
// The radius remains explicit so tiny fonts can trade decoration for legibility.
func paintGlyphOutline(
	destination []byte,
	solid []bool,
	width int,
	height int,
	radius int,
	paletteIndex byte,
) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !solid[y*width+x] {
				continue
			}

			for offsetY := -radius; offsetY <= radius; offsetY++ {
				for offsetX := -radius; offsetX <= radius; offsetX++ {
					neighborX, neighborY := x+offsetX, y+offsetY
					if neighborX < 0 || neighborX >= width || neighborY < 0 || neighborY >= height {
						continue
					}

					destination[neighborY*width+neighborX] = paletteIndex
				}
			}
		}
	}
}

// writeFontTable emits the compact fixed-width records understood by the existing renderer.
// Frame order matches the sorted rune list exactly, preserving stable fallback and measurement behavior.
func writeFontTable(path string, glyphs []rasterizedGlyph) error {
	data := make([]byte, fontTableHeaderSize+len(glyphs)*fontGlyphRecordSize)
	copy(data, "Woo!\x01")
	for frame, glyph := range glyphs {
		offset := fontTableHeaderSize + frame*fontGlyphRecordSize
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(glyph.Code))
		data[offset+3] = byte(glyph.Advance)
		data[offset+4] = byte(glyph.Height)
		binary.LittleEndian.PutUint16(data[offset+8:offset+10], uint16(frame))
	}

	return os.WriteFile(path, data, 0o644)
}

// writeFontSheet emits one indexed DC6 frame per glyph and preserves horizontal bearings.
// The renderer applies those offsets during layout, so punctuation and italic overhangs do not clip.
func writeFontSheet(path string, glyphs []rasterizedGlyph) error {
	animation, err := dc6.New(1, len(glyphs))
	if err != nil {
		return err
	}

	for index, glyph := range glyphs {
		frame := animation.Directions[0].Frames[index]
		frame.Width = uint32(glyph.Width)
		frame.Height = uint32(glyph.Height)
		frame.OffsetX = int32(glyph.OffsetX)
		frame.IndexData = glyph.Pixels
	}

	data, err := dc6.Encode(animation)
	if err != nil {
		return fmt.Errorf("encode DC6: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// fontRunes returns the editor's deliberately bounded text repertoire in code-point order.
// Printable ASCII, Latin-1, and common editing symbols cover paths and metadata without oversized sheets.
func fontRunes() []rune {
	result := make([]rune, 0, 200)
	for code := rune(32); code <= 126; code++ {
		result = append(result, code)
	}
	for code := rune(160); code <= 255; code++ {
		result = append(result, code)
	}
	result = append(result, '←', '↑', '→', '↓', '•', '…', '✓', '✕')
	slices.Sort(result)

	return slices.Compact(result)
}

// fontVFSPath keeps manifest paths namespaced under the editor-owned content root.
func fontVFSPath(name, extension string) string {
	return "darkmagic/ds1-editor/ui/fonts/" + name + extension
}

// writeJSON publishes deterministic indented metadata with a trailing newline for clean diffs.
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// fixedFloor rounds a fixed-point coordinate toward negative infinity.
func fixedFloor(value fixed.Int26_6) int {
	return int(value >> 6)
}

// fixedCeil rounds a fixed-point coordinate toward positive infinity.
func fixedCeil(value fixed.Int26_6) int {
	return int((value + 63) >> 6)
}
