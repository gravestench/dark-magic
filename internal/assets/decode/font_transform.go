package assetdecode

import (
	"image"
	"image/color"
	"io/fs"

	dc6 "github.com/gravestench/dc6/pkg"
	pl2 "github.com/gravestench/pl2"
)

// loadTextTransformFrames decodes optional PL2 banks into frame sets parallel
// to BitmapFont.Frames; any load or decode failure deliberately leaves direct tinting available.
func loadTextTransformFrames(
	source fs.FS,
	transformName string,
	sheet *dc6.DC6,
) map[int][]image.Image {
	if transformName == "" || sheet == nil {
		return nil
	}

	file, err := source.Open(transformName)
	if err != nil {
		return nil
	}
	defer file.Close() //nolint:errcheck // Optional transforms ignore read-only close failures with load failures.

	transforms, err := pl2.DecodeReader(file)
	if err != nil || len(transforms.TextColorShifts) == 0 {
		return nil
	}

	result := make(map[int][]image.Image, len(transforms.TextColorShifts))
	for transformIndex, transform := range transforms.TextColorShifts {
		result[transformIndex] = transformFontFrames(sheet, transforms.BasePalette, transform)
	}

	return result
}

// transformFontFrames preserves direction-major ordering while applying one
// PL2 lookup bank, allowing Woo! frame indices to address every transformed set identically.
func transformFontFrames(
	sheet *dc6.DC6,
	palette color.Palette,
	transform pl2.Transform,
) []image.Image {
	frames := make([]image.Image, 0)

	for _, direction := range sheet.Directions {
		for _, frame := range direction.Frames {
			frames = append(frames, transformFontFrame(frame, palette, transform))
		}
	}

	return frames
}

// transformFontFrame retains index-zero transparency and maps only the pixels
// present in the decoded frame, matching the codec's handling of truncated index data.
func transformFontFrame(
	frame *dc6.Frame,
	palette color.Palette,
	transform pl2.Transform,
) image.Image {
	decoded := image.NewRGBA(image.Rect(0, 0, int(frame.Width), int(frame.Height)))

	pixels := min(len(frame.IndexData), int(frame.Width)*int(frame.Height))
	for pixel, paletteIndex := range frame.IndexData[:pixels] {
		if paletteIndex == 0 {
			continue
		}

		x := pixel % int(frame.Width)
		y := pixel / int(frame.Width)
		decoded.Set(x, y, palette[transform[paletteIndex]])
	}

	return decoded
}
