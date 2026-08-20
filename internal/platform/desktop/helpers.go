package desktop

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

type viewport struct{ x, y, width, height float64 }

// calculateViewport maps the logical surface into the native window. Contain preserves aspect ratio with centered
// letterboxing, while stretch occupies the entire window; invalid dimensions fail before any division occurs.
func calculateViewport(
	windowWidth, windowHeight, gameWidth, gameHeight int,
	fit string,
) (viewport, error) {
	if windowWidth <= 0 || windowHeight <= 0 || gameWidth <= 0 || gameHeight <= 0 {
		return viewport{}, errors.New("viewport dimensions must be positive")
	}

	if fit == "stretch" {
		return viewport{width: float64(windowWidth), height: float64(windowHeight)}, nil
	}

	if fit != "" && fit != "contain" {
		return viewport{}, fmt.Errorf("unknown viewport fit %q", fit)
	}

	scale := math.Min(
		float64(windowWidth)/float64(gameWidth),
		float64(windowHeight)/float64(gameHeight),
	)
	width, height := float64(gameWidth)*scale, float64(gameHeight)*scale

	return viewport{
		x:      (float64(windowWidth) - width) / 2,
		y:      (float64(windowHeight) - height) / 2,
		width:  width,
		height: height,
	}, nil
}

// contiguousRGBA copies any image and non-zero origin into tightly packed, zero-origin RGBA storage required by native
// texture upload paths.
func contiguousRGBA(source image.Image) *image.RGBA {
	bounds := source.Bounds()
	result := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(result, result.Bounds(), source, bounds.Min, draw.Src)

	return result
}

// quantizeImage chooses the nearest palette RGB value for every visible pixel while preserving source alpha. Fully
// transparent pixels are left untouched so hidden RGB data cannot produce palette fringes later.
func quantizeImage(source image.Image, palette color.Palette) *image.RGBA {
	rgba := contiguousRGBA(source)

	colors := make([]color.RGBA, len(palette))
	for index, entry := range palette {
		colors[index] = color.RGBAModel.Convert(entry).(color.RGBA)
	}

	for offset := 0; offset < len(rgba.Pix); offset += 4 {
		if rgba.Pix[offset+3] == 0 {
			continue
		}

		best, distance := color.RGBA{}, int(^uint(0)>>1)

		// Alpha is excluded from distance so palette entries select visible color
		// while the assignment below leaves the source coverage byte unchanged.
		for _, candidate := range colors {
			red := int(rgba.Pix[offset]) - int(candidate.R)
			green := int(rgba.Pix[offset+1]) - int(candidate.G)

			blue := int(rgba.Pix[offset+2]) - int(candidate.B)
			if candidateDistance := red*red + green*green + blue*blue; candidateDistance < distance {
				best, distance = candidate, candidateDistance
			}
		}

		rgba.Pix[offset], rgba.Pix[offset+1], rgba.Pix[offset+2] = best.R, best.G, best.B
	}

	return rgba
}
