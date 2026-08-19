package assetdecode

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/fs"

	dcc "github.com/gravestench/dcc/pkg"
)

// DCC reads and decodes a palette-aware character or monster animation. It
// prefers random access but retains the byte-buffer fallback required by basic fs.File implementations.
func DCC(source fs.FS, name, paletteName string) (*dcc.DCC, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("dcc %q: %w", name, err)
	}
	defer file.Close() //nolint:errcheck // Decode errors retain precedence over read-only close failures.

	palette, err := optionalPalette(source, paletteName)
	if err != nil {
		return nil, err
	}

	asset, err := decodeDCCFile(file, palette)
	if err != nil {
		return nil, fmt.Errorf("dcc %q: %w", name, err)
	}

	return asset, nil
}

// decodeDCCFile selects the most efficient codec entry point supported by the
// file while applying the palette at the stage required by each codec API.
func decodeDCCFile(file fs.File, palette color.Palette) (*dcc.DCC, error) {
	if random, size, ok := randomAccess(file); ok {
		opened, err := dcc.Open(random, size)
		if err != nil {
			return nil, err
		}

		if palette != nil {
			opened.SetPalette(palette)
		}

		return opened.Decode()
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	asset, err := dcc.FromBytes(data)
	if err == nil && palette != nil {
		asset.SetPalette(palette)
	}

	return asset, err
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
