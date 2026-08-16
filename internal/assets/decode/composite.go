package assetdecode

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	cof "github.com/gravestench/cof"
)

// CompositeFrame is one decoded DCC component frame plus the COF layer rules
// that control its position, transparency, priority, and projected shadow.
// Keeping this renderer-neutral makes the game client and Realm roster images
// use exactly the same composition algorithm.
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
// canvas. Shadows are assembled and projected before visible layers, matching
// the production game renderer rather than approximating its roster art.
func ComposeCOFFrame(
	asset *cof.COF,
	direction int,
	frame int,
	components map[cof.CompositeType]CompositeFrame,
	animationBounds ...image.Rectangle,
) (image.Image, image.Rectangle, error) {
	if asset == nil {
		return nil, image.Rectangle{}, errors.New("COF composition has no asset")
	}
	if direction < 0 || direction >= len(asset.Priority) {
		return nil, image.Rectangle{}, fmt.Errorf("COF direction %d is out of range", direction)
	}
	if frame < 0 || frame >= len(asset.Priority[direction]) {
		return nil, image.Rectangle{}, fmt.Errorf("COF frame %d is out of range", frame)
	}
	var bounds image.Rectangle
	if len(animationBounds) > 0 {
		bounds = animationBounds[0]
	} else {
		for _, component := range components {
			bounds = union(bounds, component.Bounds)
		}
	}
	if bounds.Empty() {
		return nil, image.Rectangle{}, errors.New("COF composition has no component frames")
	}
	canvas := shadowCanvasBounds(bounds, components)
	output := image.NewRGBA(image.Rect(0, 0, canvas.Dx(), canvas.Dy()))
	priority := asset.Priority[direction][frame]
	shadow := compositeShadowMask(bounds, priority, components)
	drawCompositeShadow(output, shadow, bounds, canvas, 96)
	for _, componentType := range priority {
		component, ok := components[componentType]
		if !ok {
			continue
		}
		destination := component.Bounds.Min.Sub(canvas.Min)
		alpha := uint8(255)
		if component.Layer.Transparent {
			switch component.Layer.DrawEffect {
			case cof.DrawEffect(0):
				alpha = 191
			case cof.DrawEffect(1):
				alpha = 128
			case cof.DrawEffect(2):
				alpha = 64
			default:
				alpha = 128
			}
		}
		drawCompositeComponent(output, component, destination, alpha)
	}
	return output, canvas, nil
}

func union(current, next image.Rectangle) image.Rectangle {
	if current.Empty() {
		return next
	}
	return current.Union(next)
}

func blendRGBA(destination *image.RGBA, x, y int, source color.RGBA, opacity uint8) {
	if !image.Pt(x, y).In(destination.Bounds()) || source.A == 0 || opacity == 0 {
		return
	}
	alpha := uint32(source.A) * uint32(opacity) / 255
	offset := destination.PixOffset(x, y)
	if alpha == 255 {
		destination.Pix[offset], destination.Pix[offset+1] = source.R, source.G
		destination.Pix[offset+2], destination.Pix[offset+3] = source.B, 255
		return
	}
	inverse := 255 - alpha
	destination.Pix[offset] = uint8((uint32(source.R)*alpha + uint32(destination.Pix[offset])*inverse) / 255)
	destination.Pix[offset+1] = uint8((uint32(source.G)*alpha + uint32(destination.Pix[offset+1])*inverse) / 255)
	destination.Pix[offset+2] = uint8((uint32(source.B)*alpha + uint32(destination.Pix[offset+2])*inverse) / 255)
	destination.Pix[offset+3] = uint8(alpha + uint32(destination.Pix[offset+3])*inverse/255)
}

func drawCompositeComponent(output *image.RGBA, component CompositeFrame, destination image.Point, opacity uint8) {
	if len(component.Indices) > 0 && len(component.Palette) > 0 {
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
		return
	}
	if component.Image == nil {
		return
	}
	if opacity == 255 {
		draw.Draw(output, component.Image.Bounds().Add(destination), component.Image, component.Image.Bounds().Min, draw.Over)
		return
	}
	mask := image.NewUniform(color.Alpha{A: opacity})
	draw.DrawMask(output, component.Image.Bounds().Add(destination), component.Image, component.Image.Bounds().Min, mask, image.Point{}, draw.Over)
}

func projectedShadowBounds(bounds image.Rectangle) image.Rectangle {
	if bounds.Empty() {
		return image.Rectangle{}
	}
	shift := bounds.Dy() / 2
	baseline := bounds.Max.Y - 1
	return image.Rect(bounds.Min.X-shift, baseline-shift, bounds.Max.X, baseline+1)
}

func shadowCanvasBounds(visible image.Rectangle, components map[cof.CompositeType]CompositeFrame) image.Rectangle {
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
		width, height := component.Bounds.Dx(), component.Bounds.Dy()
		origin := component.Bounds.Min.Sub(bounds.Min)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var alpha uint8
				if len(component.Indices) > 0 && y*width+x < len(component.Indices) {
					index := int(component.Indices[y*width+x])
					if index > 0 && index < len(component.Palette) {
						alpha = color.RGBAModel.Convert(component.Palette[index]).(color.RGBA).A
					}
				} else if component.Image != nil {
					point := component.Image.Bounds().Min.Add(image.Pt(x, y))
					_, _, _, value := component.Image.At(point.X, point.Y).RGBA()
					alpha = uint8(value >> 8)
				}
				if alpha > mask.RGBAAt(origin.X+x, origin.Y+y).A {
					mask.SetRGBA(origin.X+x, origin.Y+y, color.RGBA{A: alpha})
				}
			}
		}
	}
	return mask
}

func drawCompositeShadow(output *image.RGBA, mask *image.RGBA, bounds, canvas image.Rectangle, opacity uint8) {
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
			blendRGBA(output, absoluteX-canvas.Min.X, absoluteY-canvas.Min.Y, color.RGBA{A: alpha}, opacity)
		}
	}
}
