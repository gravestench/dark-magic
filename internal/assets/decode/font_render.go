package assetdecode

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
)

// fontRenderState carries token position and active color across wrapped lines,
// preserving a single visible-rune stream for the complete input string.
type fontRenderState struct {
	runIndex  int
	tint      color.Color
	transform int
}

// Render lays out, wraps, aligns, and composites text into a new CPU image. It
// never creates retained nodes or native textures, keeping ownership explicit and tests headless.
func (f *BitmapFont) Render(
	text string,
	tint color.Color,
	maxWidth int,
	align string,
) (*image.RGBA, error) {
	if err := f.validateForLayout(); err != nil {
		return nil, err
	}

	if align != "left" && align != "center" && align != "right" {
		return nil, fmt.Errorf("bitmap font: invalid alignment %q", align)
	}

	text, runs := parseColorTokens(text)
	lines := f.wrap(text, maxWidth)
	width := f.renderWidth(lines, maxWidth)
	height := maxInt(1, len(lines)*f.LineHeight)
	output := image.NewRGBA(image.Rect(0, 0, width, height))
	state := fontRenderState{tint: tint, transform: -1}

	for lineIndex, line := range lines {
		x := alignedLineStart(width, f.measure(line), align)
		f.drawLine(output, line, lineIndex, x, tint, runs, &state)
	}

	return output, nil
}

// Measure returns the exact texture dimensions Render would allocate without rasterizing any glyphs.
// UI layout can therefore size controls before creating or mutating retained presentation nodes.
func (f *BitmapFont) Measure(text string, maxWidth int) (image.Point, error) {
	if err := f.validateForLayout(); err != nil {
		return image.Point{}, err
	}
	if maxWidth < 0 {
		return image.Point{}, fmt.Errorf("bitmap font: max width cannot be negative")
	}

	visibleText, _ := parseColorTokens(text)
	lines := f.wrap(visibleText, maxWidth)

	return image.Pt(f.renderWidth(lines, maxWidth), maxInt(1, len(lines)*f.LineHeight)), nil
}

// validateForLayout centralizes the minimum invariants shared by measuring and rendering.
func (f *BitmapFont) validateForLayout() error {
	if f == nil || len(f.Glyphs) == 0 || len(f.Frames) == 0 || f.LineHeight <= 0 {
		return fmt.Errorf("bitmap font: empty font")
	}

	return nil
}

// renderWidth preserves the caller's explicit width when supplied; otherwise
// it selects the widest wrapped line and retains a one-pixel minimum canvas.
func (f *BitmapFont) renderWidth(lines []string, maxWidth int) int {
	width := 1
	for _, line := range lines {
		if measured := f.measure(line); measured > width {
			width = measured
		}
	}

	if maxWidth > 0 {
		width = maxWidth
	}

	return width
}

// alignedLineStart computes the per-line origin after validation has limited
// alignment to the three public contract values.
func alignedLineStart(canvasWidth, lineWidth int, align string) int {
	switch align {
	case "center":
		return (canvasWidth - lineWidth) / 2
	case "right":
		return canvasWidth - lineWidth
	default:
		return 0
	}
}

// drawLine advances color state across lines while drawing glyphs in authored
// order, which keeps token positions stable after wrapping and newline removal.
func (f *BitmapFont) drawLine(
	output *image.RGBA,
	line string,
	lineIndex int,
	x int,
	baseTint color.Color,
	runs map[int]textColorRun,
	state *fontRenderState,
) {
	for _, code := range line {
		state.applyColorRun(baseTint, runs)

		glyph, ok := f.glyph(code)
		if !ok {
			continue
		}

		frame := f.frameForRender(glyph, state)
		bounds := frame.Bounds()
		offset := f.frameOffset(glyph.Frame)
		top := lineIndex*f.LineHeight + offset.Y
		left := x + offset.X
		destination := image.Rect(left, top, left+bounds.Dx(), top+bounds.Dy())
		draw.Draw(
			output,
			destination,
			modulatedImage{Image: frame, Tint: state.tint},
			bounds.Min,
			draw.Over,
		)

		x += glyph.Width
	}
}

// applyColorRun updates render state before the rune at the current visible
// position and always advances the position, preserving token scope across lines.
func (state *fontRenderState) applyColorRun(baseTint color.Color, runs map[int]textColorRun) {
	if next, ok := runs[state.runIndex]; ok {
		if next.reset {
			state.tint = baseTint
			state.transform = -1
		} else {
			state.tint = next.fallback
			state.transform = next.transform
		}
	}

	state.runIndex++
}

// frameForRender prefers an authored PL2 transform for the glyph frame and
// switches tint to white so the transformed palette colors remain unmodified.
func (f *BitmapFont) frameForRender(glyph Glyph, state *fontRenderState) image.Image {
	frame := f.Frames[glyph.Frame]
	if transformed := f.TextFrames[state.transform]; glyph.Frame < len(transformed) {
		frame = transformed[glyph.Frame]
		state.tint = color.White
	}

	return frame
}

// frameOffset returns an authored DC6 offset when available and the zero origin
// for manually assembled or legacy BitmapFont values that omit offset metadata.
func (f *BitmapFont) frameOffset(frame int) image.Point {
	if frame < len(f.FrameOffsets) {
		return f.FrameOffsets[frame]
	}

	return image.Point{}
}

// wrap breaks paragraphs only at glyph boundaries and removes unsupported
// runes, matching rendering's fallback-glyph behavior and established line widths.
func (f *BitmapFont) wrap(text string, maxWidth int) []string {
	paragraphs := strings.Split(text, "\n")

	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if maxWidth <= 0 {
			lines = append(lines, paragraph)
			continue
		}

		var line []rune

		width := 0

		for _, code := range paragraph {
			glyph, ok := f.glyph(code)
			if !ok {
				continue
			}

			if width > 0 && width+glyph.Width > maxWidth {
				lines = append(lines, string(line))
				line = nil
				width = 0
			}

			line = append(line, code)
			width += glyph.Width
		}

		lines = append(lines, string(line))
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// measure returns the logical advance width used for wrapping and alignment,
// including fallback glyphs but excluding unsupported runes.
func (f *BitmapFont) measure(text string) int {
	width := 0

	for _, code := range text {
		if glyph, ok := f.glyph(code); ok {
			width += glyph.Width
		}
	}

	return width
}

// glyph resolves a code point and then the question-mark fallback, ensuring
// layout and rendering make the same decision for missing characters.
func (f *BitmapFont) glyph(code rune) (Glyph, bool) {
	if glyph, ok := f.Glyphs[code]; ok {
		return glyph, true
	}

	glyph, ok := f.Glyphs['?']

	return glyph, ok
}

// modulatedImage preserves palette-authored glyph shading while applying the
// label color. Treating glyphs as masks would flatten highlights and antialiasing.
type modulatedImage struct {
	image.Image
	Tint color.Color
}

// At implements image.Image by multiplying authored color and alpha by Tint,
// retaining the original image bounds and color model through embedding.
func (m modulatedImage) At(x, y int) color.Color {
	red, green, blue, alpha := m.Image.At(x, y).RGBA()
	tintRed, tintGreen, tintBlue, tintAlpha := m.Tint.RGBA()

	return color.RGBA64{
		R: uint16(uint64(red) * uint64(tintRed) / 0xffff),
		G: uint16(uint64(green) * uint64(tintGreen) / 0xffff),
		B: uint16(uint64(blue) * uint64(tintBlue) / 0xffff),
		A: uint16(uint64(alpha) * uint64(tintAlpha) / 0xffff),
	}
}

// maxInt keeps render canvases non-empty without coupling the font package to
// a wider numeric helper or changing the established minimum-height behavior.
func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
