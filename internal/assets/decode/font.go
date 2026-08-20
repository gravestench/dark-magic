package assetdecode

import "image"

// Glyph maps one Unicode code point to its authored DC6 frame and logical
// advance box. Frame pixels retain palette shading independently of metrics.
type Glyph struct {
	Frame         int
	Width, Height int
}

// BitmapFont is a fully decoded, renderer-neutral Diablo bitmap font. It owns
// CPU images only; retained or native resource upload belongs to presentation.
type BitmapFont struct {
	Glyphs       map[rune]Glyph
	Frames       []image.Image
	FrameOffsets []image.Point
	TextFrames   map[int][]image.Image
	LineHeight   int
}
