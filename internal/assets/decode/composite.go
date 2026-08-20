package assetdecode

import (
	"errors"
	"fmt"
	"image"
	"image/color"

	cof "github.com/gravestench/cof"
)

// CompositeFrame is one decoded DCC component frame plus the COF layer rules
// governing its position, transparency, priority, and projected shadow.
// Renderer-neutral ownership keeps game and roster composition identical.
type CompositeFrame struct {
	Image   image.Image
	Indices []byte
	Palette color.Palette
	Bounds  image.Rectangle
	Layer   cof.CofLayer
}

// DCCDirectionForCOF translates a spatial COF priority direction into the
// separately interleaved direction order stored by DCC.
func DCCDirectionForCOF(direction, count int) (int, error) {
	lookups := map[int][]int{
		8:  {4, 0, 5, 1, 6, 2, 7, 3},
		16: {4, 8, 0, 9, 5, 10, 1, 11, 6, 12, 2, 13, 7, 14, 3, 15},
	}

	lookup, ok := lookups[count]
	if !ok {
		if direction < 0 || direction >= count {
			return 0, fmt.Errorf("COF direction %d is out of range for %d directions", direction, count)
		}

		return direction, nil
	}

	if direction < 0 || direction >= len(lookup) {
		return 0, fmt.Errorf("COF direction %d is out of range for %d directions", direction, count)
	}

	return lookup[direction], nil
}

// ComposeCOFFrame draws one complete character frame on a stable animation
// canvas. Shadows precede visible layers to match the production renderer's
// ordering and prevent projected pixels from covering foreground components.
func ComposeCOFFrame(
	asset *cof.COF,
	direction int,
	frame int,
	components map[cof.CompositeType]CompositeFrame,
	animationBounds ...image.Rectangle,
) (image.Image, image.Rectangle, error) {
	priority, err := compositionPriority(asset, direction, frame)
	if err != nil {
		return nil, image.Rectangle{}, err
	}

	bounds := compositionBounds(components, animationBounds)
	if bounds.Empty() {
		return nil, image.Rectangle{}, errors.New("COF composition has no component frames")
	}

	canvas := shadowCanvasBounds(bounds, components)
	output := image.NewRGBA(image.Rect(0, 0, canvas.Dx(), canvas.Dy()))
	shadow := compositeShadowMask(bounds, priority, components)
	drawCompositeShadow(output, shadow, bounds, canvas, 96)

	for _, componentType := range priority {
		component, ok := components[componentType]
		if !ok {
			continue
		}

		destination := component.Bounds.Min.Sub(canvas.Min)
		drawCompositeComponent(output, component, destination, compositeOpacity(component.Layer))
	}

	return output, canvas, nil
}

// compositionPriority validates both COF axes before indexing nested slices,
// keeping malformed assets on the error path instead of allowing a panic.
func compositionPriority(asset *cof.COF, direction, frame int) ([]cof.CompositeType, error) {
	if asset == nil {
		return nil, errors.New("COF composition has no asset")
	}

	if direction < 0 || direction >= len(asset.Priority) {
		return nil, fmt.Errorf("COF direction %d is out of range", direction)
	}

	if frame < 0 || frame >= len(asset.Priority[direction]) {
		return nil, fmt.Errorf("COF frame %d is out of range", frame)
	}

	return asset.Priority[direction][frame], nil
}

// compositionBounds honors an explicit animation canvas when present; otherwise
// it derives a stable union that contains every component for the current frame.
func compositionBounds(
	components map[cof.CompositeType]CompositeFrame,
	animationBounds []image.Rectangle,
) image.Rectangle {
	if len(animationBounds) > 0 {
		return animationBounds[0]
	}

	var bounds image.Rectangle
	for _, component := range components {
		bounds = union(bounds, component.Bounds)
	}

	return bounds
}

// union treats an empty accumulator as unset, avoiding image.Rectangle.Union's
// origin bias while building bounds from authored component coordinates.
func union(current, next image.Rectangle) image.Rectangle {
	if current.Empty() {
		return next
	}

	return current.Union(next)
}

// compositeOpacity maps legacy draw effects to the renderer's established alpha
// levels, retaining the default used by unknown transparent effects.
func compositeOpacity(layer cof.CofLayer) uint8 {
	if !layer.Transparent {
		return 255
	}

	switch layer.DrawEffect {
	case cof.DrawEffect(0):
		return 191
	case cof.DrawEffect(1):
		return 128
	case cof.DrawEffect(2):
		return 64
	default:
		return 128
	}
}
