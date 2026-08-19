package assetdecode

import (
	"fmt"
	"image/color"
	"io"
	"io/fs"
)

const (
	paletteColors   = 256
	bytesPerColor   = 3
	paletteByteSize = paletteColors * bytesPerColor
)

// Palette reads a complete Diablo II BGR palette and reserves index zero for
// transparency, matching the convention used by interface DC6 assets.
func Palette(source fs.FS, name string) (color.Palette, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("palette %q: %w", name, err)
	}
	defer file.Close() //nolint:errcheck // Read and validation errors retain precedence over close failures.

	data := make([]byte, paletteByteSize)
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, fmt.Errorf("palette %q: need %d bytes: %w", name, paletteByteSize, err)
	}

	palette := make(color.Palette, paletteColors)
	for index := range palette {
		offset := index * bytesPerColor
		palette[index] = color.RGBA{
			R: data[offset+2],
			G: data[offset+1],
			B: data[offset],
			A: 0xff,
		}
	}

	// Interface sprites rely on palette index zero remaining fully transparent.
	palette[0] = color.RGBA{}

	return palette, nil
}

// optionalPalette preserves the decoder's diagnostic palette when callers do
// not provide a palette name, while keeping palette failures distinct from
// codec failures.
func optionalPalette(source fs.FS, name string) (color.Palette, error) {
	if name == "" {
		return nil, nil
	}

	return Palette(source, name)
}
