package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravestench/dc6"
	xdraw "golang.org/x/image/draw"
)

const compositionGrid = 6

type componentSpec struct {
	Name       string
	Width      int
	Height     int
	Variants   int
	RepeatAxis string
}

// Every production component is authored directly at this final native size.
// tools/ds1-editor/assets/source/composition-kit.png remains a style reference;
// it is intentionally never sampled into the DC6 output.
var componentSpecs = []componentSpec{
	{Name: "panel_top_left", Width: 32, Height: 32},
	{Name: "panel_top", Width: 32, Height: 16, Variants: 4, RepeatAxis: "x"},
	{Name: "panel_top_right", Width: 32, Height: 32},
	{Name: "panel_left", Width: 16, Height: 32, Variants: 4, RepeatAxis: "y"},
	{Name: "panel_fill", Width: 32, Height: 32, Variants: 6, RepeatAxis: "xy"},
	{Name: "panel_right", Width: 16, Height: 32, Variants: 4, RepeatAxis: "y"},
	{Name: "panel_bottom_left", Width: 32, Height: 32},
	{Name: "panel_bottom", Width: 32, Height: 16, Variants: 4, RepeatAxis: "x"},
	{Name: "panel_bottom_right", Width: 32, Height: 32},
	{Name: "section_left", Width: 24, Height: 32},
	{Name: "section_center", Width: 16, Height: 32, Variants: 4, RepeatAxis: "x"},
	{Name: "section_right", Width: 24, Height: 32},
	{Name: "icon_idle", Width: 48, Height: 48},
	{Name: "icon_hover", Width: 48, Height: 48},
	{Name: "icon_pressed", Width: 48, Height: 48},
	{Name: "icon_selected", Width: 48, Height: 48},
	{Name: "icon_disabled", Width: 48, Height: 48},
	{Name: "tool_idle", Width: 44, Height: 44},
	{Name: "button_idle_left", Width: 24, Height: 40},
	{Name: "button_idle_center", Width: 16, Height: 40, Variants: 4, RepeatAxis: "x"},
	{Name: "button_idle_right", Width: 24, Height: 40},
	{Name: "button_hover_left", Width: 24, Height: 40},
	{Name: "button_hover_center", Width: 16, Height: 40, Variants: 4, RepeatAxis: "x"},
	{Name: "button_hover_right", Width: 24, Height: 40},
	{Name: "button_pressed_left", Width: 24, Height: 40},
	{Name: "button_pressed_center", Width: 16, Height: 40, Variants: 4, RepeatAxis: "x"},
	{Name: "button_pressed_right", Width: 24, Height: 40},
	{Name: "tab_left", Width: 24, Height: 36},
	{Name: "tab_center", Width: 16, Height: 36, Variants: 4, RepeatAxis: "x"},
	{Name: "tab_right", Width: 24, Height: 36},
	{Name: "well_left", Width: 16, Height: 28},
	{Name: "well_center", Width: 16, Height: 28, Variants: 5, RepeatAxis: "x"},
	{Name: "well_right", Width: 16, Height: 28},
	{Name: "divider", Width: 12, Height: 32, Variants: 3, RepeatAxis: "y"},
	{Name: "ornament", Width: 48, Height: 16},
	{Name: "recess_top_left", Width: 16, Height: 16},
	{Name: "recess_top", Width: 16, Height: 16, Variants: 4, RepeatAxis: "x"},
	{Name: "recess_top_right", Width: 16, Height: 16},
	{Name: "recess_left", Width: 16, Height: 16, Variants: 4, RepeatAxis: "y"},
	{Name: "recess_fill", Width: 16, Height: 16, Variants: 6, RepeatAxis: "xy"},
	{Name: "recess_right", Width: 16, Height: 16, Variants: 4, RepeatAxis: "y"},
	{Name: "recess_bottom_left", Width: 16, Height: 16},
	{Name: "recess_bottom", Width: 16, Height: 16, Variants: 4, RepeatAxis: "x"},
	{Name: "recess_bottom_right", Width: 16, Height: 16},
	{Name: "dropdown_idle_left", Width: 24, Height: 40},
	{Name: "dropdown_idle_center", Width: 16, Height: 40, Variants: 4, RepeatAxis: "x"},
	{Name: "dropdown_idle_right", Width: 32, Height: 40},
	{Name: "dropdown_hover_left", Width: 24, Height: 40},
	{Name: "dropdown_hover_center", Width: 16, Height: 40, Variants: 4, RepeatAxis: "x"},
	{Name: "dropdown_hover_right", Width: 32, Height: 40},
	{Name: "checkbox_off", Width: 24, Height: 24},
	{Name: "checkbox_on", Width: 24, Height: 24},
	{Name: "checkbox_hover", Width: 24, Height: 24},
	{Name: "radio_off", Width: 24, Height: 24},
	{Name: "radio_on", Width: 24, Height: 24},
	{Name: "radio_hover", Width: 24, Height: 24},
}

var nativeColors = struct {
	Transparent color.NRGBA
	Ink         color.NRGBA
	Stone       color.NRGBA
	StoneLight  color.NRGBA
	StoneDark   color.NRGBA
	Iron        color.NRGBA
	Steel       color.NRGBA
	Highlight   color.NRGBA
	GoldDark    color.NRGBA
	Gold        color.NRGBA
	GoldLight   color.NRGBA
	TealDark    color.NRGBA
	Teal        color.NRGBA
	Red         color.NRGBA
}{
	Transparent: color.NRGBA{},
	// Neutral values deliberately target the palette's grayscale ramp. A low
	// blue component such as {0, 0, 51} is not "near black" after indexing; it
	// is an exact saturated navy and turns large panels into blue carpet.
	Ink:        color.NRGBA{R: 7, G: 7, B: 7, A: 255},
	StoneDark:  color.NRGBA{R: 20, G: 20, B: 20, A: 255},
	Stone:      color.NRGBA{R: 34, G: 34, B: 34, A: 255},
	StoneLight: color.NRGBA{R: 47, G: 47, B: 47, A: 255},
	Iron:       color.NRGBA{R: 60, G: 60, B: 60, A: 255},
	Steel:      color.NRGBA{R: 87, G: 87, B: 87, A: 255},
	Highlight:  color.NRGBA{R: 121, G: 121, B: 121, A: 255},
	GoldDark:   color.NRGBA{R: 102, G: 51, B: 0, A: 255},
	Gold:       color.NRGBA{R: 153, G: 102, B: 0, A: 255},
	GoldLight:  color.NRGBA{R: 204, G: 153, B: 0, A: 255},
	TealDark:   color.NRGBA{R: 0, G: 51, B: 51, A: 255},
	Teal:       color.NRGBA{R: 0, G: 153, B: 153, A: 255},
	Red:        color.NRGBA{R: 153, G: 0, B: 0, A: 255},
}

// buildCompositionSheet emits native-grid sprites and records compatible
// variant lists for the runtime's deterministic tiler.
func buildCompositionSheet(sourceRoot, outputRoot string, palette color.Palette) (sheetManifest, error) {
	frameCount := 0
	for _, spec := range componentSpecs {
		frameCount += componentVariantCount(spec)
	}
	animation, err := dc6.New(1, frameCount)
	if err != nil {
		return sheetManifest{}, err
	}

	manifest := sheetManifest{
		Path:   "darkmagic/ds1-editor/ui/composition.dc6",
		Frames: make(map[string]interface{}, len(componentSpecs)),
		Sizes:  make(map[string][2]int, len(componentSpecs)),
	}
	previewFrames := make([]image.Image, 0, len(componentSpecs))
	frameIndex := 0
	for _, spec := range componentSpecs {
		count := componentVariantCount(spec)
		indices := make([]int, 0, count)
		for variant := 0; variant < count; variant++ {
			frame := drawNativeComponent(spec, variant)
			if variant == 0 {
				previewFrames = append(previewFrames, frame)
			}
			encoded := animation.Directions[0].Frames[frameIndex]
			encoded.Width, encoded.Height = uint32(spec.Width), uint32(spec.Height)
			encoded.IndexData = quantize(frame, palette)
			indices = append(indices, frameIndex)
			frameIndex++
		}
		manifest.Frames[spec.Name] = indices
		manifest.Sizes[spec.Name] = [2]int{spec.Width, spec.Height}
	}

	data, err := dc6.Encode(animation)
	if err != nil {
		return sheetManifest{}, fmt.Errorf("encode composition.dc6: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputRoot, "composition.dc6"), data, 0o644); err != nil {
		return sheetManifest{}, err
	}
	previewPath := filepath.Join(filepath.Dir(sourceRoot), "previews", "composition.png")
	if err := writeCompositionPreview(previewPath, previewFrames); err != nil {
		return sheetManifest{}, err
	}
	return manifest, nil
}

func drawNativeComponent(spec componentSpec, variant int) *image.NRGBA {
	target := image.NewNRGBA(image.Rect(0, 0, spec.Width, spec.Height))
	name := spec.Name
	switch {
	case strings.HasPrefix(name, "panel_"):
		drawPanelComponent(target, name, variant)
	case strings.HasPrefix(name, "section_"):
		drawSectionComponent(target, name, variant)
	case strings.HasPrefix(name, "icon_") || name == "tool_idle":
		drawIconPlate(target, name, variant)
	case strings.HasPrefix(name, "button_"):
		drawButtonComponent(target, name, variant)
	case strings.HasPrefix(name, "tab_"):
		drawTabComponent(target, name, variant)
	case strings.HasPrefix(name, "well_"):
		drawWellComponent(target, name, variant)
	case strings.HasPrefix(name, "recess_"):
		drawRecessComponent(target, name, variant)
	case strings.HasPrefix(name, "dropdown_"):
		drawDropdownComponent(target, name, variant)
	case strings.HasPrefix(name, "checkbox_"):
		drawCheckboxComponent(target, name, variant)
	case strings.HasPrefix(name, "radio_"):
		drawRadioComponent(target, name, variant)
	case name == "divider":
		drawDivider(target, variant)
	case name == "ornament":
		drawOrnament(target)
	}
	stabilizeComponentEdges(target, spec.RepeatAxis)
	return target
}

func componentVariantCount(spec componentSpec) int {
	if spec.Variants > 0 {
		return spec.Variants
	}
	return 3
}

func drawPanelComponent(target *image.NRGBA, name string, variant int) {
	bounds := target.Bounds()
	if name == "panel_fill" {
		fillStone(target, variant)
		return
	}
	switch name {
	case "panel_top":
		drawHorizontalRail(target, true, variant, false)
	case "panel_bottom":
		drawHorizontalRail(target, false, variant, false)
	case "panel_left":
		drawVerticalRail(target, true, variant, false)
	case "panel_right":
		drawVerticalRail(target, false, variant, false)
	case "panel_top_left", "panel_top_right", "panel_bottom_left", "panel_bottom_right":
		top := strings.Contains(name, "top")
		left := strings.HasSuffix(name, "left")
		drawCornerRails(target, top, left)
		centerX, centerY := 8, 8
		if !left {
			centerX = bounds.Max.X - 9
		}
		if !top {
			centerY = bounds.Max.Y - 9
		}
		drawRivet(target, centerX, centerY, nativeColors.Gold)
		drawCornerCurl(target, centerX, centerY, top, left)
	}
}

func fillStone(target *image.NRGBA, variant int) {
	bounds := target.Bounds()
	draw.Draw(target, bounds, &image.Uniform{C: nativeColors.StoneDark}, image.Point{}, draw.Src)
	// Repetition variants are accents, not noise. At most one shallow nick is
	// authored into a sufficiently large tile; most variants remain quiet.
	if bounds.Dx() >= 24 && bounds.Dy() >= 24 {
		x := bounds.Min.X + 7 + (variant*7)%(bounds.Dx()-14)
		y := bounds.Min.Y + 7 + (variant*11)%(bounds.Dy()-14)
		length := 2 + variant%3
		for step := 0; step < length; step++ {
			setIfInside(target, x+step, y+step/2, nativeColors.Stone)
		}
		if variant%2 == 0 {
			setIfInside(target, x+length-1, y+2, nativeColors.Ink)
		}
	}
}

func drawHorizontalRail(target *image.NRGBA, top bool, variant int, teal bool) {
	bounds := target.Bounds()
	thickness := min(16, bounds.Dy())
	start := bounds.Min.Y
	if !top {
		start = bounds.Max.Y - thickness
	}
	draw.Draw(target, image.Rect(bounds.Min.X, start, bounds.Max.X, start+thickness),
		&image.Uniform{C: nativeColors.StoneDark}, image.Point{}, draw.Src)
	outer := start
	inner := start + thickness - 1
	if !top {
		outer, inner = inner, outer
	}
	direction := 1
	if !top {
		direction = -1
	}
	for offset, shade := range []color.NRGBA{nativeColors.Ink, nativeColors.Highlight, nativeColors.Steel} {
		y := outer + offset*direction
		line(target, bounds.Min.X, y, bounds.Max.X-1, y, shade)
	}
	line(target, bounds.Min.X, inner, bounds.Max.X-1, inner, nativeColors.Ink)
	line(target, bounds.Min.X, inner-direction, bounds.Max.X-1, inner-direction, nativeColors.GoldDark)
	line(target, bounds.Min.X, inner-2*direction, bounds.Max.X-1, inner-2*direction, nativeColors.Iron)
	drawHorizontalBraid(target, start, thickness, variant, teal)
	accent := nativeColors.Gold
	if teal {
		accent = nativeColors.Teal
	}
	x := bounds.Min.X + 5 + (variant*7)%max(8, bounds.Dx()-8)
	y := start + thickness/2
	if variant%2 == 0 {
		drawSmallStud(target, x, y, accent)
	}
}

func drawVerticalRail(target *image.NRGBA, left bool, variant int, teal bool) {
	bounds := target.Bounds()
	thickness := min(16, bounds.Dx())
	start := bounds.Min.X
	if !left {
		start = bounds.Max.X - thickness
	}
	draw.Draw(target, image.Rect(start, bounds.Min.Y, start+thickness, bounds.Max.Y),
		&image.Uniform{C: nativeColors.StoneDark}, image.Point{}, draw.Src)
	outer := start
	inner := start + thickness - 1
	if !left {
		outer, inner = inner, outer
	}
	direction := 1
	if !left {
		direction = -1
	}
	for offset, shade := range []color.NRGBA{nativeColors.Ink, nativeColors.Highlight, nativeColors.Steel} {
		x := outer + offset*direction
		line(target, x, bounds.Min.Y, x, bounds.Max.Y-1, shade)
	}
	line(target, inner, bounds.Min.Y, inner, bounds.Max.Y-1, nativeColors.Ink)
	line(target, inner-direction, bounds.Min.Y, inner-direction, bounds.Max.Y-1, nativeColors.GoldDark)
	line(target, inner-2*direction, bounds.Min.Y, inner-2*direction, bounds.Max.Y-1, nativeColors.Iron)
	drawVerticalBraid(target, start, thickness, variant, teal)
	accent := nativeColors.Gold
	if teal {
		accent = nativeColors.Teal
	}
	y := bounds.Min.Y + 5 + (variant*7)%max(8, bounds.Dy()-8)
	x := start + thickness/2
	if variant%2 == 0 {
		drawSmallStud(target, x, y, accent)
	}
}

func drawHorizontalBraid(target *image.NRGBA, start, thickness, variant int, teal bool) {
	accent := nativeColors.Steel
	if teal {
		accent = nativeColors.TealDark
	}
	center := start + thickness/2
	for x := target.Bounds().Min.X; x < target.Bounds().Max.X; x++ {
		phase := (x + variant*2) % 8
		offset := phase
		if phase > 4 {
			offset = 8 - phase
		}
		offset -= 2
		setIfInside(target, x, center+offset, accent)
		setIfInside(target, x, center-offset, nativeColors.Iron)
	}
}

func drawVerticalBraid(target *image.NRGBA, start, thickness, variant int, teal bool) {
	accent := nativeColors.Steel
	if teal {
		accent = nativeColors.TealDark
	}
	center := start + thickness/2
	for y := target.Bounds().Min.Y; y < target.Bounds().Max.Y; y++ {
		phase := (y + variant*2) % 8
		offset := phase
		if phase > 4 {
			offset = 8 - phase
		}
		offset -= 2
		setIfInside(target, center+offset, y, accent)
		setIfInside(target, center-offset, y, nativeColors.Iron)
	}
}

func drawSmallStud(target *image.NRGBA, centerX, centerY int, accent color.NRGBA) {
	setIfInside(target, centerX, centerY-1, nativeColors.GoldDark)
	setIfInside(target, centerX-1, centerY, nativeColors.GoldDark)
	setIfInside(target, centerX, centerY, accent)
	setIfInside(target, centerX+1, centerY, nativeColors.GoldLight)
	setIfInside(target, centerX, centerY+1, nativeColors.GoldDark)
}

func drawCornerRails(target *image.NRGBA, top, left bool) {
	drawHorizontalRail(target, top, 0, false)
	drawVerticalRail(target, left, 0, false)
}

func drawCornerCurl(target *image.NRGBA, centerX, centerY int, top, left bool) {
	directionX, directionY := 1, 1
	if !left {
		directionX = -1
	}
	if !top {
		directionY = -1
	}
	points := [][2]int{{3, 0}, {4, 1}, {5, 2}, {5, 3}, {4, 4}, {3, 4}}
	for _, point := range points {
		x := centerX + point[0]*directionX
		y := centerY + point[1]*directionY
		setIfInside(target, x, y, nativeColors.Steel)
	}
}

func drawSectionComponent(target *image.NRGBA, name string, variant int) {
	fillStone(target, variant+7)
	bounds := target.Bounds()
	for offset, shade := range []color.NRGBA{
		nativeColors.Ink, nativeColors.Steel, nativeColors.GoldDark,
	} {
		line(target, bounds.Min.X, bounds.Min.Y+offset, bounds.Max.X-1, bounds.Min.Y+offset, shade)
		line(target, bounds.Min.X, bounds.Max.Y-1-offset, bounds.Max.X-1, bounds.Max.Y-1-offset, shade)
	}
	if strings.HasSuffix(name, "left") {
		drawPointedCap(target, true, nativeColors.Gold)
	} else if strings.HasSuffix(name, "right") {
		drawPointedCap(target, false, nativeColors.Gold)
	}
}

func drawPointedCap(target *image.NRGBA, left bool, accent color.NRGBA) {
	bounds := target.Bounds()
	centerY := bounds.Min.Y + bounds.Dy()/2
	centerX := bounds.Min.X + 7
	if !left {
		centerX = bounds.Max.X - 8
	}
	drawRivet(target, centerX, centerY, accent)
	for distance := 0; distance < 5; distance++ {
		x := centerX - 6 - distance
		if !left {
			x = centerX + 6 + distance
		}
		setIfInside(target, x, centerY-distance/2, nativeColors.Steel)
		setIfInside(target, x, centerY+distance/2, nativeColors.Steel)
	}
}

func drawIconPlate(target *image.NRGBA, name string, variant int) {
	bounds := target.Bounds()
	fillStone(target, variant+11)
	accent := nativeColors.Gold
	pressed := name == "icon_pressed"
	switch name {
	case "icon_hover", "icon_selected":
		accent = nativeColors.Teal
	case "icon_disabled":
		accent = nativeColors.Steel
	}
	drawChamferedBorder(target, bounds, accent, pressed)
	for _, point := range []image.Point{
		{X: 7, Y: 7}, {X: bounds.Max.X - 8, Y: 7},
		{X: 7, Y: bounds.Max.Y - 8}, {X: bounds.Max.X - 8, Y: bounds.Max.Y - 8},
	} {
		drawRivet(target, point.X, point.Y, accent)
	}
}

func drawChamferedBorder(target *image.NRGBA, bounds image.Rectangle, accent color.NRGBA, pressed bool) {
	for inset := 0; inset < 4; inset++ {
		shade := []color.NRGBA{nativeColors.Ink, nativeColors.Highlight, nativeColors.Steel, accent}[inset]
		if pressed && inset == 1 {
			shade = nativeColors.Iron
		}
		left, top := bounds.Min.X+inset, bounds.Min.Y+inset
		right, bottom := bounds.Max.X-1-inset, bounds.Max.Y-1-inset
		line(target, left+4, top, right-4, top, shade)
		line(target, left+4, bottom, right-4, bottom, shade)
		line(target, left, top+4, left, bottom-4, shade)
		line(target, right, top+4, right, bottom-4, shade)
		line(target, left+1, top+3, left+3, top+1, shade)
		line(target, right-3, top+1, right-1, top+3, shade)
		line(target, left+1, bottom-3, left+3, bottom-1, shade)
		line(target, right-3, bottom-1, right-1, bottom-3, shade)
	}
}

func drawButtonComponent(target *image.NRGBA, name string, variant int) {
	state := "idle"
	if strings.Contains(name, "hover") {
		state = "hover"
	} else if strings.Contains(name, "pressed") {
		state = "pressed"
	}
	drawControlField(target, state, variant)
	if strings.HasSuffix(name, "left") {
		drawControlCap(target, true, state)
	} else if strings.HasSuffix(name, "right") {
		drawControlCap(target, false, state)
	}
}

func drawControlField(target *image.NRGBA, state string, variant int) {
	fillStone(target, variant+19)
	accent := nativeColors.GoldDark
	if state == "hover" {
		accent = nativeColors.Teal
	}
	drawHorizontalRail(target, true, variant, state == "hover")
	drawHorizontalRail(target, false, variant, state == "hover")
	if state == "pressed" {
		bounds := target.Bounds()
		draw.Draw(target, image.Rect(bounds.Min.X, bounds.Min.Y+7, bounds.Max.X, bounds.Max.Y-7),
			&image.Uniform{C: nativeColors.Ink}, image.Point{}, draw.Src)
		line(target, bounds.Min.X, bounds.Min.Y+7, bounds.Max.X-1, bounds.Min.Y+7, accent)
	}
}

func drawControlCap(target *image.NRGBA, left bool, state string) {
	bounds := target.Bounds()
	accent := nativeColors.Gold
	if state == "hover" {
		accent = nativeColors.Teal
	}
	x := bounds.Min.X + 7
	if !left {
		x = bounds.Max.X - 8
	}
	drawRivet(target, x, bounds.Min.Y+bounds.Dy()/2, accent)
	drawVerticalRail(target, left, 0, state == "hover")
}

func drawTabComponent(target *image.NRGBA, name string, variant int) {
	fillStone(target, variant+23)
	drawHorizontalRail(target, true, variant, true)
	drawHorizontalRail(target, false, variant, false)
	if strings.HasSuffix(name, "left") {
		drawControlCap(target, true, "hover")
	} else if strings.HasSuffix(name, "right") {
		drawControlCap(target, false, "hover")
	}
}

func drawWellComponent(target *image.NRGBA, name string, variant int) {
	bounds := target.Bounds()
	draw.Draw(target, bounds, &image.Uniform{C: nativeColors.Ink}, image.Point{}, draw.Src)
	line(target, bounds.Min.X, bounds.Min.Y, bounds.Max.X-1, bounds.Min.Y, nativeColors.Iron)
	line(target, bounds.Min.X, bounds.Min.Y+1, bounds.Max.X-1, bounds.Min.Y+1, nativeColors.Stone)
	line(target, bounds.Min.X, bounds.Max.Y-2, bounds.Max.X-1, bounds.Max.Y-2, nativeColors.Steel)
	line(target, bounds.Min.X, bounds.Max.Y-1, bounds.Max.X-1, bounds.Max.Y-1, nativeColors.Ink)
	if variant%3 == 1 {
		target.SetNRGBA(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+4, nativeColors.StoneLight)
	}
	if strings.HasSuffix(name, "left") {
		drawVerticalRail(target, true, variant, false)
	} else if strings.HasSuffix(name, "right") {
		drawVerticalRail(target, false, variant, false)
	}
}

func drawRecessComponent(target *image.NRGBA, name string, variant int) {
	bounds := target.Bounds()
	draw.Draw(target, bounds, &image.Uniform{C: nativeColors.Ink}, image.Point{}, draw.Src)
	if strings.Contains(name, "fill") && variant == 4 {
		// A single optional scuff gives deterministic variation without a tiled
		// dot field competing with DT1 thumbnails.
		target.SetNRGBA(bounds.Min.X+5, bounds.Min.Y+10, nativeColors.StoneDark)
	}
	if strings.Contains(name, "top") {
		line(target, bounds.Min.X, bounds.Min.Y, bounds.Max.X-1, bounds.Min.Y, nativeColors.Ink)
		line(target, bounds.Min.X, bounds.Min.Y+1, bounds.Max.X-1, bounds.Min.Y+1, nativeColors.Iron)
		line(target, bounds.Min.X, bounds.Min.Y+2, bounds.Max.X-1, bounds.Min.Y+2, nativeColors.GoldDark)
	}
	if strings.Contains(name, "bottom") {
		line(target, bounds.Min.X, bounds.Max.Y-3, bounds.Max.X-1, bounds.Max.Y-3, nativeColors.Steel)
		line(target, bounds.Min.X, bounds.Max.Y-2, bounds.Max.X-1, bounds.Max.Y-2, nativeColors.Highlight)
		line(target, bounds.Min.X, bounds.Max.Y-1, bounds.Max.X-1, bounds.Max.Y-1, nativeColors.Ink)
	}
	if strings.HasSuffix(name, "left") {
		line(target, bounds.Min.X, bounds.Min.Y, bounds.Min.X, bounds.Max.Y-1, nativeColors.Ink)
		line(target, bounds.Min.X+1, bounds.Min.Y, bounds.Min.X+1, bounds.Max.Y-1, nativeColors.Iron)
		line(target, bounds.Min.X+2, bounds.Min.Y, bounds.Min.X+2, bounds.Max.Y-1, nativeColors.GoldDark)
	}
	if strings.HasSuffix(name, "right") {
		line(target, bounds.Max.X-3, bounds.Min.Y, bounds.Max.X-3, bounds.Max.Y-1, nativeColors.Steel)
		line(target, bounds.Max.X-2, bounds.Min.Y, bounds.Max.X-2, bounds.Max.Y-1, nativeColors.Highlight)
		line(target, bounds.Max.X-1, bounds.Min.Y, bounds.Max.X-1, bounds.Max.Y-1, nativeColors.Ink)
	}
}

func drawDropdownComponent(target *image.NRGBA, name string, variant int) {
	state := "idle"
	if strings.Contains(name, "hover") {
		state = "hover"
	}
	drawControlField(target, state, variant)
	if strings.HasSuffix(name, "left") {
		drawControlCap(target, true, state)
		return
	}
	if !strings.HasSuffix(name, "right") {
		return
	}
	drawControlCap(target, false, state)
	bounds := target.Bounds()
	centerX, centerY := bounds.Min.X+11, bounds.Min.Y+bounds.Dy()/2
	accent := nativeColors.Gold
	if state == "hover" {
		accent = nativeColors.Teal
	}
	for row := 0; row < 4; row++ {
		line(target, centerX-row, centerY-2+row, centerX+row, centerY-2+row, accent)
	}
}

func drawCheckboxComponent(target *image.NRGBA, name string, variant int) {
	fillStone(target, variant+31)
	bounds := target.Bounds()
	accent := nativeColors.Gold
	if strings.HasSuffix(name, "hover") {
		accent = nativeColors.Teal
	}
	drawChamferedBorder(target, bounds, accent, false)
	if !strings.HasSuffix(name, "on") {
		return
	}
	line(target, 6, 12, 10, 16, nativeColors.GoldDark)
	line(target, 10, 16, 18, 7, nativeColors.GoldLight)
	line(target, 7, 12, 10, 15, nativeColors.Gold)
	line(target, 10, 15, 17, 7, nativeColors.Gold)
}

func drawRadioComponent(target *image.NRGBA, name string, variant int) {
	bounds := target.Bounds()
	draw.Draw(target, bounds, &image.Uniform{C: nativeColors.Transparent}, image.Point{}, draw.Src)
	accent := nativeColors.Gold
	if strings.HasSuffix(name, "hover") {
		accent = nativeColors.Teal
	}
	centerX, centerY := bounds.Dx()/2, bounds.Dy()/2
	for radius, shade := range []color.NRGBA{
		nativeColors.Ink, nativeColors.Highlight, nativeColors.Steel,
		nativeColors.StoneDark, nativeColors.GoldDark,
	} {
		drawDiamondOutline(target, centerX, centerY, 10-radius, shade)
	}
	if strings.HasSuffix(name, "on") {
		drawRivet(target, centerX, centerY, accent)
	} else if variant%2 == 1 {
		setIfInside(target, centerX, centerY, nativeColors.Stone)
	}
}

func drawDiamondOutline(target *image.NRGBA, centerX, centerY, radius int, shade color.NRGBA) {
	for offset := -radius; offset <= radius; offset++ {
		y := radius - absolute(offset)
		setIfInside(target, centerX+offset, centerY-y, shade)
		setIfInside(target, centerX+offset, centerY+y, shade)
	}
}

func drawDivider(target *image.NRGBA, variant int) {
	bounds := target.Bounds()
	center := bounds.Min.X + bounds.Dx()/2
	line(target, center-2, bounds.Min.Y, center-2, bounds.Max.Y-1, nativeColors.Ink)
	line(target, center-1, bounds.Min.Y, center-1, bounds.Max.Y-1, nativeColors.Steel)
	line(target, center, bounds.Min.Y, center, bounds.Max.Y-1, nativeColors.GoldDark)
	line(target, center+1, bounds.Min.Y, center+1, bounds.Max.Y-1, nativeColors.Ink)
	drawRivet(target, center, bounds.Min.Y+8+(variant*7)%16, nativeColors.Gold)
}

func drawOrnament(target *image.NRGBA) {
	bounds := target.Bounds()
	centerX, centerY := bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2
	line(target, bounds.Min.X+2, centerY, bounds.Max.X-3, centerY, nativeColors.Steel)
	line(target, bounds.Min.X+6, centerY-2, bounds.Max.X-7, centerY-2, nativeColors.GoldDark)
	drawRivet(target, centerX, centerY, nativeColors.GoldLight)
	for offset := 1; offset <= 7; offset++ {
		setIfInside(target, centerX-offset, centerY-offset/2, nativeColors.Steel)
		setIfInside(target, centerX+offset, centerY-offset/2, nativeColors.Steel)
	}
}

func drawRivet(target *image.NRGBA, centerX, centerY int, accent color.NRGBA) {
	for _, pixel := range []struct {
		x, y  int
		shade color.NRGBA
	}{
		{-1, -2, nativeColors.GoldDark}, {0, -2, nativeColors.Gold},
		{-2, -1, nativeColors.GoldDark}, {-1, -1, accent}, {0, -1, nativeColors.GoldLight},
		{1, -1, nativeColors.Gold}, {-2, 0, nativeColors.Gold}, {-1, 0, nativeColors.GoldLight},
		{0, 0, accent}, {1, 0, nativeColors.GoldDark}, {-1, 1, nativeColors.Gold},
		{0, 1, nativeColors.GoldDark},
	} {
		setIfInside(target, centerX+pixel.x, centerY+pixel.y, pixel.shade)
	}
}

func line(target *image.NRGBA, x1, y1, x2, y2 int, shade color.NRGBA) {
	dx, dy := absolute(x2-x1), -absolute(y2-y1)
	stepX, stepY := -1, -1
	if x1 < x2 {
		stepX = 1
	}
	if y1 < y2 {
		stepY = 1
	}
	err := dx + dy
	for {
		setIfInside(target, x1, y1, shade)
		if x1 == x2 && y1 == y2 {
			break
		}
		double := 2 * err
		if double >= dy {
			err += dy
			x1 += stepX
		}
		if double <= dx {
			err += dx
			y1 += stepY
		}
	}
}

func setIfInside(target *image.NRGBA, x, y int, shade color.NRGBA) {
	if image.Pt(x, y).In(target.Bounds()) {
		target.SetNRGBA(x, y, shade)
	}
}

func absolute(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// The preview scales finished sprites only for human inspection. Production
// DC6 frames above retain their exact one-to-one authored pixels.
func writeCompositionPreview(path string, frames []image.Image) error {
	const cellSize = 112
	rows := (len(frames) + compositionGrid - 1) / compositionGrid
	preview := image.NewRGBA(image.Rect(0, 0, compositionGrid*cellSize, rows*cellSize))
	draw.Draw(preview, preview.Bounds(), &image.Uniform{C: color.RGBA{R: 18, G: 23, B: 30, A: 255}},
		image.Point{}, draw.Src)
	for index, frame := range frames {
		column, row := index%compositionGrid, index/compositionGrid
		scale := min(4, min((cellSize-16)/frame.Bounds().Dx(), (cellSize-16)/frame.Bounds().Dy()))
		width, height := frame.Bounds().Dx()*scale, frame.Bounds().Dy()*scale
		left := column*cellSize + (cellSize-width)/2
		top := row*cellSize + (cellSize-height)/2
		xdraw.NearestNeighbor.Scale(preview, image.Rect(left, top, left+width, top+height),
			frame, frame.Bounds(), xdraw.Src, nil)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(file, preview); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// stabilizeComponentEdges gives every repeatable frame identical opposing
// boundaries. Variant choice can therefore change per tile without seams.
func stabilizeComponentEdges(target *image.NRGBA, axis string) {
	bounds := target.Bounds()
	if axis == "x" || axis == "xy" {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			target.Set(bounds.Max.X-1, y, target.At(bounds.Min.X, y))
		}
	}
	if axis == "y" || axis == "xy" {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			target.Set(x, bounds.Max.Y-1, target.At(x, bounds.Min.Y))
		}
	}
}
