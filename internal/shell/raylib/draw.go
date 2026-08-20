package raylibshell

import (
	"fmt"
	"math"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/dark-magic/internal/shell"
)

type completionLine struct {
	text     string
	selected bool
}

type transcriptLine struct {
	text                       string
	result, error, warning     bool
	dim, notice, heading, code bool
}

type overlayLayout struct {
	offsetY       int32
	panelHeight   int32
	transcriptTop int32
	contentBottom int32
	promptTop     int32
	lineHeight    int32
	fontSize      float32
	opacity       float64
	available     int
}

// Draw advances animation and renders the current modal view. All raylib calls
// stay on the renderer thread; evaluation only publishes through the channel.
func (o *Overlay) Draw(width, height int) {
	o.updateAnimation(time.Now())

	if o.progress <= 0 {
		return
	}

	layout, completion := o.calculateLayout(height)
	o.drawPanel(width, layout)
	o.drawHeader(layout)

	lines := o.visibleTimeline(width, layout.available)
	for index, line := range lines {
		y := layout.transcriptTop + int32(index)*layout.lineHeight
		o.drawText(line.text, 18, y, layout.fontSize, fade(transcriptColor(line), layout.opacity))
	}

	if o.view == viewLua {
		o.drawLuaFooter(layout, completion)
	} else {
		o.drawLogFooter(layout)
	}
}

// calculateLayout derives panel geometry from live settings and reserves room
// for completion rows only in the Lua view.
func (o *Overlay) calculateLayout(height int) (overlayLayout, []completionLine) {
	settings := o.settings.Values()
	panelHeight := int32(max(220, int(float64(height)*settings.ConsoleHeight)))
	panelHeight = min(panelHeight, int32(height))

	positionProgress, opacity := o.presentation()
	opacity *= settings.Opacity
	offsetY := int32(-float64(panelHeight) * (1 - positionProgress))

	transcriptTop := offsetY + 94
	promptTop := offsetY + panelHeight - 58
	contentBottom := promptTop
	fontSize := float32(settings.FontSize)
	lineHeight := int32(math.Ceil(float64(fontSize) * 1.28))
	completion := o.completionLines()

	if o.view == viewLua {
		contentBottom -= int32(len(completion) * 18)
	}

	if o.view == viewLogs {
		contentBottom = offsetY + panelHeight - 34
	}

	available := max(1, int((contentBottom-transcriptTop)/lineHeight)-1)

	return overlayLayout{
		offsetY:       offsetY,
		panelHeight:   panelHeight,
		transcriptTop: transcriptTop,
		contentBottom: contentBottom,
		promptTop:     promptTop,
		lineHeight:    lineHeight,
		fontSize:      fontSize,
		opacity:       opacity,
		available:     available,
	}, completion
}

// drawPanel paints the modal background and border before any text so glyphs
// remain visible over the game scene.
func (o *Overlay) drawPanel(width int, layout overlayLayout) {
	rl.DrawRectangle(
		0,
		layout.offsetY,
		int32(width),
		layout.panelHeight,
		fade(rl.NewColor(8, 7, 6, 238), layout.opacity),
	)

	border := rl.NewRectangle(
		0,
		float32(layout.offsetY),
		float32(width),
		float32(layout.panelHeight),
	)
	rl.DrawRectangleLinesEx(border, 2, fade(rl.NewColor(176, 119, 38, 255), layout.opacity))
}

// drawHeader shows the target and effective mutability before the active view,
// preventing restricted sessions from looking like developer shells.
func (o *Overlay) drawHeader(layout overlayLayout) {
	policy := o.session.Policy()

	mode := "read-only"
	if policy.Mutable {
		mode = "mutable"
	}

	o.drawText(
		"DARK MAGIC CONSOLE",
		16,
		layout.offsetY+12,
		22,
		fade(rl.NewColor(222, 163, 58, 255), layout.opacity),
	)
	o.drawText(
		fmt.Sprintf("target %s | policy %s (%s)", o.session.Target(), policy.Name, mode),
		16,
		layout.offsetY+40,
		16,
		fade(rl.Gray, layout.opacity),
	)
	o.drawTabs(layout.offsetY, layout.opacity)
}

// visibleTimeline clamps the active view's independent scroll offset and returns
// the final bottom-aligned window without mutating cached transcript lines.
func (o *Overlay) visibleTimeline(width, available int) []transcriptLine {
	lines := o.timeline(width)

	offset := &o.luaOffset
	if o.view == viewLogs {
		offset = &o.logOffset
	}

	*offset = min(*offset, max(0, len(lines)-available))
	end := max(0, len(lines)-*offset)

	start := max(0, end-available)
	if start >= end {
		return nil
	}

	return lines[start:end]
}

// transcriptColor applies one stable precedence chain when a line carries
// multiple style flags, matching the historical immediate-mode renderer.
func transcriptColor(line transcriptLine) rl.Color {
	color := rl.LightGray

	switch {
	case line.error:
		color = rl.NewColor(255, 104, 88, 255)
	case line.warning:
		color = rl.NewColor(245, 190, 75, 255)
	case line.notice:
		color = rl.NewColor(224, 183, 92, 255)
	case line.heading:
		color = rl.NewColor(222, 163, 58, 255)
	case line.code:
		color = rl.NewColor(130, 190, 225, 255)
	case line.result:
		color = rl.NewColor(125, 211, 167, 255)
	case line.dim:
		color = rl.Gray
	}

	return color
}

// drawLuaFooter renders completion choices, prompt state, and Lua-specific key
// hints beneath the transcript window.
func (o *Overlay) drawLuaFooter(layout overlayLayout, completion []completionLine) {
	for index, line := range completion {
		color := rl.Gray
		if line.selected {
			color = rl.NewColor(222, 163, 58, 255)
		}

		o.drawText(
			line.text,
			24,
			layout.contentBottom+int32(index*18),
			14,
			fade(color, layout.opacity),
		)
	}

	status := ""
	if o.busy {
		status = " [evaluating]"
	}

	o.drawText(
		"> "+o.inputWithCaret()+status,
		16,
		layout.promptTop,
		layout.fontSize,
		fade(rl.RayWhite, layout.opacity),
	)

	const help = "F1 Lua  F2 Logs  PgUp/PgDn scroll  ` close  Enter run  " +
		"Shift+Enter newline  Tab complete"
	o.drawText(
		help,
		16,
		layout.offsetY+layout.panelHeight-26,
		14,
		fade(rl.Gray, layout.opacity),
	)
}

// drawLogFooter renders the log-navigation key hints at the panel bottom.
func (o *Overlay) drawLogFooter(layout overlayLayout) {
	const help = "F1 Lua  F2 Logs  Up/Down scroll  PgUp/PgDn page  ` or Esc close"

	o.drawText(
		help,
		16,
		layout.offsetY+layout.panelHeight-26,
		14,
		fade(rl.Gray, layout.opacity),
	)
}

// completionLines returns at most five candidates around the current selection,
// retaining candidate order supplied by Session.
func (o *Overlay) completionLines() []completionLine {
	if len(o.candidates) == 0 {
		return nil
	}

	limit := min(5, len(o.candidates))

	start := 0
	if o.candidateAt >= limit {
		start = min(len(o.candidates)-limit, o.candidateAt-limit/2)
	}

	lines := make([]completionLine, 0, limit)
	for index := start; index < start+limit; index++ {
		candidate := o.candidates[index]
		lines = append(lines, completionLine{
			text:     fmt.Sprintf("%s  %s", candidate.Value, candidate.Detail),
			selected: index == o.candidateAt,
		})
	}

	return lines
}

// drawTabs highlights only the active modal view while retaining fixed hit-free labels.
func (o *Overlay) drawTabs(offsetY int32, opacity float64) {
	luaColor, logColor := rl.Gray, rl.Gray
	if o.view == viewLua {
		luaColor = rl.NewColor(222, 163, 58, 255)
	} else {
		logColor = rl.NewColor(222, 163, 58, 255)
	}

	o.drawText("[F1 LUA]", 16, offsetY+68, 17, fade(luaColor, opacity))
	o.drawText("[F2 LOGS]", 116, offsetY+68, 17, fade(logColor, opacity))
}

// timeline rebuilds wrapped display lines only when events, width, limit, or view
// change, then retains the newest configured line tail.
func (o *Overlay) timeline(width int) []transcriptLine {
	fontSize := o.settings.Values().FontSize
	columns := max(12, int(float64(width-36)/(fontSize*0.61)))
	revision := o.session.TimelineRevision()
	limit := o.settings.Values().TranscriptLimit

	unchanged := revision == o.displayRevision &&
		columns == o.displayColumns &&
		limit == o.displayLimit &&
		o.view == o.displayView
	if unchanged {
		return o.displayLines
	}

	events := o.session.TranscriptTimeline()
	if o.view == viewLogs {
		events = o.session.LogTimeline()
	}

	o.displayLines = wrapTranscript(timelineLines(events), columns)
	if len(o.displayLines) > limit {
		o.displayLines = o.displayLines[len(o.displayLines)-limit:]
	}

	o.displayRevision = revision
	o.displayColumns = columns
	o.displayLimit = limit
	o.displayView = o.view

	return o.displayLines
}

// fade scales existing alpha by clamped overlay opacity without changing RGB.
func fade(color rl.Color, opacity float64) rl.Color {
	color.A = uint8(float64(color.A) * max(0, min(1, opacity)))

	return color
}

// drawText uses the embedded font when loaded and falls back to raylib's default
// font for tests or early frames before GPU font initialization.
func (o *Overlay) drawText(text string, x, y int32, size float32, tint rl.Color) {
	if o.fontLoaded {
		rl.DrawTextEx(o.font, text, rl.NewVector2(float32(x), float32(y)), size, 1, tint)

		return
	}

	rl.DrawText(text, x, y, int32(size), tint)
}

// timelineLines expands multi-line events while carrying semantic style flags to
// every physical line and prefixing only the first command line.
func timelineLines(events []shell.TimelineEvent) []transcriptLine {
	lines := make([]transcriptLine, 0, len(events))

	for _, event := range events {
		parts := strings.Split(event.Text, "\n")
		for index, part := range parts {
			part = strings.ReplaceAll(part, "\t", "    ")
			line := transcriptLine{text: part}

			switch event.Kind {
			case "motd":
				line.notice = true
			case "command":
				if index == 0 {
					line.text = "> " + line.text
				}
			case "value":
				line.result = true
				trimmed := strings.TrimSpace(line.text)
				line.heading = strings.HasPrefix(trimmed, "#")
				line.code = strings.HasPrefix(trimmed, "```") || strings.HasPrefix(line.text, "  ")
			case "error", "log-error":
				line.error = true
			case "log-warn":
				line.warning = true
			case "log-debug":
				line.dim = true
			}

			lines = append(lines, line)
		}
	}

	return lines
}

// wrapTranscript splits by rune columns and copies all style flags to each
// wrapped segment so Unicode and semantic coloring survive line wrapping.
func wrapTranscript(lines []transcriptLine, columns int) []transcriptLine {
	var wrapped []transcriptLine

	for _, line := range lines {
		runes := []rune(line.text)
		for len(runes) > columns {
			wrapped = append(wrapped, line.withText(string(runes[:columns])))
			runes = runes[columns:]
		}

		wrapped = append(wrapped, line.withText(string(runes)))
	}

	return wrapped
}

// withText copies presentation flags while replacing only the physical text segment.
func (line transcriptLine) withText(text string) transcriptLine {
	line.text = text

	return line
}
