package assetdecode

import (
	"encoding/binary"
	"fmt"
)

const (
	fontHeaderBytes = 12
	fontGlyphBytes  = 14
)

// FontTable decodes Diablo II's Woo! glyph metric table with strict record
// bounds, preventing truncated dimensions or frame indices from reaching layout.
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
		code, glyph := decodeFontGlyph(body[offset : offset+fontGlyphBytes])
		if glyph.Width <= 0 || glyph.Height <= 0 {
			return nil, fmt.Errorf(
				"font table: glyph %U has invalid size %dx%d",
				code,
				glyph.Width,
				glyph.Height,
			)
		}

		glyphs[code] = glyph
	}

	return glyphs, nil
}

// decodeFontGlyph isolates the fixed binary offsets that define one complete
// Woo! record, keeping format knowledge out of validation and map construction.
func decodeFontGlyph(record []byte) (rune, Glyph) {
	code := rune(binary.LittleEndian.Uint16(record[0:2]))
	glyph := Glyph{
		Width:  int(record[3]),
		Height: int(record[4]),
		Frame:  int(binary.LittleEndian.Uint16(record[8:10])),
	}

	return code, glyph
}
