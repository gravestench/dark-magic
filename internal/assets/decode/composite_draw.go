package assetdecode

import (
	"image"
	"image/color"
	"image/draw"
)

// blendRGBA alpha-blends one pixel into an RGBA canvas while clipping projected
// shadows and offset components that extend beyond the allocated image.
func blendRGBA(destination *image.RGBA, x, y int, source color.RGBA, opacity uint8) {
	if !image.Pt(x, y).In(destination.Bounds()) || source.A == 0 || opacity == 0 {
		return
	}

	alpha := uint32(source.A) * uint32(opacity) / 255

	offset := destination.PixOffset(x, y)
	if alpha == 255 {
		destination.Pix[offset] = source.R
		destination.Pix[offset+1] = source.G
		destination.Pix[offset+2] = source.B
		destination.Pix[offset+3] = 255

		return
	}

	inverse := 255 - alpha
	destination.Pix[offset] = uint8(
		(uint32(source.R)*alpha + uint32(destination.Pix[offset])*inverse) / 255,
	)
	destination.Pix[offset+1] = uint8(
		(uint32(source.G)*alpha + uint32(destination.Pix[offset+1])*inverse) / 255,
	)
	destination.Pix[offset+2] = uint8(
		(uint32(source.B)*alpha + uint32(destination.Pix[offset+2])*inverse) / 255,
	)
	destination.Pix[offset+3] = uint8(alpha + uint32(destination.Pix[offset+3])*inverse/255)
}

// drawCompositeComponent prefers indexed pixels when available because their
// palette indices preserve legacy transparency and authored color semantics.
func drawCompositeComponent(
	output *image.RGBA,
	component CompositeFrame,
	destination image.Point,
	opacity uint8,
) {
	if len(component.Indices) > 0 && len(component.Palette) > 0 {
		drawIndexedCompositeComponent(output, component, destination, opacity)

		return
	}

	if component.Image == nil {
		return
	}

	drawImageCompositeComponent(output, component.Image, destination, opacity)
}

// drawIndexedCompositeComponent treats palette index zero and out-of-range
// indices as transparent, matching the source codec's established behavior.
func drawIndexedCompositeComponent(
	output *image.RGBA,
	component CompositeFrame,
	destination image.Point,
	opacity uint8,
) {
	width, height := component.Bounds.Dx(), component.Bounds.Dy()
	for y := 0; y < height; y++ {
		row := y * width
		for x := 0; x < width && row+x < len(component.Indices); x++ {
			index := int(component.Indices[row+x])
			if index == 0 || index >= len(component.Palette) {
				continue
			}

			value := color.RGBAModel.Convert(component.Palette[index]).(color.RGBA)
			blendRGBA(output, destination.X+x, destination.Y+y, value, opacity)
		}
	}
}

// drawImageCompositeComponent uses the standard draw path for decoded images,
// adding a uniform mask only when the COF layer requires partial opacity.
func drawImageCompositeComponent(
	output *image.RGBA,
	source image.Image,
	destination image.Point,
	opacity uint8,
) {
	destinationBounds := source.Bounds().Add(destination)
	if opacity == 255 {
		draw.Draw(output, destinationBounds, source, source.Bounds().Min, draw.Over)

		return
	}

	mask := image.NewUniform(color.Alpha{A: opacity})
	draw.DrawMask(
		output,
		destinationBounds,
		source,
		source.Bounds().Min,
		mask,
		image.Point{},
		draw.Over,
	)
}
