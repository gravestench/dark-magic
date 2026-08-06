// Package assetdecode decodes presentation assets stored in Diablo II content
// archives. It is deliberately independent of the archive implementation: any
// layered fs.FS can supply the bytes.
package assetdecode

import (
	"fmt"
	"image/color"
	"io/fs"

	dc6 "github.com/gravestench/dc6/pkg"
)

const (
	paletteColors   = 256
	bytesPerColor   = 3
	paletteByteSize = paletteColors * bytesPerColor
)

// Palette reads a Diablo II RGB palette. Palette index zero is transparent,
// which is the convention used by interface DC6 assets.
func Palette(source fs.FS, name string) (color.Palette, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("palette %q: %w", name, err)
	}
	if len(data) < paletteByteSize {
		return nil, fmt.Errorf("palette %q: got %d bytes, need at least %d", name, len(data), paletteByteSize)
	}
	palette := make(color.Palette, paletteColors)
	for index := range palette {
		offset := index * bytesPerColor
		palette[index] = color.RGBA{R: data[offset], G: data[offset+1], B: data[offset+2], A: 0xff}
	}
	palette[0] = color.RGBA{}
	return palette, nil
}

// DC6 reads and decodes a DC6, optionally applying a palette from the same
// content filesystem. Passing an empty palette name keeps the decoder's
// grayscale diagnostic palette.
func DC6(source fs.FS, name, paletteName string) (*dc6.DC6, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("dc6 %q: %w", name, err)
	}
	asset, err := dc6.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("dc6 %q: %w", name, err)
	}
	if paletteName != "" {
		palette, err := Palette(source, paletteName)
		if err != nil {
			return nil, err
		}
		asset.SetPalette(palette)
	}
	return asset, nil
}

// Frame returns one decoded DC6 frame with bounds checking.
func Frame(asset *dc6.DC6, direction, frame int) (*dc6.Frame, error) {
	if asset == nil {
		return nil, fmt.Errorf("dc6: nil asset")
	}
	if direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("dc6: direction %d out of range [0,%d)", direction, len(asset.Directions))
	}
	frames := asset.Directions[direction].Frames
	if frame < 0 || frame >= len(frames) {
		return nil, fmt.Errorf("dc6: frame %d out of range [0,%d) for direction %d", frame, len(frames), direction)
	}
	return frames[frame], nil
}
