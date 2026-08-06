package assetdecode

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/fs"
	"strings"
)

const (
	fontHeaderBytes = 12
	fontGlyphBytes  = 14
)

type Glyph struct {
	Frame         int
	Width, Height int
}

type BitmapFont struct {
	Glyphs       map[rune]Glyph
	Frames       []image.Image
	FrameOffsets []image.Point
	LineHeight   int
}

// FontTable decodes Diablo II's Woo! glyph metric table with strict bounds.
func FontTable(data []byte) (map[rune]Glyph, error) {
	if len(data) < fontHeaderBytes || string(data[:5]) != "Woo!\x01" {
		return nil, fmt.Errorf("font table: invalid or truncated header")
	}
	body := data[fontHeaderBytes:]
	if len(body)%fontGlyphBytes != 0 {
		return nil, fmt.Errorf("font table: truncated glyph record")
	}
	glyphs := make(map[rune]Glyph, len(body)/fontGlyphBytes)
	for offset := 0; offset < len(body); offset += fontGlyphBytes {
		record := body[offset : offset+fontGlyphBytes]
		code := rune(binary.LittleEndian.Uint16(record[0:2]))
		glyph := Glyph{Width: int(record[3]), Height: int(record[4]), Frame: int(binary.LittleEndian.Uint16(record[8:10]))}
		if glyph.Width <= 0 || glyph.Height <= 0 {
			return nil, fmt.Errorf("font table: glyph %U has invalid size %dx%d", code, glyph.Width, glyph.Height)
		}
		glyphs[code] = glyph
	}
	return glyphs, nil
}

func LoadBitmapFont(source fs.FS, tableName, sheetName, paletteName string) (*BitmapFont, error) {
	table, err := fs.ReadFile(source, tableName)
	if err != nil {
		return nil, fmt.Errorf("font table %q: %w", tableName, err)
	}
	glyphs, err := FontTable(table)
	if err != nil {
		return nil, fmt.Errorf("font table %q: %w", tableName, err)
	}
	sheet, err := DC6(source, sheetName, paletteName)
	if err != nil {
		return nil, err
	}
	font := &BitmapFont{Glyphs: glyphs}
	for _, direction := range sheet.Directions {
		for _, frame := range direction.Frames {
			decoded, err := FrameImage(sheet, frame)
			if err != nil {
				return nil, err
			}
			font.Frames = append(font.Frames, decoded)
			font.FrameOffsets = append(font.FrameOffsets, image.Pt(int(frame.OffsetX), int(frame.OffsetY)))
		}
	}
	for code, glyph := range glyphs {
		if glyph.Frame < 0 || glyph.Frame >= len(font.Frames) {
			return nil, fmt.Errorf("font table %q: glyph %U frame %d out of range", tableName, code, glyph.Frame)
		}
		visualHeight := font.Frames[glyph.Frame].Bounds().Dy()
		if visualHeight > font.LineHeight {
			font.LineHeight = visualHeight
		}
	}
	return font, nil
}

func (f *BitmapFont) Render(text string, tint color.Color, maxWidth int, align string) (*image.RGBA, error) {
	if f == nil || len(f.Glyphs) == 0 || len(f.Frames) == 0 || f.LineHeight <= 0 {
		return nil, fmt.Errorf("bitmap font: empty font")
	}
	if align != "left" && align != "center" && align != "right" {
		return nil, fmt.Errorf("bitmap font: invalid alignment %q", align)
	}
	lines := f.wrap(text, maxWidth)
	width := 1
	for _, line := range lines {
		if measured := f.measure(line); measured > width {
			width = measured
		}
	}
	if maxWidth > 0 {
		width = maxWidth
	}
	output := image.NewRGBA(image.Rect(0, 0, width, maxInt(1, len(lines)*f.LineHeight)))
	for lineIndex, line := range lines {
		lineWidth := f.measure(line)
		x := 0
		if align == "center" {
			x = (width - lineWidth) / 2
		} else if align == "right" {
			x = width - lineWidth
		}
		for _, code := range line {
			glyph, ok := f.glyph(code)
			if !ok {
				continue
			}
			frame := f.Frames[glyph.Frame]
			bounds := frame.Bounds()
			offset := image.Point{}
			if glyph.Frame < len(f.FrameOffsets) {
				offset = f.FrameOffsets[glyph.Frame]
			}
			top := lineIndex*f.LineHeight + offset.Y
			left := x + offset.X
			destination := image.Rect(left, top, left+bounds.Dx(), top+bounds.Dy())
			draw.Draw(output, destination, modulatedImage{Image: frame, Tint: tint}, bounds.Min, draw.Over)
			x += glyph.Width
		}
	}
	return output, nil
}

func (f *BitmapFont) wrap(text string, maxWidth int) []string {
	paragraphs := strings.Split(text, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if maxWidth <= 0 {
			lines = append(lines, paragraph)
			continue
		}
		var line []rune
		width := 0
		for _, code := range paragraph {
			glyph, ok := f.glyph(code)
			if !ok {
				continue
			}
			if width > 0 && width+glyph.Width > maxWidth {
				lines = append(lines, string(line))
				line, width = nil, 0
			}
			line = append(line, code)
			width += glyph.Width
		}
		lines = append(lines, string(line))
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func (f *BitmapFont) measure(text string) int {
	width := 0
	for _, code := range text {
		if glyph, ok := f.glyph(code); ok {
			width += glyph.Width
		}
	}
	return width
}

func (f *BitmapFont) glyph(code rune) (Glyph, bool) {
	if glyph, ok := f.Glyphs[code]; ok {
		return glyph, true
	}
	glyph, ok := f.Glyphs['?']
	return glyph, ok
}

// modulatedImage preserves the palette-authored shading within a glyph while
// applying the label color like the original sprite renderer. Treating a glyph
// as a mask flattens its highlights, shadows, and antialiased edge colors.
type modulatedImage struct {
	image.Image
	Tint color.Color
}

func (m modulatedImage) At(x, y int) color.Color {
	red, green, blue, alpha := m.Image.At(x, y).RGBA()
	tintRed, tintGreen, tintBlue, tintAlpha := m.Tint.RGBA()
	return color.RGBA64{
		R: uint16(uint64(red) * uint64(tintRed) / 0xffff),
		G: uint16(uint64(green) * uint64(tintGreen) / 0xffff),
		B: uint16(uint64(blue) * uint64(tintBlue) / 0xffff),
		A: uint16(uint64(alpha) * uint64(tintAlpha) / 0xffff),
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
