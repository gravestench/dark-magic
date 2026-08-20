package assetdecode

import (
	"image/color"
	"strings"
	"unicode/utf8"
)

// textColorRun couples an authored transform with its direct-color fallback;
// reset runs deliberately carry neither so the caller's tint can be restored.
type textColorRun struct {
	transform int
	fallback  color.Color
	reset     bool
}

// namedTextColors preserves the transform slots used by Diablo UI strings and
// supplies stable RGB approximations when a font has no PL2 transform bank.
var namedTextColors = map[string]textColorRun{
	// PL2 text shift zero is the reserved or unshifted slot and is commonly an
	// all-zero lookup table. White therefore selects the palette-authored font.
	"white":  {transform: -1, fallback: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}},
	"red":    {transform: 1, fallback: color.RGBA{R: 0xff, G: 0x4d, B: 0x4d, A: 0xff}},
	"green":  {transform: 2, fallback: color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xff}},
	"blue":   {transform: 3, fallback: color.RGBA{R: 0x69, G: 0x69, B: 0xff, A: 0xff}},
	"gold":   {transform: 4, fallback: color.RGBA{R: 0xc7, G: 0xb3, B: 0x77, A: 0xff}},
	"grey":   {transform: 5, fallback: color.RGBA{R: 0x69, G: 0x69, B: 0x69, A: 0xff}},
	"black":  {transform: 6, fallback: color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}},
	"orange": {transform: 8, fallback: color.RGBA{R: 0xff, G: 0xa8, B: 0x00, A: 0xff}},
	"yellow": {transform: 9, fallback: color.RGBA{R: 0xff, G: 0xff, B: 0x64, A: 0xff}},
}

// parseColorTokens removes Diablo label tokens and records the visible-rune
// position where each color run begins. Reset tokens restore the caller's tint,
// so a colored phrase does not require a hard-coded token for the base color.
func parseColorTokens(text string) (string, map[int]textColorRun) {
	var clean strings.Builder

	colors := make(map[int]textColorRun)
	visible := 0

	for len(text) > 0 {
		if text[0] == '[' {
			if end := strings.IndexByte(text, ']'); end >= 0 {
				token := strings.ToLower(strings.TrimSpace(text[1:end]))
				if token == "/" || token == "reset" || strings.HasPrefix(token, "/") {
					colors[visible] = textColorRun{reset: true}
				} else if next, ok := namedTextColors[token]; ok {
					colors[visible] = next
				}

				text = text[end+1:]

				continue
			}
		}

		code, size := utf8.DecodeRuneInString(text)
		clean.WriteRune(code)

		if code != '\n' {
			visible++
		}

		text = text[size:]
	}

	return clean.String(), colors
}
