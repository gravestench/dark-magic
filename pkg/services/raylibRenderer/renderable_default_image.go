package raylibRenderer

import (
	"image"
	"image/color"
	"image/draw"
)

// draws a black box with a green border and an X connecting the corners
func defaultImage(width, height int) image.Image {
	// Create a new RGBA image with a black background
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bgColor := color.Black
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw a green border (1px wide)
	borderColor := color.RGBA{0, 255, 0, 255}

	// Draw a green X connecting opposing corners
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x == y || x == width-y {
				img.Set(x, y, borderColor)
			}
		}
	}

	return img
}
