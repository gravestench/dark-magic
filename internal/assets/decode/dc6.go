package assetdecode

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"

	dc6 "github.com/gravestench/dc6/pkg"
)

// DC6 reads and decodes a DC6, optionally applying a palette from the same
// content filesystem. An empty palette name retains the codec's diagnostic palette.
func DC6(source fs.FS, name, paletteName string) (*dc6.DC6, error) {
	file, err := source.Open(name)
	if err != nil {
		return nil, fmt.Errorf("dc6 %q: %w", name, err)
	}
	defer file.Close() //nolint:errcheck // Decode errors retain precedence over read-only close failures.

	palette, err := optionalPalette(source, paletteName)
	if err != nil {
		return nil, err
	}

	asset, err := decodeDC6File(file, palette)
	if err != nil {
		return nil, fmt.Errorf("dc6 %q: %w", name, err)
	}

	return asset, nil
}

// decodeDC6File selects the random-access codec when the archive exposes it
// and otherwise preserves compatibility with sequential-only file implementations.
func decodeDC6File(file fs.File, palette color.Palette) (*dc6.DC6, error) {
	if random, size, ok := randomAccess(file); ok {
		opened, err := dc6.Open(random, size)
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

	asset, err := dc6.FromBytes(data)
	if err == nil && palette != nil {
		asset.SetPalette(palette)
	}

	return asset, err
}

// Frame returns one decoded DC6 frame with bounds checking, preventing codec
// internals from leaking panics through asset lookup paths.
func Frame(asset *dc6.DC6, direction, frame int) (*dc6.Frame, error) {
	if asset == nil {
		return nil, fmt.Errorf("dc6: nil asset")
	}

	if direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("dc6: direction %d out of range [0,%d)", direction, len(asset.Directions))
	}

	frames := asset.Directions[direction].Frames
	if frame < 0 || frame >= len(frames) {
		return nil, fmt.Errorf(
			"dc6: frame %d out of range [0,%d) for direction %d",
			frame,
			len(frames),
			direction,
		)
	}

	return frames[frame], nil
}

// FrameImage converts a DC6 frame using its explicit owning asset. This avoids
// relying on a private frame-to-asset back-pointer that may be absent after cloning or caching.
func FrameImage(asset *dc6.DC6, frame *dc6.Frame) (*image.RGBA, error) {
	if asset == nil || frame == nil {
		return nil, fmt.Errorf("dc6: asset and frame are required")
	}

	pixels := uint64(frame.Width) * uint64(frame.Height)
	if pixels > uint64(len(frame.IndexData)) {
		return nil, fmt.Errorf("dc6: frame index data has %d bytes, need %d", len(frame.IndexData), pixels)
	}

	colors := rgbaPalette(asset.Palette())

	result := image.NewRGBA(image.Rect(0, 0, int(frame.Width), int(frame.Height)))
	for pixelIndex, paletteIndex := range frame.IndexData[:pixels] {
		if int(paletteIndex) >= len(colors) {
			return nil, fmt.Errorf("dc6: palette index %d is outside %d colors", paletteIndex, len(colors))
		}

		converted := colors[paletteIndex]
		offset := pixelIndex * 4
		result.Pix[offset] = converted.R
		result.Pix[offset+1] = converted.G
		result.Pix[offset+2] = converted.B
		result.Pix[offset+3] = converted.A
	}

	return result, nil
}

// rgbaPalette converts colors once per frame so indexed pixels can be copied
// without repeating color-model conversions inside the inner loop.
func rgbaPalette(palette color.Palette) []color.RGBA {
	colors := make([]color.RGBA, len(palette))
	for index, value := range palette {
		colors[index] = color.RGBAModel.Convert(value).(color.RGBA)
	}

	return colors
}
