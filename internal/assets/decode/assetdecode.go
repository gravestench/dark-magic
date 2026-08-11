// Package assetdecode decodes presentation assets stored in Diablo II content
// archives. It is deliberately independent of the archive implementation: any
// layered fs.FS can supply the bytes.
package assetdecode

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
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
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("palette %q: %w", name, err)
	}
	defer file.Close()
	data := make([]byte, paletteByteSize)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, fmt.Errorf("palette %q: need %d bytes: %w", name, paletteByteSize, err)
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
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("cof %q: %w", name, err)
	}
	defer file.Close()
	asset, err := cof.UnmarshalReader(file)
	if err != nil {
		return nil, fmt.Errorf("cof %q: %w", name, err)
	}
	return asset, nil
}

// AnimationData reads the global typed timing/event catalog used by composite
// units. The codec consumes the stream directly; callers never need to buffer
// the complete binary file or know its 256-block layout.
func AnimationData(source fs.FS, name string) (*cof.AnimationData, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("animation data %q: %w", name, err)
	}
	defer file.Close()
	asset, err := cof.LoadReader(file)
	if err != nil {
		return nil, fmt.Errorf("animation data %q: %w", name, err)
	}
	return asset, nil
}

// DCC reads and decodes a palette-aware character or monster animation.
func DCC(source fs.FS, name, paletteName string) (*dcc.DCC, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("dcc %q: %w", name, err)
	}
	defer file.Close()
	var palette color.Palette
	if paletteName != "" {
		palette, err = Palette(source, paletteName)
		if err != nil {
			return nil, err
		}
	}
	var asset *dcc.DCC
	if random, size, ok := randomAccess(file); ok {
		opened, err := dcc.Open(random, size)
		if err != nil {
			return nil, fmt.Errorf("dcc %q: %w", name, err)
		}
		if palette != nil {
			opened.SetPalette(palette)
		}
		asset, err = opened.Decode()
	} else {
		data, readErr := io.ReadAll(file)
		if readErr != nil {
			return nil, fmt.Errorf("dcc %q: %w", name, readErr)
		}
		asset, err = dcc.FromBytes(data)
		if err == nil && palette != nil {
			asset.SetPalette(palette)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("dcc %q: %w", name, err)
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
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("dc6 %q: %w", name, err)
	}
	defer file.Close()
	var palette color.Palette
	if paletteName != "" {
		palette, err = Palette(source, paletteName)
		if err != nil {
			return nil, err
		}
	}
	var asset *dc6.DC6
	if random, size, ok := randomAccess(file); ok {
		opened, err := dc6.Open(random, size)
		if err != nil {
			return nil, fmt.Errorf("dc6 %q: %w", name, err)
		}
		if palette != nil {
			opened.SetPalette(palette)
		}
		asset, err = opened.Decode()
	} else {
		data, readErr := io.ReadAll(file)
		if readErr != nil {
			return nil, fmt.Errorf("dc6 %q: %w", name, readErr)
		}
		asset, err = dc6.FromBytes(data)
		if err == nil && palette != nil {
			asset.SetPalette(palette)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("dc6 %q: %w", name, err)
	}
	return asset, nil
}

func randomAccess(file fs.File) (io.ReaderAt, int64, bool) {
	reader, ok := file.(io.ReaderAt)
	if !ok {
		return nil, 0, false
	}
	info, err := file.Stat()
	if err != nil || info.Size() < 0 {
		return nil, 0, false
	}
	return reader, info.Size(), true
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
	colors := make([]color.RGBA, len(palette))
	for index, value := range palette {
		colors[index] = color.RGBAModel.Convert(value).(color.RGBA)
	}
	result := image.NewRGBA(image.Rect(0, 0, int(frame.Width), int(frame.Height)))
	for pixelIndex, paletteIndex := range frame.IndexData[:pixels] {
		if int(paletteIndex) >= len(colors) {
			return nil, fmt.Errorf("dc6: palette index %d is outside %d colors", paletteIndex, len(colors))
		}
		converted := colors[paletteIndex]
		offset := pixelIndex * 4
		result.Pix[offset], result.Pix[offset+1], result.Pix[offset+2], result.Pix[offset+3] = converted.R, converted.G, converted.B, converted.A
	}
	return result, nil
}
