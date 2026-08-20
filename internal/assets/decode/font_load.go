package assetdecode

import (
	"fmt"
	"image"
	"io/fs"

	dc6 "github.com/gravestench/dc6/pkg"
)

// LoadBitmapFont joins Woo! metrics, DC6 glyph frames, and a palette without a
// PL2 text transform, retaining the simpler loading contract for direct-color fonts.
func LoadBitmapFont(
	source fs.FS,
	tableName string,
	sheetName string,
	paletteName string,
) (*BitmapFont, error) {
	return LoadBitmapFontWithTransform(source, tableName, sheetName, paletteName, "")
}

// LoadBitmapFontWithTransform additionally loads authored PL2 text-color banks.
// Invalid optional transforms fall back to direct tinting, but core font assets
// remain strict because missing metrics or glyph frames make layout unsafe.
func LoadBitmapFontWithTransform(
	source fs.FS,
	tableName string,
	sheetName string,
	paletteName string,
	transformName string,
) (*BitmapFont, error) {
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
	if err := appendFontSheetFrames(font, sheet); err != nil {
		return nil, err
	}

	// Text transforms are an optional enhancement; the untransformed frames
	// remain usable when a mod omits or corrupts its PL2 file.
	font.TextFrames = loadTextTransformFrames(source, transformName, sheet)
	if err := validateFontGlyphFrames(font, tableName); err != nil {
		return nil, err
	}

	return font, nil
}

// appendFontSheetFrames flattens direction-major DC6 frames and records their
// authored offsets in the same order expected by Woo! frame indices.
func appendFontSheetFrames(font *BitmapFont, sheet *dc6.DC6) error {
	for _, direction := range sheet.Directions {
		for _, frame := range direction.Frames {
			decoded, err := FrameImage(sheet, frame)
			if err != nil {
				return err
			}

			font.Frames = append(font.Frames, decoded)
			font.FrameOffsets = append(
				font.FrameOffsets,
				image.Pt(int(frame.OffsetX), int(frame.OffsetY)),
			)
		}
	}

	return nil
}

// validateFontGlyphFrames rejects dangling metric references and derives line
// height from visual frames so authored offsets do not distort vertical spacing.
func validateFontGlyphFrames(font *BitmapFont, tableName string) error {
	for code, glyph := range font.Glyphs {
		if glyph.Frame < 0 || glyph.Frame >= len(font.Frames) {
			return fmt.Errorf(
				"font table %q: glyph %U frame %d out of range",
				tableName,
				code,
				glyph.Frame,
			)
		}

		visualHeight := font.Frames[glyph.Frame].Bounds().Dy()
		if visualHeight > font.LineHeight {
			font.LineHeight = visualHeight
		}
	}

	return nil
}
