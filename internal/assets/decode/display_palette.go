package assetdecode

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

// DisplayPalette loads an arbitrary non-empty quantization palette. JSON files
// contain either an array of CSS-style hex colors or {"colors":[...]}; all
// other files are interpreted as packed BGR triples, including normal 768-byte
// Diablo pal.dat files and deliberately tiny mod palettes.
func DisplayPalette(source fs.FS, name string) (color.Palette, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("display palette %q: %w", name, err)
	}
	if strings.EqualFold(filepath.Ext(name), ".json") {
		var colors []string
		if err := json.Unmarshal(data, &colors); err != nil {
			var document struct {
				Colors []string `json:"colors"`
			}
			if objectErr := json.Unmarshal(data, &document); objectErr != nil {
				return nil, fmt.Errorf("display palette %q: decode JSON: %w", name, err)
			}
			colors = document.Colors
		}
		palette := make(color.Palette, len(colors))
		for index, value := range colors {
			parsed, err := parseHexColor(value)
			if err != nil {
				return nil, fmt.Errorf("display palette %q color %d: %w", name, index, err)
			}
			palette[index] = parsed
		}
		if len(palette) == 0 {
			return nil, fmt.Errorf("display palette %q is empty", name)
		}
		return palette, nil
	}
	if len(data) == 0 || len(data)%3 != 0 {
		return nil, fmt.Errorf("display palette %q: packed BGR length %d is not a positive multiple of 3", name, len(data))
	}
	palette := make(color.Palette, len(data)/3)
	for index := range palette {
		offset := index * 3
		palette[index] = color.RGBA{R: data[offset+2], G: data[offset+1], B: data[offset], A: 0xff}
	}
	return palette, nil
}

func parseHexColor(value string) (color.RGBA, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) != 6 && len(value) != 8 {
		return color.RGBA{}, fmt.Errorf("%q must contain RRGGBB or RRGGBBAA", value)
	}
	parsed, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("parse %q: %w", value, err)
	}
	if len(value) == 6 {
		return color.RGBA{R: uint8(parsed >> 16), G: uint8(parsed >> 8), B: uint8(parsed), A: 0xff}, nil
	}
	return color.RGBA{R: uint8(parsed >> 24), G: uint8(parsed >> 16), B: uint8(parsed >> 8), A: uint8(parsed)}, nil
}
