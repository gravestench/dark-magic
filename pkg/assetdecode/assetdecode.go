// Package assetdecode decodes presentation assets stored in Diablo II content
// archives. It is deliberately independent of the archive implementation: any
// layered fs.FS can supply the bytes.
package assetdecode

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/fs"

	cof "github.com/gravestench/cof"
	dc6 "github.com/gravestench/dc6/pkg"
	dcc "github.com/gravestench/dcc/pkg"
)

const (
	paletteColors   = 256
	bytesPerColor   = 3
	paletteByteSize = paletteColors * bytesPerColor
)

// Palette reads a Diablo II BGR palette. Palette index zero is transparent,
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
		palette[index] = color.RGBA{R: data[offset+2], G: data[offset+1], B: data[offset], A: 0xff}
	}
	palette[0] = color.RGBA{}
	return palette, nil
}

// COF reads composite ordering, layer, timing, and event metadata.
func COF(source fs.FS, name string) (*cof.COF, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("cof %q: %w", name, err)
	}
	asset, err := cof.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("cof %q: %w", name, err)
	}
	return asset, nil
}

// DCC reads and decodes a palette-aware character or monster animation.
func DCC(source fs.FS, name, paletteName string) (*dcc.DCC, error) {
	data, err := fs.ReadFile(source, name)
	if err != nil {
		return nil, fmt.Errorf("dcc %q: %w", name, err)
	}
	asset, err := dcc.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("dcc %q: %w", name, err)
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

// DCCFrames returns normalized images for one direction. Normalization keeps
// frame placement metadata separate from zero-based texture pixels.
func DCCFrames(asset *dcc.DCC, direction int) ([]image.Image, error) {
	if asset == nil {
		return nil, fmt.Errorf("dcc: nil asset")
	}
	directions := asset.Directions()
	if direction < 0 || direction >= len(directions) {
		return nil, fmt.Errorf("dcc: direction %d out of range [0,%d)", direction, len(directions))
	}
	frames := directions[direction].Frames()
	result := make([]image.Image, len(frames))
	for index, frame := range frames {
		bounds := frame.Bounds()
		normalized := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
		draw.Draw(normalized, normalized.Bounds(), frame, bounds.Min, draw.Src)
		result[index] = normalized
	}
	return result, nil
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

// FrameImage converts a DC6 frame using its explicit owning asset. This avoids
// relying on the decoder's private frame-to-asset back-pointer, which is not
// part of its public contract and may be absent after cloning or caching.
func FrameImage(asset *dc6.DC6, frame *dc6.Frame) (*image.RGBA, error) {
	if asset == nil || frame == nil {
		return nil, fmt.Errorf("dc6: asset and frame are required")
	}
	pixels := uint64(frame.Width) * uint64(frame.Height)
	if pixels > uint64(len(frame.IndexData)) {
		return nil, fmt.Errorf("dc6: frame index data has %d bytes, need %d", len(frame.IndexData), pixels)
	}
	palette := asset.Palette()
	result := image.NewRGBA(image.Rect(0, 0, int(frame.Width), int(frame.Height)))
	for y := 0; y < int(frame.Height); y++ {
		for x := 0; x < int(frame.Width); x++ {
			result.Set(x, y, palette[frame.IndexData[y*int(frame.Width)+x]])
		}
	}
	return result, nil
}
