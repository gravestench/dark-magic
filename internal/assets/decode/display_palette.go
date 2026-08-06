package assetdecode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
)

// DisplayPalette loads an arbitrary non-empty quantization palette. JSON files
// contain either an array of CSS-style hex colors or {"colors":[...]}; GIMP
// .gpl files contain RGB rows with optional names; all other files are
// interpreted as packed BGR triples, including normal 768-byte Diablo pal.dat
// files and deliberately tiny mod palettes.
func DisplayPalette(source fs.FS, name string) (color.Palette, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("display palette %q: %w", name, err)
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json":
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
	case ".gpl":
		palette, err := parseGIMPPalette(data)
		if err != nil {
			return nil, fmt.Errorf("display palette %q: %w", name, err)
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

func parseGIMPPalette(data []byte) (color.Palette, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNumber := 0
	headerSeen := false
	palette := color.Palette{}
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if !headerSeen {
			if line != "GIMP Palette" {
				return nil, fmt.Errorf("line %d: expected GIMP Palette header", lineNumber)
			}
			headerSeen = true
			continue
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "Name:") || strings.HasPrefix(line, "Columns:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("line %d: expected red, green, and blue values", lineNumber)
		}
		var channels [3]uint8
		for index, field := range fields[:3] {
			value, err := strconv.Atoi(field)
			if err != nil || value < 0 || value > 255 {
				return nil, fmt.Errorf("line %d: RGB value %q must be an integer from 0 through 255", lineNumber, field)
			}
			channels[index] = uint8(value)
		}
		palette = append(palette, color.RGBA{R: channels[0], G: channels[1], B: channels[2], A: 0xff})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read GPL: %w", err)
	}
	if !headerSeen {
		return nil, fmt.Errorf("expected GIMP Palette header")
	}
	if len(palette) == 0 {
		return nil, fmt.Errorf("GPL palette is empty")
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
