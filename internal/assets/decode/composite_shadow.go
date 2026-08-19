package assetdecode

import (
	"image"
	"image/color"

	cof "github.com/gravestench/cof"
)

// projectedShadowBounds returns the flattened, left-shifted silhouette extent
// used by the production renderer for a standing component canvas.
func projectedShadowBounds(bounds image.Rectangle) image.Rectangle {
	if bounds.Empty() {
		return image.Rectangle{}
	}

	shift := bounds.Dy() / 2
	baseline := bounds.Max.Y - 1

	return image.Rect(bounds.Min.X-shift, baseline-shift, bounds.Max.X, baseline+1)
}

// shadowCanvasBounds expands visible bounds symmetrically when any component
// casts a shadow, so animation anchors do not move between shadowed frames.
func shadowCanvasBounds(
	visible image.Rectangle,
	components map[cof.CompositeType]CompositeFrame,
) image.Rectangle {
	projected := visible

	for _, component := range components {
		if component.Layer.Shadow != 0 {
			projected = projected.Union(projectedShadowBounds(visible))

			break
		}
	}

	horizontal := max(visible.Min.X-projected.Min.X, projected.Max.X-visible.Max.X, 0)
	vertical := max(visible.Min.Y-projected.Min.Y, projected.Max.Y-visible.Max.Y, 0)

	return image.Rect(
		visible.Min.X-horizontal,
		visible.Min.Y-vertical,
		visible.Max.X+horizontal,
		visible.Max.Y+vertical,
	)
}

// compositeShadowMask merges shadow-casting layers in COF priority order and
// retains the greatest source alpha at each point in the visible bounds.
func compositeShadowMask(
	bounds image.Rectangle,
	priority []cof.CompositeType,
	components map[cof.CompositeType]CompositeFrame,
) *image.RGBA {
	mask := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for _, componentType := range priority {
		component, ok := components[componentType]
		if !ok || component.Layer.Shadow == 0 {
			continue
		}

		mergeComponentShadow(mask, bounds, component)
	}

	return mask
}

// mergeComponentShadow overlays one component's alpha into the shared maximum
// mask, ensuring overlapping layers cannot make a shadow less opaque.
func mergeComponentShadow(mask *image.RGBA, bounds image.Rectangle, component CompositeFrame) {
	width, height := component.Bounds.Dx(), component.Bounds.Dy()
	origin := component.Bounds.Min.Sub(bounds.Min)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := compositePixelAlpha(component, x, y, width)

			point := origin.Add(image.Pt(x, y))
			if alpha > mask.RGBAAt(point.X, point.Y).A {
				mask.SetRGBA(point.X, point.Y, color.RGBA{A: alpha})
			}
		}
	}
}

// compositePixelAlpha preserves the indexed-image preference used for visible
// drawing, while falling back to Image only when indexed data has no pixel.
func compositePixelAlpha(component CompositeFrame, x, y, width int) uint8 {
	pixel := y*width + x
	if len(component.Indices) > 0 && pixel < len(component.Indices) {
		index := int(component.Indices[pixel])
		if index > 0 && index < len(component.Palette) {
			return color.RGBAModel.Convert(component.Palette[index]).(color.RGBA).A
		}

		return 0
	}

	if component.Image == nil {
		return 0
	}

	point := component.Image.Bounds().Min.Add(image.Pt(x, y))
	_, _, _, alpha := component.Image.At(point.X, point.Y).RGBA()

	return uint8(alpha >> 8)
}

// drawCompositeShadow projects mask rows toward the baseline before blending,
// preserving the renderer's fixed two-source-pixels-per-shadow-pixel slope.
func drawCompositeShadow(
	output *image.RGBA,
	mask *image.RGBA,
	bounds image.Rectangle,
	canvas image.Rectangle,
	opacity uint8,
) {
	width, height := mask.Bounds().Dx(), mask.Bounds().Dy()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := mask.RGBAAt(x, y).A
			if alpha == 0 {
				continue
			}

			distance := height - 1 - y
			shift := (distance + 1) / 2
			absoluteX := bounds.Min.X + x - shift
			absoluteY := bounds.Max.Y - 1 - shift
			blendRGBA(
				output,
				absoluteX-canvas.Min.X,
				absoluteY-canvas.Min.Y,
				color.RGBA{A: alpha},
				opacity,
			)
		}
	}
}
